import Link from 'next/link'
import { cn } from '@/lib/cn'

export function EventDetailTabs({ eventId, active }: { eventId: string; active: 'collect' | 'post' }) {
  const tabs = [
    { key: 'collect', label: 'Collect inputs', href: `/dashboard/events/${eventId}` },
    { key: 'post', label: 'Post & Chat', href: `/dashboard/events/${eventId}/post` },
  ] as const

  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-2">
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
                  ? 'border-gray-200 bg-indigo-50 text-gray-900'
                  : 'border-gray-200 bg-gray-50 text-gray-600 hover:bg-gray-100'
              )}
            >
              {t.label}
            </Link>
          )
        })}
      </div>
    </div>
  )
}
