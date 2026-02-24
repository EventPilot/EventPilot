import { notFound } from 'next/navigation'
import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { EventHeader } from '@/components/domain/event-header'
import { EventDetailTabs } from '@/components/domain/event-detail-tabs'
import { Timeline } from '@/components/domain/timeline'
import { RoleInputs } from '@/components/domain/role-inputs'
import { DraftPreview } from '@/components/domain/draft-preview'
import { getEventById } from '@/lib/data/events'

export default async function EventDetailCollectPage({ params }: { params: { id: string } }) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const event = await getEventById(params.id)
  if (!event) notFound()

  return (
    <AppShell title="Event detail">
      <div className="space-y-6">
        <EventHeader event={event} />

        <EventDetailTabs eventId={event.id} active="collect" />

        <Card className="bg-amber-50 p-4">
          <div className="flex items-center justify-between gap-4">
            <div className="text-sm">Event finished. Waiting on Photographer + Customer inputs.</div>
            <Button variant="secondary" size="sm">Send reminder</Button>
          </div>
        </Card>

        <div className="grid grid-cols-12 gap-6">
          <div className="col-span-4">
            <Timeline items={event.timeline} />
          </div>
          <div className="col-span-4">
            <RoleInputs inputs={event.roleInputs} />
          </div>
          <div className="col-span-4">
            <DraftPreview draft={event.draft} inputs={event.roleInputs} />
          </div>
        </div>
      </div>
    </AppShell>
  )
}
