import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { EventDetailTabs } from '@/components/domain/event-detail-tabs'
import { ChatPanel } from '@/components/domain/chat-panel'

export default async function EventDetailPostChatPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const { data: chat } = await supabase
    .from('chat')
    .select('id')
    .eq('event_id', id)
    .eq('user_id', user.id)
    .maybeSingle()

  let initialMessages: Array<{ role: 'user' | 'assistant'; text: string }> = []
  if (chat?.id) {
    const { data: rows } = await supabase
      .from('chat_message')
      .select('sender_type, message, created_at')
      .eq('chat_id', chat.id)
      .order('created_at', { ascending: true })

    initialMessages =
      rows?.map((row: { sender_type: string; message: string }) => ({
        role: row.sender_type === 'user' ? 'user' : 'assistant',
        text: row.message,
      })) ?? []
  }

  return (
    <AppShell title="Event detail">
      <div className="flex flex-col gap-6 h-full">
        <EventDetailTabs eventId={id} active="post" />

        <div className="grid grid-cols-12 gap-6 flex-1 min-h-0">
          <div className="col-span-8">
            <div className="text-sm text-gray-500">Post content coming soon.</div>
          </div>
          <div className="col-span-4 flex flex-col min-h-0 h-full">
            <ChatPanel eventId={id} initialMessages={initialMessages} />
          </div>
        </div>
      </div>
    </AppShell>
  )
}
