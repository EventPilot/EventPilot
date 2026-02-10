import Link from 'next/link'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { StatusPill } from '@/components/ui/status-pill'
import type { Event } from '@/lib/data/events'

export function EventHeader({ event, rightSlot }: { event: Event; rightSlot?: React.ReactNode }) {
  return (
    <Card className="p-6">
      <div className="flex items-start justify-between gap-6">
        <div>
          <div className="text-lg font-semibold">{event.title}</div>
          <div className="text-sm text-gray-500 mt-1">
            {event.start} • {event.timezone} • {event.location}
          </div>
          <div className="mt-3">
            <StatusPill status={event.status} />
          </div>
        </div>

        {rightSlot ?? (
          <div className="flex items-center gap-3">
            <Link href={`/dashboard/events/${event.id}/post`}>
              <Button>Generate draft</Button>
            </Link>
            <Button variant="secondary">Edit event</Button>
          </div>
        )}
      </div>
    </Card>
  )
}
