import { createClient } from '@supabase/supabase-js'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
const serviceKey = process.env.SUPABASE_SERVICE_ROLE_KEY!

const supabase = createClient(supabaseUrl, serviceKey, {
  auth: {
    autoRefreshToken: false,
    persistSession: false
  }
})

async function createUserWithProfile() {
  const email = 'alice@test.com'
  const password = 'test123456'
  const name = 'Alice Test'

  // Create auth user
  const { data: authData, error: authError } = await supabase.auth.admin.createUser({
    email,
    password,
    email_confirm: true,
  })

  if (authError) {
    console.error('Auth error:', authError)
    return
  }

  console.log('✅ Auth user created:', authData.user.id)

  // Create profile in public.users
  const { error: profileError } = await supabase
    .from('users')
    .insert({
      id: authData.user.id,
      name: name
    })

  if (profileError) {
    console.error('❌ Profile error:', profileError)
    // Clean up auth user if profile creation fails
    await supabase.auth.admin.deleteUser(authData.user.id)
    console.log('Rolled back auth user')
    return
  }

  console.log('✅ User profile created')
  console.log('\nTest credentials:')
  console.log('Email:', email)
  console.log('Password:', password)
  console.log('User ID:', authData.user.id)
}

createUserWithProfile()