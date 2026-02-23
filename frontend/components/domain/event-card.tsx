import Link from 'next/link'
import { Card } from '@/components/ui/card'
import { StatusPill } from '@/components/ui/status-pill'
import { Button } from '@/components/ui/button'
// import type { EventStatus } from '@/lib/data/events'

type EventRole = 'owner' | 'engineer' | 'photographer'

export function EventCard({
  title,
  subtitle,
  role,
  href,
}: {
  title: string
  subtitle: string
  // status: EventStatus
  role: EventRole
  href: string
}) {
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="text-sm font-medium">{title}</div>
          <div className="text-xs text-gray-500 mt-1">{subtitle}</div>
          <div className="mt-3">
            {role}
          </div>
          {/* <div className="mt-3">
            <StatusPill status={status} />
          </div> */}
        </div>

        <Link href={href}>
          <Button variant="secondary" size="sm">
            View
          </Button>
        </Link>
      </div>
    </Card>
  )
}