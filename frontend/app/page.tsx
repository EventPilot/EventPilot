import { createClient } from '@/lib/supabase/server'
import Link from 'next/link'
import { redirect } from 'next/navigation'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

export default async function HomePage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (user) {
    redirect('/dashboard')
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <Card className="w-full max-w-3xl p-10">
        <div className="text-2xl font-semibold">EventPilot</div>
        <div className="text-sm text-gray-500 mt-2">Calendar-to-social marketing agent for small engineering teams.</div>

        <div className="mt-6 grid grid-cols-3 gap-4">
          <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
            <div className="text-sm font-medium">1) Track milestones</div>
            <div className="text-xs text-gray-500 mt-1">Calendar events for launches, demos, and customer milestones.</div>
          </div>
          <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
            <div className="text-sm font-medium">2) Collect inputs</div>
            <div className="text-xs text-gray-500 mt-1">Prompt owner / photographer / customer for media + details.</div>
          </div>
          <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
            <div className="text-sm font-medium">3) Publish</div>
            <div className="text-xs text-gray-500 mt-1">Generate a draft, refine with chat, post to X (then LinkedIn/IG).</div>
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
