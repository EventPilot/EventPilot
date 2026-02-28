'use client'

import { useTransition } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { deleteEventAction } from '@/app/dashboard/events/[id]/actions'

export function EventActions({ eventId }: { eventId: string}) {
  const [isPending, startTransition] = useTransition()

	function handleDelete() {
    if (!confirm('Are you sure? This will remove all members and delete the event.')) return

    startTransition(async () => {
      await deleteEventAction(eventId)
    })
  }

  return (
    <div className="flex items-center gap-3">
      <Link href={`/dashboard/events/${eventId}/post`}>
        <Button>Generate draft</Button>
      </Link>
      <Link href={`/dashboard/events/${eventId}/edit`}>
        <Button variant="secondary">Edit event</Button>
      </Link>
      <Button variant="danger" onClick={handleDelete} disabled={isPending}>
        {isPending ? 'Deleting...' : 'Delete event'}
      </Button>
    </div>
  )
}