import { redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { EventForm } from './event-form'

export default async function NewEventPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')
  const { data: profile } = await supabase.from('user').select('name').eq('id', user.id).single()

  return (
    <AppShell
      title="Create event"
      userName={profile?.name ?? user.email?.split('@')[0] ?? 'Account'}
      userSubline={user.email ?? ''}
      showCreateAction={false}
    >
      <EventForm ownerEmail={user?.email ?? ''} />
    </AppShell>
  )
}
