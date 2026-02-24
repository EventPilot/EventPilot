import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'

export default async function InboxPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  return (
    <AppShell title="Inbox">
      <Card className="p-6">
        <div className="text-lg font-semibold">Inbox</div>
        <div className="text-sm text-gray-500 mt-1">Role responses, reminders, and approvals will show up here.</div>
        <div className="mt-6 rounded-2xl border border-gray-200 bg-gray-50 p-6 text-sm text-gray-600">
          Placeholder screen (next step: wire to Supabase tables for prompts + responses).
        </div>
      </Card>
    </AppShell>
  )
}
