import { notFound, redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { EventHeader } from '@/components/domain/event-header'
import { EventDetailTabs } from '@/components/domain/event-detail-tabs'
import { XPostCard } from '@/components/domain/x-post-card'
import { ChatPanel } from '@/components/domain/chat-panel'
// import { getEventById } from '@/lib/data/events'

export default async function EventDetailPostChatPage({ params }: { params: { id: string } }) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  // const event = await getEventById(params.id)
  // if (!event) notFound()

  const published = { ...event, status: 'Published' as const }

  return (
    <AppShell title="Event detail">
      <div className="space-y-6">
        {/* <EventHeader
          event={published}
          rightSlot={
            <div className="flex items-center gap-3">
              <Button variant="secondary">Copy link</Button>
              <Button>Republish</Button>
            </div>
          }
        /> */}

        {/* <EventDetailTabs eventId={event.id} active="post" /> */}

        <div className="grid grid-cols-12 gap-6">
          {/* <div className="col-span-8 space-y-6">
            <Card className="p-6">
              <div className="text-lg font-semibold">Published post</div>
              <div className="text-sm text-gray-500 mt-1">Preview (matches X format closely)</div>
              <div className="mt-5">
                <XPostCard content={event.draft.text || 'Post content goes here.'} hashtags={event.draft.hashtags} />
              </div>
            </Card>

            <Card className="p-6 bg-gray-50">
              <div className="text-lg font-semibold">Other platform drafts</div>
              <div className="text-sm text-gray-500 mt-1">Use the chat to generate LinkedIn/Instagram variants.</div>

              <div className="mt-4 grid grid-cols-3 gap-4">
                {[{ name: 'LinkedIn' }, { name: 'Instagram' }, { name: 'Press kit' }].map((p) => (
                  <Card key={p.name} className="p-4">
                    <div className="text-sm font-medium">{p.name}</div>
                    <div className="text-xs text-gray-500 mt-1">Not generated</div>
                    <div className="mt-4">
                      <Button variant="secondary" size="sm">Generate</Button>
                    </div>
                  </Card>
                ))}
              </div>
            </Card>
          </div> */}
          <div>Haven't implemented yet</div>
          <div className="col-span-4">
            <ChatPanel />
          </div>
        </div>
      </div>
    </AppShell>
  )
}
