'use client'

import { useTransition, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { createEvent } from './actions'

// ─── Field ────────────────────────────────────────────────────────────────────

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

function EmailInput({
  placeholder,
  value,
  onChange,
  onRemove,
  showRemove,
}: {
  placeholder: string
  value: string
  onChange: (v: string) => void
  onRemove?: () => void
  showRemove?: boolean
}) {
  return (
    <div className="flex items-center gap-2">
      <input
        type="email"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-10 flex-1 rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none"
      />
      {/* remove button */}
      {showRemove && (
        <button
          type="button"
          onClick={onRemove}
          className="flex h-8 w-8 items-center justify-center rounded-lg text-gray-400 hover:bg-gray-100 hover:text-gray-600"
          aria-label="Remove"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M1 1L13 13M13 1L1 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
          </svg>
        </button>
      )}
    </div>
  )
}

function RoleRow({
  label,
  checked,
  onToggle,
  locked,
}: {
  label: string
  checked: boolean
  onToggle?: () => void
  locked?: boolean
}) {
  return (
    <label
      className={`flex items-center gap-3 rounded-xl border px-4 py-3 text-sm transition-colors ${
        checked ? 'border-indigo-200 bg-indigo-50' : 'border-gray-200 bg-gray-50'
      } ${locked ? 'cursor-default' : 'cursor-pointer hover:border-indigo-200'}`}
    >
      <input
        type="checkbox"
        checked={checked}
        disabled={locked}
        onChange={onToggle}
        className="h-4 w-4 rounded accent-indigo-600"
      />
      <span className={locked ? 'text-gray-700' : ''}>{label}</span>
      {locked && <span className="ml-auto text-xs text-gray-400">required</span>}
    </label>
  )
}

const TONE_OPTIONS = ['Professional', 'Casual', 'Celebratory', 'Formal']

export function EventForm({ ownerEmail }: { ownerEmail: string }) {
  const [isPending, startTransition] = useTransition()
  const router = useRouter()
  // Role state
  const [includePhotographer, setIncludePhotographer] = useState(false)
  const [photographerEmail, setPhotographerEmail] = useState('')
  const [includeEngineer, setIncludeEngineer] = useState(false)
  const [engineerEmails, setEngineerEmails] = useState<string[]>([''])
  const [tone, setTone] = useState('Professional')  // Tone state
  const [error, setError] = useState<string | null>(null) // error message
  // Engineer helpers
  function addEngineer() {
    setEngineerEmails((prev) => [...prev, ''])
  }
  function removeEngineer(i: number) {
    setEngineerEmails((prev) => prev.filter((_, idx) => idx !== i))
  }
  function updateEngineer(i: number, value: string) {
    setEngineerEmails((prev) => prev.map((v, idx) => (idx === i ? value : v)))
  }

  function handleSubmit(formData: FormData) {
    setError(null)
    const allEmails = [
      includePhotographer ? photographerEmail.trim() : null,
      ...(includeEngineer ? engineerEmails.map((e) => e.trim()) : []),
    ].filter(Boolean) as string[]
 
    if (ownerEmail && allEmails.includes(ownerEmail)) {
      setError('The event owner cannot be assigned an additional role')
      return
    }
    if (new Set(allEmails).size !== allEmails.length) {
      setError('The same person cannot be assigned to multiple roles')
      return
    }
    // Serialize role assignments into JSON for the action
    const roles: { email: string; role: string }[] = []

    if (includePhotographer && photographerEmail.trim()) {
      roles.push({ email: photographerEmail.trim(), role: 'Photographer' })
    }
    if (includeEngineer) {
      engineerEmails
        .map((e) => e.trim())
        .filter(Boolean)
        .forEach((email) => roles.push({ email, role: 'Engineer' }))
    }

    formData.set('roles', JSON.stringify(roles))
    formData.set('tone', tone)
    startTransition(async () => {
      await createEvent(formData)
    })
  }

  const activeRoleCount = [
    true, // owner always
    includePhotographer && !!photographerEmail.trim(),
    includeEngineer && engineerEmails.some((e) => e.trim()),
  ].filter(Boolean).length

  return (
    <form action={handleSubmit}>
      <div className="grid grid-cols-12 gap-6">
        {/* Progress bar */}
        <div className="col-span-12">
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <span className="rounded-full border border-gray-200 bg-indigo-50 px-3 py-1 text-xs">1 Basics</span>
              <span className="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-xs text-gray-600">2 Post plan</span>
              <span className="text-xs text-gray-500 ml-2">Fill in the event. Prompts can auto-trigger after it ends.</span>
            </div>
          </Card>
        </div>

        {/* Left: Event basics */}
        <div className="col-span-8">
          <Card className="p-6">
            <div className="text-lg font-semibold">Event basics</div>
            <div className="mt-6 space-y-5">
              <Field label="Event title" placeholder="Customer Rocket Launch — Artemis Demo" name="title" required />
              <div className="grid grid-cols-2 gap-4">
                <Field label="Event date" placeholder="" name="event_date" type="datetime-local" required />
                <Field label="Location" placeholder="Wallops Flight Facility" name="location" />
              </div>
              <Field label="Description" placeholder="What is this milestone, and why it matters? (short)" name="description" tall required />
              <Field label="Tags" placeholder="Launch, Milestone, Customer" name="tags" />
            </div>
          </Card>
        </div>

        {/* Right: Post plan */}
        <div className="col-span-4 space-y-6">
          <Card className="p-6">
            <div className="text-lg font-semibold">Post plan</div>
            <div className="mt-5 space-y-5">
              {/* Roles */}
              <div>
                <div className="text-xs text-gray-500">Roles to prompt</div>
                <div className="mt-2 space-y-2">
                  {/* Owner — always on */}
                  <RoleRow label="Owner" checked locked />
                  {/* Photographer */}
                  <RoleRow
                    label="Photographer"
                    checked={includePhotographer}
                    onToggle={() => {
                      setIncludePhotographer((v) => !v)
                      if (includePhotographer) setPhotographerEmail('')
                    }}
                  />
                  {includePhotographer && (
                    <div className="pl-1">
                      <EmailInput
                        placeholder="photographer@example.com"
                        value={photographerEmail}
                        onChange={setPhotographerEmail}
                      />
                    </div>
                  )}
                  {/* Engineers */}
                  <RoleRow
                    label="Engineer"
                    checked={includeEngineer}
                    onToggle={() => {
                      setIncludeEngineer((v) => !v)
                      if (includeEngineer) setEngineerEmails([''])
                    }}
                  />
                  {includeEngineer && (
                    <div className="space-y-2 pl-1">
                      {engineerEmails.map((email, i) => (
                        <EmailInput
                          key={i}
                          placeholder={`engineer${engineerEmails.length > 1 ? ` ${i + 1}` : ''}@example.com`}
                          value={email}
                          onChange={(v) => updateEngineer(i, v)}
                          onRemove={() => removeEngineer(i)}
                          showRemove={engineerEmails.length > 1}
                        />
                      ))}
                      <button
                        type="button"
                        onClick={addEngineer}
                        className="flex items-center gap-1.5 rounded-lg px-2 py-1 text-xs text-indigo-600 hover:bg-indigo-50"
                      >
                        <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                          <path d="M6 1V11M1 6H11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                        </svg>
                        Add engineer
                      </button>
                    </div>
                  )}
                </div>
              </div>
              {/* Tone */}
              <div>
                <div className="text-xs text-gray-500">Tone</div>
                <select
                  value={tone}
                  onChange={(e) => setTone(e.target.value)}
                  className="mt-2 h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none appearance-none cursor-pointer"
                  style={{ backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath d='M2 4L6 8L10 4' stroke='%239ca3af' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round' fill='none'/%3E%3C/svg%3E")`, backgroundRepeat: 'no-repeat', backgroundPosition: 'right 14px center' }}
                >
                  {TONE_OPTIONS.map((opt) => (
                    <option key={opt} value={opt}>{opt}</option>
                  ))}
                </select>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <div className="text-lg font-semibold">Ready?</div>
            <div className="text-sm text-gray-500 mt-1">
              {activeRoleCount === 1
                ? 'Only you will be prompted.'
                : `${activeRoleCount} roles will be prompted.`}
            </div>

            <div className="mt-5 flex items-center gap-3">
              <Button type="submit" disabled={isPending}>
                {isPending ? 'Creating...' : 'Create event'}
              </Button>
              <Button type="button" variant="secondary" onClick={() => router.back()}>
                Cancel
              </Button>
            </div>
            {error && (
              <div className="mt-3 rounded-lg border border-red-200 bg-red-50 px-4 py-2 text-xs text-red-600">
                {error}
              </div>
            )}
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