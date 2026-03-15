'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { createClient } from '@/lib/supabase/client'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

export default function SignupPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const router = useRouter()
  const supabase = createClient()

  useEffect(() => {
    const checkUser = async () => {
      const { data: { user } } = await supabase.auth.getUser()
      if (user) router.push('/dashboard')
    }
    checkUser()
  }, [router, supabase])

  const handleSignup = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await fetch('/api/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, name }),
      })

      const data = await res.json()

      if (!res.ok) {
        setError(data.error || 'Signup failed')
        return
      }

      router.push('/dashboard')
      router.refresh()
    } catch {
      setError('An unexpected error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <Card className="w-full max-w-md p-8">
        <div className="text-2xl font-semibold">Create account</div>
        <div className="text-sm text-gray-500 mt-2">
          Already have an account?{' '}
          <Link href="/login" className="text-blue-700 hover:underline">
            Sign in
          </Link>
        </div>

        <form onSubmit={handleSignup} className="mt-6 space-y-4">
          {error && <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</div>}

          <div>
            <div className="text-xs text-gray-500">Name</div>
            <input
              className="mt-2 h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your name"
            />
          </div>

          <div>
            <div className="text-xs text-gray-500">Email</div>
            <input
              className="mt-2 h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
            />
          </div>

          <div>
            <div className="text-xs text-gray-500">Password</div>
            <input
              className="mt-2 h-11 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 text-sm outline-none"
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Minimum 6 characters"
            />
          </div>

          <Button type="submit" disabled={loading} className="w-full">
            {loading ? 'Creating account…' : 'Sign up'}
          </Button>

          <Link href="/" className="block text-center text-sm text-gray-500 hover:underline">
            Back
          </Link>
        </form>
      </Card>
    </div>
  )
}
