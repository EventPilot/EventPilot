import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'

export default async function InboxPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')
  const { data: profile } = await supabase.from('user').select('name').eq('id', user.id).single()

  return (
    <AppShell title="Inbox" userName={profile?.name ?? user.email?.split('@')[0] ?? 'Account'} userSubline={user.email ?? ''}>
      <Card className="p-6">
        <div className="text-lg font-semibold">Inbox</div>
        <div className="mt-1 text-sm text-gray-500 dark:text-slate-400">Role responses, reminders, and approvals will show up here.</div>
        <div className="mt-6 rounded-2xl border border-gray-200 bg-gray-50 p-6 text-sm text-gray-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300">
          Placeholder screen (next step: wire to Supabase tables for prompts + responses).
        </div>
      </Card>
    </AppShell>
  )
}
