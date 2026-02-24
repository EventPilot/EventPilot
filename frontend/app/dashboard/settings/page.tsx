import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { Card } from '@/components/ui/card'
import { LogoutButton } from '@/components/auth/logout-button'

export default async function SettingsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const { data: profile } = await supabase.from('user').select('name').eq('id', user.id).single()

  return (
    <AppShell title="Settings" userName={profile?.name ?? user.email?.split('@')[0] ?? 'Account'} userSubline={user.email ?? ''}>
      <div className="grid grid-cols-12 gap-6">
        <div className="col-span-8 space-y-6">
          <Card className="p-6">
            <div className="text-lg font-semibold">Workspace</div>
            <div className="text-sm text-gray-500 mt-1">Connection + defaults (hard-coded for now).</div>

            <div className="mt-6 space-y-4">
              <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
                <div className="text-sm font-medium">Platform</div>
                <div className="text-xs text-gray-500 mt-1">X enabled; LinkedIn/Instagram coming next.</div>
              </div>

              <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
                <div className="text-sm font-medium">Model preference</div>
                <div className="text-xs text-gray-500 mt-1">Claude by default; OpenAI fallback.</div>
              </div>

              <div className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
                <div className="text-sm font-medium">Supabase</div>
                <div className="text-xs text-gray-500 mt-1">Auth connected. Events/drafts will be wired later.</div>
              </div>
            </div>
          </Card>
        </div>

        <div className="col-span-4 space-y-6">
          <Card className="p-6">
            <div className="text-lg font-semibold">Account</div>
            <div className="text-sm text-gray-500 mt-1">Signed in as</div>
            <div className="mt-4 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-sm">
              <div><span className="text-gray-500">Email:</span> {user.email}</div>
              <div className="mt-2"><span className="text-gray-500">Name:</span> {profile?.name ?? '—'}</div>
            </div>
            <div className="mt-4">
              <LogoutButton />
            </div>
          </Card>
        </div>
      </div>
    </AppShell>
  )
}
