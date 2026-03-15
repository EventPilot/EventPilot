import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'

export default async function DraftsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')
  const { data: profile } = await supabase.from('user').select('name').eq('id', user.id).single()

  return (
    <AppShell title="Drafts" userName={profile?.name ?? user.email?.split('@')[0] ?? 'Account'} userSubline={user.email ?? ''}>
      <Card className="p-6">
        <div className="text-lg font-semibold">Drafts</div>
        <div className="text-sm text-gray-500 mt-1">All generated drafts waiting for approval.</div>
        <div className="mt-6 rounded-2xl border border-gray-200 bg-gray-50 p-6 text-sm text-gray-600">
          Placeholder screen.
        </div>
      </Card>
    </AppShell>
  )
}
