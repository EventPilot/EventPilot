import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { listEvents } from '@/lib/data/events'
import { EventCard } from '@/components/domain/event-card'

export default async function EventsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const events = await listEvents()

  return (
    <AppShell title="Events">
      <Card className="p-6">
        <div className="flex items-center justify-between gap-4">
          <div>
            <div className="text-lg font-semibold">All events</div>
            <div className="text-sm text-gray-500 mt-1">Hard-coded sample events for now.</div>
          </div>
          <Link href="/dashboard/events/new">
            <Button>+ Add event</Button>
          </Link>
        </div>

        <div className="mt-6 grid grid-cols-2 gap-4">
          {events.map((e) => (
            <EventCard
              key={e.id}
              title={e.title}
              subtitle={`${e.start} • ${e.location}`}
              status={e.status}
              href={`/dashboard/events/${e.id}`}
            />
          ))}
        </div>
      </Card>
    </AppShell>
  )
}
