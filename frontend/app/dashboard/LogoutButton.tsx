'use client'

import { useRouter } from 'next/navigation'

export default function LogoutButton() {
  const router = useRouter()

  const handleLogout = async () => {
    const res = await fetch('/api/auth/logout', {
      method: 'POST',
    })

    if (res.ok) {
      router.push('/login')
      router.refresh()
    }
  }

  return (
    <button onClick={handleLogout}>
      Logout
    </button>
  )
}