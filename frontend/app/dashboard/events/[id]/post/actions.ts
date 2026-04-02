'use server'

import { createClient } from '@/lib/supabase/server'
import { redirect } from 'next/navigation'

type AgentTask = {
  id: string
  title: string
  kind: string
  status: string
  target_user_name?: string
  target_role?: string
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

  return response.json() as Promise<{ chat_id: string; run: AgentRun | null }>
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
