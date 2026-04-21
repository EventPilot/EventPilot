import { createClient } from '@/lib/supabase/server'
import Link from 'next/link'
import { redirect } from 'next/navigation'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/ui/theme-toggle'

export default async function HomePage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (user) {
    redirect('/dashboard')
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="fixed top-4 right-4">
        <ThemeToggle />
      </div>
      <Card className="w-full max-w-3xl p-10">
        <div className="text-2xl font-semibold text-gray-900 dark:text-slate-100">EventPilot</div>
        <div className="text-sm text-gray-500 mt-2 dark:text-slate-400">Calendar-to-social marketing agent for small engineering teams.</div>

        <div className="mt-6 grid grid-cols-3 gap-4">
          <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-slate-700 dark:bg-slate-900">
            <div className="text-sm font-medium text-gray-900 dark:text-slate-100">1) Track milestones</div>
            <div className="text-xs text-gray-500 mt-1 dark:text-slate-400">Calendar events for launches, demos, and customer milestones.</div>
          </div>
          <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-slate-700 dark:bg-slate-900">
            <div className="text-sm font-medium text-gray-900 dark:text-slate-100">2) Collect inputs</div>
            <div className="text-xs text-gray-500 mt-1 dark:text-slate-400">Prompt owner / photographer / customer for media + details.</div>
          </div>
          <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-slate-700 dark:bg-slate-900">
            <div className="text-sm font-medium text-gray-900 dark:text-slate-100">3) Publish</div>
            <div className="text-xs text-gray-500 mt-1 dark:text-slate-400">Generate a draft, refine in the editor, post to X (then LinkedIn/IG).</div>
          </div>
        </div>

        <div className="mt-8 flex items-center gap-3">
          <Link href="/login"><Button>Login</Button></Link>
          <Link href="/signup"><Button variant="secondary">Sign up</Button></Link>
        </div>
      </Card>
    </div>
  )
}
