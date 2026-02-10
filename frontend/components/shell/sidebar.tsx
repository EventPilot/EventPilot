'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/cn'

const NAV = [
  { name: 'Home', href: '/dashboard' },
  { name: 'Events', href: '/dashboard/events' },
  { name: 'Create', href: '/dashboard/events/new' },
  { name: 'Inbox', href: '/dashboard/inbox' },
  { name: 'Drafts', href: '/dashboard/drafts' },
  { name: 'Settings', href: '/dashboard/settings' },
]

export function Sidebar({ userName = 'Wayne Chiang', subline = 'Owner • Workspace A' }: { userName?: string; subline?: string }) {
  const pathname = usePathname()

  return (
    <aside className="w-[260px] shrink-0 border-r border-gray-200 bg-white flex flex-col">
      <div className="px-6 pt-10">
        <div className="text-lg font-semibold">EventPilot</div>
        <div className="text-xs text-gray-500">Marketing Agent</div>
      </div>

      <nav className="px-4 pt-10">
        <ul className="space-y-3">
          {NAV.map((item) => {
            const active = pathname === item.href || (item.href !== '/dashboard' && pathname.startsWith(item.href))
            return (
              <li key={item.name}>
                <Link
                  href={item.href}
                  className={cn(
                    'flex items-center rounded-xl border border-gray-200 px-4 py-3 text-sm',
                    active ? 'bg-indigo-50' : 'bg-white hover:bg-gray-50'
                  )}
                >
                  {item.name}
                </Link>
              </li>
            )
          })}
        </ul>
      </nav>

      <div className="mt-auto px-4 pb-6 pt-10">
        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <div className="text-sm font-medium">{userName}</div>
          <div className="text-xs text-gray-500">{subline}</div>
        </div>
      </div>
    </aside>
  )
}
