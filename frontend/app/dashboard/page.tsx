import Link from 'next/link'
import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { EventCard } from '@/components/domain/event-card'
import { TaskCard } from '@/components/domain/task-card'
import { getNextEvent, listEvents } from '@/lib/data/events'

export default async function DashboardHomePage() {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    redirect('/login')
  }

  const { data: profile } = await supabase.from('user').select('name').eq('id', user.id).single()

  const name = profile?.name ?? user.email?.split('@')[0] ?? 'Account'

  const current = await getNextEvent()
  const events = await listEvents()

  return (
    <AppShell title="Home" userName={name} userSubline="Owner • Workspace A">
      <div className="grid grid-cols-12 gap-6">
        <div className="col-span-8 space-y-6">
          <Card className="p-6">
            <div className="flex items-start justify-between gap-6">
              <div>
                <div className="text-lg font-semibold">Current scheduled event</div>
                <div className="text-sm text-gray-500 mt-1">
                  The next milestone that will trigger post generation.
                </div>
              </div>

              <Link href="/dashboard/events/new">
                <Button>+ Add event</Button>
              </Link>
            </div>

            <div className="mt-5">
              {current ? (
                <EventCard
                  title={current.title}
                  subtitle={`${current.start} • ${current.location}`}
                  status={current.status}
                  role='engineer'
                  href={`/dashboard/events/${current.id}`}
                />
              ) : (
                <div className="rounded-2xl border border-gray-200 bg-gray-50 p-6 text-sm text-gray-600">
                  No events yet. Create one to start collecting inputs after it ends.
                </div>
              )}
            </div>
          </Card>

          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-lg font-semibold">Upcoming</div>
                <div className="text-sm text-gray-500 mt-1">All scheduled milestones and drafts.</div>
              </div>
              <Link href="/dashboard/events">
                <Button variant="secondary">View all</Button>
              </Link>
            </div>

            <div className="mt-5 grid grid-cols-2 gap-4">
              {events.slice(0, 4).map((e) => (
                <EventCard
                  key={e.id}
                  title={e.title}
                  subtitle={`${e.start} • ${e.location}`}
                  status={e.status}
                  role='photographer'
                  href={`/dashboard/events/${e.id}`}
                />
              ))}
            </div>
          </Card>
        </div>

        <div className="col-span-4 space-y-6">
          <Card className="p-6">
            <div className="text-lg font-semibold">Action items</div>
            <div className="text-sm text-gray-500 mt-1">Things that block a post from being generated.</div>

            <div className="mt-5 space-y-4">
              <TaskCard title="Photographer upload pending" meta="Customer Rocket Launch — Artemis Demo" />
              <TaskCard title="Customer quote needed" meta="Customer Rocket Launch — Artemis Demo" />
              <TaskCard title="Review draft for Press Kit" meta="Press Kit Review" />
            </div>
          </Card>

          <Card className="p-6">
            <div className="text-lg font-semibold">Quick links</div>
            <div className="mt-4 flex flex-col gap-3">
              <Link href="/dashboard/events/new" className="text-sm text-blue-700 hover:underline">
                Create a new event
              </Link>
              <Link href="/dashboard/drafts" className="text-sm text-blue-700 hover:underline">
                Review drafts
              </Link>
              <Link href="/dashboard/settings" className="text-sm text-blue-700 hover:underline">
                Workspace settings
              </Link>
            </div>
          </Card>
        </div>
      </div>
    </AppShell>
  )
}
