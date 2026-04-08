'use client'

import Link from 'next/link'
import { useEffect, useRef, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/ui/theme-toggle'

export function Topbar({
  title,
  userName = 'Account',
  userSubline,
  showCreateAction = true,
}: {
  title: string
  userName?: string
  userSubline?: string
  showCreateAction?: boolean
}) {
  const router = useRouter()
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    function handlePointerDown(event: MouseEvent) {
      if (!menuRef.current?.contains(event.target as Node)) {
        setMenuOpen(false)
      }
    }

    document.addEventListener('mousedown', handlePointerDown)
    return () => document.removeEventListener('mousedown', handlePointerDown)
  }, [])

  async function handleLogout() {
    const res = await fetch('/api/logout', { method: 'POST' })
    if (res.ok) {
      router.push('/login')
      router.refresh()
    }
  }

  return (
    <div className="h-[72px] border-b border-gray-200 bg-white dark:border-slate-800 dark:bg-slate-950">
      <div className="flex h-full items-center justify-between px-6">
        <div className="flex items-center">
          <Link
            href="/dashboard"
            className="inline-flex items-center rounded-2xl border border-transparent px-1 py-1 transition hover:bg-gray-100 dark:hover:bg-slate-900"
            aria-label="Go to home"
          >
            <span className="text-lg text-gray-900 dark:text-slate-100">
              <span className="font-semibold">Event</span> Pilot
            </span>
          </Link>
        </div>

        <div className="flex items-center gap-3">
          {showCreateAction ? (
            <Link href="/dashboard/events/new">
              <Button>+ Create event</Button>
            </Link>
          ) : null}

          <ThemeToggle />

          <div className="relative" ref={menuRef}>
            <button
              type="button"
              onClick={() => setMenuOpen((open) => !open)}
              className="flex h-9 items-center rounded-full border border-gray-200 bg-gray-50 px-4 text-xs text-gray-600 transition hover:bg-gray-100 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              Me
            </button>

            {menuOpen ? (
              <div className="absolute right-0 top-12 z-20 w-56 rounded-2xl border border-gray-200 bg-white p-2 shadow-lg dark:border-slate-700 dark:bg-slate-900 dark:shadow-[0_18px_50px_rgba(2,6,23,0.55)]">
                <div className="border-b border-gray-100 px-3 py-2 dark:border-slate-800">
                  <div className="text-sm font-medium text-gray-900 dark:text-slate-100">{userName}</div>
                  <div className="mt-1 text-xs text-gray-500 dark:text-slate-400">{userSubline ?? 'Signed in'}</div>
                </div>

                <Link
                  href="/dashboard/settings"
                  className="mt-2 flex rounded-xl px-3 py-2 text-sm text-gray-700 transition hover:bg-gray-50 dark:text-slate-200 dark:hover:bg-slate-800"
                  onClick={() => setMenuOpen(false)}
                >
                  Settings
                </Link>

                <button
                  type="button"
                  onClick={handleLogout}
                  className="mt-1 flex w-full rounded-xl px-3 py-2 text-left text-sm text-red-600 transition hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/40"
                >
                  Logout
                </button>
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </div>
  )
}
