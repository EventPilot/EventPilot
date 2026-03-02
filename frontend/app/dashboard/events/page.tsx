import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import Link from 'next/link'
import { EventCard } from '@/components/domain/event-card'

export default async function EventsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  // Get events where the user is a member
  const { data: memberRows } = await supabase
    .from('event_member')
    .select(`
      role,
      event (
        id,
        title,
        description,
        event_date,
        created_at,
        location,
        status
      )
    `)
    .eq('user_id', user.id)
    .order('created_at', { referencedTable: 'event', ascending: false })

  const events = memberRows?.map((row: any) => ({
    ...row.event,
    role: row.role,
  })) ?? []

  console.log(events)

  return (
    <AppShell title="Events">
      <Card className="p-6">
        <div className="flex items-center justify-between gap-4">
          <div>
            <div className="text-lg font-semibold">All events</div>
            <div className="text-sm text-gray-500 mt-1">
              {events.length} event{events.length !== 1 ? 's' : ''}
            </div>
          </div>
          <Link href="/dashboard/events/new">
            <Button>+ Add event</Button>
          </Link>
        </div>

        <div className="mt-6 grid grid-cols-3 gap-4">
          {events.map((e: any) => (
            <EventCard
              key={e.id}
              title={e.title}
              subtitle={e.location ? `${e.event_date} • ${e.location}`: `${e.event_date}`}
              status={e.status}
              role={e.role}
              href={`/dashboard/events/${e.id}`}
            />
          ))}
        </div>
      </Card>
    </AppShell>)
}
