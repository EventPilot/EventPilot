'use client'

import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'

export function LogoutButton() {
  const router = useRouter()

  const handleLogout = async () => {
    const res = await fetch('/api/auth/logout', { method: 'POST' })
    if (res.ok) {
      router.push('/login')
      router.refresh()
    }
  }

  return (
    <Button variant="secondary" onClick={handleLogout}>
      Logout
    </Button>
  )
}
