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

function inferImageMimeType(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() ?? ''
  switch (ext) {
    case 'jpg':
    case 'jpeg':
      return 'image/jpeg'
    case 'png':
      return 'image/png'
    case 'gif':
      return 'image/gif'
    case 'webp':
      return 'image/webp'
    default:
      return 'application/octet-stream'
  }
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

export type MediaEntry = {
  id: string
  event_id: string
  storage_path: string
  url: string
  created_at: string
}

export async function uploadMediaAction(eventId: string, formData: FormData) {
  const supabase = await createClient()
  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (!session?.access_token) {
    redirect('/login')
  }

  const file = formData.get('file') as File | null
  if (!file) throw new Error('No file provided')

  const backendForm = new FormData()
  backendForm.append('file', file, file.name)

  const response = await fetch(`${apiBaseUrl()}/api/events/${eventId}/media`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${session.access_token}`,
    },
    body: backendForm,
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error(await response.text())
  }

  return (await response.json()) as MediaEntry
}

export async function deleteMediaAction(mediaId: string) {
  const supabase = await createClient()
  const {
    data: { session },
  } = await supabase.auth.getSession()
  if (!session?.access_token) {
    redirect('/login')
  }

  const response = await fetch(`${apiBaseUrl()}/api/media/${mediaId}`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${session.access_token}`,
    },
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error(await response.text())
  }
}

export async function publishPostAction(eventId: string) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const handle = process.env.BSKY_HANDLE
  const password = process.env.BSKY_PASSWORD
  if (!handle || !password) {
    throw new Error('BSKY_HANDLE and BSKY_PASSWORD must be set in the environment')
  }

  const { data: post, error: postError } = await supabase
    .from('post')
    .select('content')
    .eq('event_id', eventId)
    .maybeSingle()
  if (postError) throw new Error(postError.message)
  if (!post?.content) throw new Error('No draft post to publish — generate one first')

  const sessionRes = await fetch('https://bsky.social/xrpc/com.atproto.server.createSession', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identifier: handle, password }),
    cache: 'no-store',
  })
  if (!sessionRes.ok) {
    throw new Error(`Bluesky auth failed (${sessionRes.status}): ${await sessionRes.text()}`)
  }
  const bskySession = (await sessionRes.json()) as { accessJwt: string; did: string; handle: string }

  const { data: mediaRows, error: mediaError } = await supabase
    .from('media')
    .select('id, storage_path')
    .eq('event_id', eventId)
    .order('created_at', { ascending: true })
    .limit(4)
  if (mediaError) throw new Error(mediaError.message)

  const images: Array<{ alt: string; image: unknown }> = []
  for (const row of mediaRows ?? []) {
    const { data: blob, error: dlError } = await supabase.storage
      .from('event-media')
      .download(row.storage_path)
    if (dlError || !blob) {
      throw new Error(`Failed to download media ${row.id}: ${dlError?.message ?? 'no data'}`)
    }

    const mimeType = blob.type || inferImageMimeType(row.storage_path)
    const buffer = await blob.arrayBuffer()

    const uploadRes = await fetch('https://bsky.social/xrpc/com.atproto.repo.uploadBlob', {
      method: 'POST',
      headers: {
        'Content-Type': mimeType,
        Authorization: `Bearer ${bskySession.accessJwt}`,
      },
      body: buffer,
      cache: 'no-store',
    })
    if (!uploadRes.ok) {
      throw new Error(`Bluesky uploadBlob failed (${uploadRes.status}): ${await uploadRes.text()}`)
    }
    const { blob: blobRef } = (await uploadRes.json()) as { blob: unknown }
    images.push({ alt: '', image: blobRef })
  }

  const postRecord: Record<string, unknown> = {
    $type: 'app.bsky.feed.post',
    text: post.content,
    createdAt: new Date().toISOString(),
  }
  if (images.length > 0) {
    postRecord.embed = {
      $type: 'app.bsky.embed.images',
      images,
    }
  }

  const recordRes = await fetch('https://bsky.social/xrpc/com.atproto.repo.createRecord', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${bskySession.accessJwt}`,
    },
    body: JSON.stringify({
      repo: bskySession.did,
      collection: 'app.bsky.feed.post',
      record: postRecord,
    }),
    cache: 'no-store',
  })
  if (!recordRes.ok) {
    throw new Error(`Bluesky post failed (${recordRes.status}): ${await recordRes.text()}`)
  }
  const record = (await recordRes.json()) as { uri: string; cid: string }

  const rkey = record.uri.split('/').pop() ?? ''
  const postUrl = `https://bsky.app/profile/${bskySession.handle}/post/${rkey}`

  const { error: updateError } = await supabase
    .from('post')
    .update({ status: 'published', url: postUrl })
    .eq('event_id', eventId)
  if (updateError) {
    console.error('[publishPostAction] posted to Bluesky but DB update failed:', updateError)
  }

  revalidatePath(`/dashboard/events/${eventId}/post`)
}
