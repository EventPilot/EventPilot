// app/dashboard/page.tsx
import { createClient } from '@/lib/supabase/server'
import { redirect } from 'next/navigation'
import LogoutButton from './LogoutButton'

export default async function DashboardPage() {
  const supabase = await createClient()
  
  const { data: { user }, error } = await supabase.auth.getUser()

  if (error || !user) {
    redirect('/login')
  }

  const { data: profile } = await supabase
    .from('user')
    .select('*')
    .eq('id', user.id)
    .single()

  return (
    <div>
      <h1>Dashboard</h1>
      
      <div>
        <p>Welcome, {profile?.name}!</p>
        <LogoutButton />
      </div>

      <hr />

      <h2>Your Profile</h2>
      <ul>
        <li>Email: {user.email}</li>
        <li>Name: {profile?.name}</li>
        <li>User ID: {user.id}</li>
        <li>Account created: {new Date(user.created_at).toLocaleDateString()}</li>
      </ul>
    </div>
  )
}