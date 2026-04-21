import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { EventDetailTabs } from '@/components/domain/event-detail-tabs'
import { EditorLayout } from '@/components/domain/editor-layout'
import { Button } from '@/components/ui/button'
import { publishPostAction, type AgentRun, type ChatMessageEntry, type MediaEntry } from './actions'

type AgentTaskRow = {
  id: string
  position: number
  title: string
  kind: string
  status: string
  target_user_id?: string
  instructions?: string
  result?: string
}

type AgentRunRow = {
  id: string
  chat_id: string
  event_id: string
  status: string
  plan_summary: string
  blocked_on_chat_id?: string
  current_task_index: number
  created_at: string
  updated_at: string
}

type MemberRow = {
  user_id: string
  role?: string
  user?: { name?: string } | null
}

export default async function EventDetailPostChatPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const { data: event } = await supabase
    .from('event')
    .select('id, title, description, event_date, location, status')
    .eq('id', id)
    .maybeSingle()

  const { data: membership } = await supabase
    .from('event_member')
    .select('role')
    .eq('event_id', id)
    .eq('user_id', user.id)
    .maybeSingle()

  const isOwner = membership?.role === 'Owner'

  let initialEntries: ChatMessageEntry[] = []
  if (isOwner) {
    const { data: chat } = await supabase
      .from('chat')
      .select('id')
      .eq('event_id', id)
      .eq('user_id', user.id)
      .maybeSingle()

    if (chat?.id) {
      const { data: rows } = await supabase
        .from('chat_message')
        .select('id, sender_type, message, created_at, message_type, metadata')
        .eq('chat_id', chat.id)
        .order('created_at', { ascending: true })

      initialEntries =
        rows?.map((row: { id: string; sender_type: string; message: string; created_at: string; message_type?: string; metadata?: Record<string, unknown> }) => ({
          id: row.id,
          role: row.sender_type === 'user' ? 'user' : 'assistant',
          text: row.message,
          createdAt: row.created_at,
          messageType: row.message_type ?? 'message',
          metadata: row.metadata ?? {},
        })) ?? []
    }
  }

  const { data: latestRunRows } = await supabase
    .from('agent_run')
    .select('id, chat_id, event_id, status, plan_summary, blocked_on_chat_id, current_task_index, created_at, updated_at')
    .eq('event_id', id)
    .eq('requested_by_user_id', user.id)
    .order('updated_at', { ascending: false })
    .limit(1)

  const latestRun = latestRunRows?.[0] as AgentRunRow | undefined

  let initialRun: AgentRun | null = null
  if (latestRun) {
    const { data: taskRows } = await supabase
      .from('agent_task')
      .select('id, position, title, kind, status, target_user_id, instructions, result')
      .eq('run_id', latestRun.id)
      .order('position', { ascending: true })

    const { data: memberRows } = await supabase
      .from('event_member')
      .select('user_id, role, user(name)')
      .eq('event_id', id)

    const memberLookup = new Map<string, MemberRow>()
    for (const member of (memberRows ?? []) as MemberRow[]) {
      memberLookup.set(member.user_id, member)
    }

    initialRun = {
      ...latestRun,
      tasks: ((taskRows ?? []) as AgentTaskRow[]).map((task) => ({
        ...task,
        target_user_name: task.target_user_id ? memberLookup.get(task.target_user_id)?.user?.name : undefined,
        target_role: task.target_user_id ? memberLookup.get(task.target_user_id)?.role : undefined,
      })),
    }
  }

  const { data: post } = await supabase
    .from('post')
    .select('content, status, url, created_at')
    .eq('event_id', id)
    .maybeSingle()

  // Load saved media for this event
  const { data: mediaRows } = await supabase
    .from('media')
    .select('id, event_id, storage_path, created_at')
    .eq('event_id', id)
    .order('created_at', { ascending: true })

  const signedUrls = await Promise.all(
    (mediaRows ?? []).map((row) =>
      supabase.storage.from('event-media').createSignedUrl(row.storage_path, 60 * 60 * 24 * 7)
    )
  )
  const initialMedia: MediaEntry[] = (mediaRows ?? []).map((row, i) => ({
    ...row,
    url: signedUrls[i].data?.signedUrl ?? '',
  }))

  const publishAction = publishPostAction.bind(null, id)
  const canPublish = Boolean(post?.content) && post?.status !== 'published'

  return (
    <AppShell title="Editor">
      <div className="flex h-full min-h-0 flex-col gap-6">
        <EventDetailTabs
          eventId={id}
          active="post"
          action={
            <form action={publishAction}>
              <Button type="submit" disabled={!canPublish}>
                {post?.status === 'published' ? 'Posted' : post?.content ? 'Post' : 'No draft yet'}
              </Button>
            </form>
          }
        />

        <EditorLayout
          eventId={id}
          initialEntries={initialEntries}
          initialRun={initialRun}
          postContent={post?.content ?? ''}
          initialMedia={initialMedia}
        />
      </div>
    </AppShell>
  )
}
