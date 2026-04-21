import { createBrowserClient } from '@supabase/ssr'
export function createClient() {
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL
  const supabaseKey =
    process.env.NEXT_PUBLIC_SUPABASE_API_KEY ?? process.env.SUPABASE_API_KEY

  return createBrowserClient(
    supabaseUrl!,
    supabaseKey!
  )
}
