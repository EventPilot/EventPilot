import { notFound, redirect } from 'next/navigation'
import { createClient } from '@/lib/supabase/server'
import { AppShell } from '@/components/shell/app-shell'
import { EditEventForm } from '@/components/domain/edit-form'

export default async function EditEventPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  const { data: event, error } = await supabase
    .from('event')
    .select('*')
    .eq('id', id)
    .single()

  if (error || !event) notFound()

  return (
    <AppShell title="Edit event">
      <EditEventForm event={event} />
    </AppShell>
  )
}