'use client'

import { useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { createEvent } from './actions'

function Field({
  label,
  placeholder,
  name,
  tall,
  type = 'text',
  required,
}: {
  label: string
  placeholder: string
  name?: string
  tall?: boolean
  type?: string
  required?: boolean
}) {
  return (
    <div>
      <div className="text-xs text-gray-500">{label}</div>
      {tall ? (
        <textarea
          name={name}
          placeholder={placeholder}
          required={required}
          className="mt-2 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm outline-none min-h-[90px]"
        />
      ) : (
        <input
          name={name}
          type={type}
          placeholder={placeholder}
          required={required}
          className="mt-2 h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none"
        />
      )}
    </div>
  )
}

export function EventForm() {
  const [isPending, startTransition] = useTransition()
  const router = useRouter()

  function handleSubmit(formData: FormData) {
    startTransition(async () => {
      await createEvent(formData)
    })
  }

  return (
    <form action={handleSubmit}>
      <div className="grid grid-cols-12 gap-6">
        <div className="col-span-12">
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <span className="rounded-full border border-gray-200 bg-indigo-50 px-3 py-1 text-xs">1 Basics</span>
              <span className="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-xs text-gray-600">2 Post plan</span>
              <span className="text-xs text-gray-500 ml-2">Fill in the event. Prompts can auto-trigger after it ends.</span>
            </div>
          </Card>
        </div>

        <div className="col-span-8">
          <Card className="p-6">
            <div className="text-lg font-semibold">Event basics</div>

            <div className="mt-6 space-y-5">
              <Field label="Event title" placeholder="Customer Rocket Launch — Artemis Demo" name="title" required />
              <Field label="Client / partner" placeholder="Artemis Aerospace" />
              <div className="grid grid-cols-2 gap-4">
                <Field label="Event date" placeholder="" name="event_date" type="date" required />
                <Field label="End date & time" placeholder="02/18/2026 4:00 PM" />
              </div>
              <Field label="Location" placeholder="Wallops Flight Facility" />
              <Field label="Description" placeholder="What is this milestone, and why it matters? (short)" name="description" tall required />
              <Field label="Tags" placeholder="Launch, Milestone, Customer" />
            </div>
          </Card>
        </div>

        <div className="col-span-4 space-y-6">
          <Card className="p-6">
            <div className="text-lg font-semibold">Post plan</div>

            <div className="mt-5 space-y-4">
              <div>
                <div className="text-xs text-gray-500">Platforms</div>
                <div className="mt-2 rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm">X (enabled)</div>
              </div>

              <div>
                <div className="text-xs text-gray-500">Roles to prompt</div>
                <div className="mt-2 space-y-2">
                  {['Owner (required)', 'Photographer', 'Customer/Partner'].map((r) => (
                    <div key={r} className="rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm">☑ {r}</div>
                  ))}
                </div>
              </div>

              <div>
                <div className="text-xs text-gray-500">Tone</div>
                <div className="mt-2 rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm">Professional update ▼</div>
              </div>

              <div>
                <div className="text-xs text-gray-500">Prompt timing</div>
                <div className="mt-2 rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm">
                  <div>Send immediately after event ends</div>
                  <div className="text-xs text-gray-500 mt-1">Remind after 6h if incomplete</div>
                </div>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <div className="text-lg font-semibold">Ready?</div>
            <div className="text-sm text-gray-500 mt-1">Create the event and start collecting inputs.</div>

            <div className="mt-5 flex items-center gap-3">
              <Button type="submit" disabled={isPending}>
                {isPending ? 'Creating...' : 'Create event'}
              </Button>
              <Button type="button" variant="secondary" onClick={() => router.back()}>
                Cancel
              </Button>
            </div>

            <div className="mt-5 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-xs text-gray-600">
              <div className="font-medium text-gray-900">What happens next</div>
              <div className="mt-2">1) Event appears on Home</div>
              <div>2) Prompts auto-send after end time</div>
              <div>3) Draft generated for review</div>
            </div>
          </Card>
        </div>
      </div>
    </form>
  )
}