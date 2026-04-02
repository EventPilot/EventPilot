'use server'

import { createClient } from '@/lib/supabase/server'
import { revalidatePath } from 'next/cache'
import { redirect } from 'next/navigation'

type AgentTask = {
  id: string
  position: number
  title: string
  kind: string
  status: string
  target_user_id?: string
  target_user_name?: string
  target_role?: string
  instructions?: string
  result?: string
}

export type AgentRun = {
  id: string
  chat_id: string
  event_id: string
  status: string
  plan_summary: string
  tasks: AgentTask[]
  blocked_on_chat_id?: string
  current_task_index: number
  created_at: string
  updated_at: string
}

export type ChatMessageEntry = {
  id: string
  role: 'user' | 'assistant'
  text: string
  createdAt: string
  messageType: string
  metadata?: Record<string, unknown>
}

export type SendChatMessageResult = {
  chat_id: string
  run: AgentRun | null
  queued: boolean
  message: {
    id: string
    message: string
    message_type: string
    created_at: string
    metadata?: Record<string, unknown>
  }
}

function apiBaseUrl() {
  return (
    process.env.EVENTPILOT_API_BASE_URL ??
    process.env.NEXT_PUBLIC_API_BASE_URL ??
    'http://localhost:8080'
  )
}

export async function sendChatMessageAction(eventId: string, message: string) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (!session?.access_token) {
    throw new Error('Missing auth session')
  }

  const response = await fetch(`${apiBaseUrl()}/api/events/${eventId}/chat/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${session.access_token}`,
    },
    body: JSON.stringify({ message }),
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error(await response.text())
  }

  return response.json() as Promise<SendChatMessageResult>
}

export async function approveAgentRunAction(runId: string) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (!session?.access_token) {
    throw new Error('Missing auth session')
  }

  const response = await fetch(`${apiBaseUrl()}/api/agent-runs/${runId}/approve`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${session.access_token}`,
    },
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error(await response.text())
  }

  return response.json() as Promise<AgentRun>
}

export async function getLatestAgentRunAction(eventId: string) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const { data: runs, error: runError } = await supabase
    .from('agent_run')
    .select('id, chat_id, event_id, status, plan_summary, blocked_on_chat_id, current_task_index, created_at, updated_at')
    .eq('event_id', eventId)
    .eq('requested_by_user_id', user.id)
    .order('updated_at', { ascending: false })
    .limit(1)

  if (runError) {
    throw new Error(runError.message)
  }

  const run = runs?.[0]
  if (!run) {
    return null
  }

  const { data: tasks, error: taskError } = await supabase
    .from('agent_task')
    .select('id, position, title, kind, status, target_user_id, instructions, result')
    .eq('run_id', run.id)
    .order('position', { ascending: true })

  if (taskError) {
    throw new Error(taskError.message)
  }

  const targetIds = Array.from(
    new Set((tasks ?? []).map((task) => task.target_user_id).filter((value): value is string => Boolean(value)))
  )

  const memberLookup = new Map<string, { name?: string; role?: string }>()
  if (targetIds.length > 0) {
    const { data: members, error: memberError } = await supabase
      .from('event_member')
      .select('user_id, role, user(name)')
      .eq('event_id', eventId)
      .in('user_id', targetIds)

    if (memberError) {
      throw new Error(memberError.message)
    }

    for (const member of members ?? []) {
      memberLookup.set(member.user_id, {
        name: typeof member.user === 'object' && member.user && 'name' in member.user ? String(member.user.name ?? '') : undefined,
        role: member.role ?? undefined,
      })
    }
  }

  return {
    ...run,
    tasks: (tasks ?? []).map((task) => ({
      ...task,
      target_user_name: task.target_user_id ? memberLookup.get(task.target_user_id)?.name : undefined,
      target_role: task.target_user_id ? memberLookup.get(task.target_user_id)?.role : undefined,
    })),
  } satisfies AgentRun
}

export async function publishPostAction(eventId: string) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (!session?.access_token) {
    throw new Error('Missing auth session')
  }

  const response = await fetch(`${apiBaseUrl()}/api/events/${eventId}/post/publish`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${session.access_token}`,
    },
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error(await response.text())
  }

  revalidatePath(`/dashboard/events/${eventId}/post`)
}
