import type { ReactNode } from 'react'
import Link from 'next/link'
import { cn } from '@/lib/cn'

export function EventDetailTabs({
  eventId,
  active,
  action,
}: {
  eventId: string
  active: 'collect' | 'post'
  action?: ReactNode
}) {
  const tabs = [
    { key: 'collect', label: 'Collect inputs', href: `/dashboard/events/${eventId}` },
    { key: 'post', label: 'Editor', href: `/dashboard/events/${eventId}/post` },
  ] as const

  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-2 dark:border-slate-800 dark:bg-slate-950">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          {tabs.map((t) => {
            const isActive = t.key === active
            return (
              <Link
                key={t.key}
                href={t.href}
                className={cn(
                  'rounded-full border px-4 py-2 text-xs',
                  isActive
                    ? 'border-gray-200 bg-indigo-50 text-gray-900 dark:border-slate-700 dark:bg-blue-950/60 dark:text-slate-100'
                    : 'border-gray-200 bg-gray-50 text-gray-600 hover:bg-gray-100 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800'
                )}
              >
                {t.label}
              </Link>
            )
          })}
        </div>

        {action ? <div className="flex items-center">{action}</div> : null}
      </div>
    </div>
  )
}
