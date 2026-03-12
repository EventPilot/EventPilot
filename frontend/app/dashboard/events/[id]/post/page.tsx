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

  return (
    <AppShell title="Event detail">
      <div className="flex flex-col gap-6 h-full">
        <EventDetailTabs eventId={id} active="post" />

        <div className="grid grid-cols-12 gap-6 flex-1 min-h-0">
          <div className="col-span-8">
            <div className="text-sm text-gray-500">Post content coming soon.</div>
          </div>
          <div className="col-span-4 flex flex-col min-h-0 h-full">
            <ChatPanel eventId={id} />
          </div>
        </div>
      </div>
    </AppShell>
  )
}
