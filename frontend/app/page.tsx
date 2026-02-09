import { createClient } from '@/lib/supabase/server'
import Link from 'next/link'
import { redirect } from 'next/navigation'

export default async function HomePage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (user) {
    redirect('/dashboard')
  }

  return (
    <div>
      <h1>Welcome to EventPilot</h1>
      <p>Calendar and marketing agent</p>
      <p>
        <Link href="/login">Login</Link> | <Link href="/signup">Sign Up</Link>
      </p>
    </div>
  )
}