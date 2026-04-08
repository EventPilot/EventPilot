'use client'

import { useTransition, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { createEvent } from './actions'

type CustomRole = {
  id: string
  name: string
  emails: string[]
}

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
      <div className="text-xs text-gray-500 dark:text-slate-400">{label}</div>
      {tall ? (
        <textarea
          name={name}
          placeholder={placeholder}
          required={required}
          className="mt-2 min-h-[90px] w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm outline-none dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:placeholder:text-slate-500"
        />
      ) : (
        <input
          name={name}
          type={type}
          placeholder={placeholder}
          required={required}
          className="mt-2 h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:placeholder:text-slate-500"
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
        className="h-10 flex-1 rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:placeholder:text-slate-500"
      />
      {/* remove button */}
      {showRemove && (
        <button
          type="button"
          onClick={onRemove}
          className="flex h-8 w-8 items-center justify-center rounded-lg text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:text-slate-500 dark:hover:bg-slate-800 dark:hover:text-slate-300"
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
        checked
          ? 'border-indigo-200 bg-indigo-50 dark:border-blue-800 dark:bg-blue-950/50'
          : 'border-gray-200 bg-gray-50 dark:border-slate-700 dark:bg-slate-900'
      } ${locked ? 'cursor-default' : 'cursor-pointer hover:border-indigo-200 dark:hover:border-blue-700'}`}
    >
      <input
        type="checkbox"
        checked={checked}
        disabled={locked}
        onChange={onToggle}
        className="h-4 w-4 rounded accent-indigo-600"
      />
      <span className={locked ? 'text-gray-700 dark:text-slate-200' : ''}>{label}</span>
      {locked && <span className="ml-auto text-xs text-gray-400 dark:text-slate-500">required</span>}
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
  const [customRoles, setCustomRoles] = useState<CustomRole[]>([])
  const [tone, setTone] = useState('Professional')
  const [error, setError] = useState<string | null>(null)

  // Engineer helpers
  function addEngineer() { setEngineerEmails((prev) => [...prev, '']) }
  function removeEngineer(i: number) { setEngineerEmails((prev) => prev.filter((_, idx) => idx !== i)) }
  function updateEngineer(i: number, value: string) { setEngineerEmails((prev) => prev.map((v, idx) => idx === i ? value : v)) }

  // Custom role helpers
  function addCustomRole() {
    setCustomRoles((prev) => [...prev, { id: crypto.randomUUID(), name: '', emails: [''] }])
  }
  function removeCustomRole(id: string) {
    setCustomRoles((prev) => prev.filter((r) => r.id !== id))
  }
  function updateCustomRoleName(id: string, name: string) {
    setCustomRoles((prev) => prev.map((r) => r.id === id ? { ...r, name } : r))
  }
  function addCustomRoleEmail(id: string) {
    setCustomRoles((prev) => prev.map((r) => r.id === id ? { ...r, emails: [...r.emails, ''] } : r))
  }
  function removeCustomRoleEmail(id: string, i: number) {
    setCustomRoles((prev) => prev.map((r) => r.id === id ? { ...r, emails: r.emails.filter((_, idx) => idx !== i) } : r))
  }
  function updateCustomRoleEmail(id: string, i: number, value: string) {
    setCustomRoles((prev) => prev.map((r) => r.id === id ? { ...r, emails: r.emails.map((e, idx) => idx === i ? value : e) } : r))
  }

  function handleSubmit(formData: FormData) {
    setError(null)
    const allEmails = [
      includePhotographer ? photographerEmail.trim() : null,
      ...(includeEngineer ? engineerEmails.map((e) => e.trim()) : []),
      ...customRoles.flatMap((r) => r.emails.map((e) => e.trim())),
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
      engineerEmails.map((e) => e.trim()).filter(Boolean).forEach((email) => roles.push({ email, role: 'Engineer' }))
    }
    for (const cr of customRoles) {
      const roleName = cr.name.trim() || 'Custom'
      cr.emails.map((e) => e.trim()).filter(Boolean).forEach((email) => roles.push({ email, role: roleName }))
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
    ...customRoles.map((r) => r.emails.some((e) => e.trim())),
  ].filter(Boolean).length

  return (
    <form action={handleSubmit}>
      <div className="grid grid-cols-12 gap-6">
        {/* Progress bar */}
        <div className="col-span-12">
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <span className="rounded-full border border-gray-200 bg-indigo-50 px-3 py-1 text-xs dark:border-slate-700 dark:bg-blue-950/60 dark:text-slate-100">1 Basics</span>
              <span className="rounded-full border border-gray-200 bg-gray-50 px-3 py-1 text-xs text-gray-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300">2 Post plan</span>
              <span className="ml-2 text-xs text-gray-500 dark:text-slate-400">Fill in the event. Prompts can auto-trigger after it ends.</span>
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
                <div className="text-xs text-gray-500 dark:text-slate-400">Roles to prompt</div>
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
                        className="flex items-center gap-1.5 rounded-lg px-2 py-1 text-xs text-indigo-600 hover:bg-indigo-50 dark:text-blue-300 dark:hover:bg-blue-950/50"
                      >
                        <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                          <path d="M6 1V11M1 6H11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                        </svg>
                        Add engineer
                      </button>
                    </div>
                  )}
                  {/* Custom roles */}
                  {customRoles.map((cr) => (
                    <div key={cr.id} className="rounded-xl border border-gray-200 bg-gray-50 p-3 dark:border-slate-700 dark:bg-slate-900">
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          placeholder="Role name (e.g. Engineer, MC, Videographer)"
                          value={cr.name}
                          onChange={(e) => updateCustomRoleName(cr.id, e.target.value)}
                          className="h-9 flex-1 rounded-lg border border-gray-200 bg-white px-3 text-sm outline-none dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100 dark:placeholder:text-slate-500"
                        />
                        <button
                          type="button"
                          onClick={() => removeCustomRole(cr.id)}
                          className="flex h-8 w-8 items-center justify-center rounded-lg text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:text-slate-500 dark:hover:bg-slate-800 dark:hover:text-slate-300"
                          aria-label="Remove role"
                        >
                          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                            <path d="M1 1L13 13M13 1L1 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                          </svg>
                        </button>
                      </div>
                      <div className="mt-2 space-y-2 pl-1">
                        {cr.emails.map((email, i) => (
                          <EmailInput
                            key={i}
                            placeholder={`${cr.name || 'member'}${cr.emails.length > 1 ? ` ${i + 1}` : ''}@example.com`}
                            value={email}
                            onChange={(v) => updateCustomRoleEmail(cr.id, i, v)}
                            onRemove={() => removeCustomRoleEmail(cr.id, i)}
                            showRemove={cr.emails.length > 1}
                          />
                        ))}
                        <button
                          type="button"
                          onClick={() => addCustomRoleEmail(cr.id)}
                          className="flex items-center gap-1.5 rounded-lg px-2 py-1 text-xs text-indigo-600 hover:bg-indigo-50 dark:text-blue-300 dark:hover:bg-blue-950/50"
                        >
                          <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                            <path d="M6 1V11M1 6H11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                          </svg>
                          Add person
                        </button>
                      </div>
                    </div>
                  ))}
                  {/* Add custom role button */}
                  <button
                    type="button"
                    onClick={addCustomRole}
                    className="flex w-full items-center justify-center gap-1.5 rounded-xl border border-dashed border-gray-300 py-2 text-xs text-gray-500 hover:border-indigo-300 hover:text-indigo-600 dark:border-slate-600 dark:text-slate-400 dark:hover:border-blue-600 dark:hover:text-blue-300"
                  >
                    <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                      <path d="M6 1V11M1 6H11" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                    </svg>
                    Add role
                  </button>
                </div>
              </div>
              {/* Tone */}
              <div>
                <div className="text-xs text-gray-500 dark:text-slate-400">Tone</div>
                <select
                  value={tone}
                  onChange={(e) => setTone(e.target.value)}
                  className="mt-2 h-11 w-full cursor-pointer appearance-none rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
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
            <div className="mt-1 text-sm text-gray-500 dark:text-slate-400">
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
              <div className="mt-3 rounded-lg border border-red-200 bg-red-50 px-4 py-2 text-xs text-red-600 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300">
                {error}
              </div>
            )}
            <div className="mt-5 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-xs text-gray-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300">
              <div className="font-medium text-gray-900 dark:text-slate-100">What happens next</div>
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
