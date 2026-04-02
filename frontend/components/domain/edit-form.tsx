'use client'

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { updateEventAction } from '@/app/dashboard/events/[id]/actions'
import { toLocalDateTimeValue } from '../helpers'

interface EventData {
  id: string
  title: string
  description: string
  event_date: string
  location: string
}

function Field({ label, name, defaultValue, tall, type = 'text', required }: {
  label: string; name: string; defaultValue?: string; tall?: boolean; type?: string; required?: boolean
}) {
  return (
    <div>
      <div className="text-xs text-gray-500 dark:text-slate-400">{label}</div>
      {tall ? (
        <textarea
          name={name}
          defaultValue={defaultValue}
          required={required}
          className="mt-2 min-h-[90px] w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm outline-none dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
        />
      ) : (
        <input
          name={name}
          type={type}
          defaultValue={defaultValue}
          required={required}
          className="mt-2 h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
        />
      )}
    </div>
  )
}

export function EditEventForm({ event }: { event: EventData }) {
  const [isPending, startTransition] = useTransition()
  const router = useRouter()

  function handleSubmit(formData: FormData) {
    const localDate = formData.get('event_date') as string
    const isoDate = new Date(localDate).toISOString()
    formData.set('event_date', isoDate)
    
    startTransition(async () => {
      await updateEventAction(event.id, formData)
    })
  }

  return (
    <form action={handleSubmit}>
      <Card className="p-6 max-w-2xl">
        <div className="text-lg font-semibold">Edit event</div>

        <div className="mt-6 space-y-5">
          <Field label="Event title" name="title" defaultValue={event.title} required />
          <Field label="Event date" name="event_date" type="datetime-local" defaultValue={toLocalDateTimeValue(event.event_date)} required />
          <Field label="Location" name="location" defaultValue={event.location} />
          <Field label="Description" name="description" defaultValue={event.description} tall required />
        </div>

        <div className="mt-6 flex items-center gap-3">
          <Button type="submit" disabled={isPending}>
            {isPending ? 'Saving...' : 'Save changes'}
          </Button>
          <Button type="button" variant="secondary" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </Card>
    </form>
  )
}
