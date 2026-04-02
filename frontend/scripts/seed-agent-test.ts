import { readFileSync, existsSync } from 'fs'
import { resolve } from 'path'
import { fileURLToPath } from 'url'
import { createClient, type Session, type User } from '@supabase/supabase-js'

type AuthResult = {
  user: User
  session: Session
  client: ReturnType<typeof createClient>
}

type EventRow = {
  id: string
  title: string
  description: string | null
  event_date: string
  location: string | null
  status: string | null
}

type EventMemberRow = {
  user_id: string
  role: string
  user?: {
    id: string
    name?: string | null
    email?: string | null
  } | null
}

type AgentTask = {
  id: string
  title: string
  kind: string
  status: string
}

type AgentRun = {
  id: string
  status: string
  plan_summary: string
  tasks: AgentTask[]
}

type ChatResponse = {
  chat_id: string
  run: AgentRun | null
}

const scriptDir = resolve(fileURLToPath(new URL('.', import.meta.url)))
const root = resolve(scriptDir, '..', '..')
const apiEnvPath = resolve(root, 'api', '.env')
const apiEnv = loadEnvFile(apiEnvPath)

const supabaseUrl =
  process.env.NEXT_PUBLIC_SUPABASE_URL ??
  process.env.SUPABASE_URL ??
  apiEnv.SUPABASE_URL

const supabaseAnonKey =
  process.env.NEXT_PUBLIC_SUPABASE_API_KEY ??
  process.env.SUPABASE_API_KEY ??
  apiEnv.SUPABASE_API_KEY

const apiBaseUrl =
  process.env.EVENTPILOT_API_BASE_URL ??
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  'http://localhost:8080'

const testUserEmail = process.env.TEST_USER_EMAIL ?? 'test1@fake.com'
const testUserPassword = process.env.TEST_USER_PASSWORD ?? '123456'
const testUserID = process.env.TEST_USER_ID ?? 'bb8b5a53-fd77-4a3b-ab38-51b36836c9bd'
const testEventID = process.env.TEST_EVENT_ID ?? '7db5a07b-7c77-4084-b174-0b50983f632e'
const fakeMessage =
  process.env.TEST_FAKE_MESSAGE ??
  'We need a post for Flight Test. Please review what is missing, figure out whether anyone else on the event needs to be asked for input, and prepare the next steps.'

if (!supabaseUrl || !supabaseAnonKey) {
  throw new Error('Missing Supabase URL or anon key. Set env vars or populate api/.env.')
}

async function main() {
  const auth = await loginExistingUser()
  if (auth.user.id !== testUserID) {
    throw new Error(`Authenticated unexpected user. Expected ${testUserID}, got ${auth.user.id}`)
  }

  const event = await loadEvent(auth.client)
  const members = await loadEventMembers(auth.client)
  const currentMembership = members.find((member) => member.user_id === auth.user.id)
  if (!currentMembership) {
    throw new Error(`User ${auth.user.id} is not a member of event ${testEventID}`)
  }

  const otherMembers = members.filter((member) => member.user_id !== auth.user.id)
  const response = await sendFakeMessage(auth.session.access_token, testEventID, fakeMessage)

  console.log('Agent run test complete\n')
  console.log(`User: ${auth.user.email} (${auth.user.id})`)
  console.log(`Event: ${event.title} (${event.id})`)
  console.log(`Role on event: ${currentMembership.role}`)
  console.log('Other event members:')
  if (otherMembers.length === 0) {
    console.log('  none')
  } else {
    for (const member of otherMembers) {
      const label = member.user?.name ?? member.user?.email ?? member.user_id
      console.log(`  - ${label} (${member.user_id}) role=${member.role}`)
    }
  }
  console.log(`\nFake message: ${fakeMessage}`)
  console.log(`Chat ID: ${response.chat_id}`)
  if (!response.run) {
    throw new Error('Chat endpoint returned no run payload')
  }
  console.log(`Run ID: ${response.run.id}`)
  console.log(`Run status: ${response.run.status}`)
  console.log(`Plan summary: ${response.run.plan_summary}`)
  console.log('Tasks:')
  for (const task of response.run.tasks) {
    console.log(`  - [${task.status}] ${task.kind}: ${task.title}`)
  }
}

async function loginExistingUser(): Promise<AuthResult> {
  const client = createClient(supabaseUrl!, supabaseAnonKey!, {
    auth: {
      autoRefreshToken: false,
      persistSession: false,
    },
  })

  const login = await client.auth.signInWithPassword({
    email: testUserEmail,
    password: testUserPassword,
  })

  if (login.error || !login.data.user || !login.data.session) {
    throw login.error ?? new Error(`Unable to authenticate ${testUserEmail}`)
  }

  return {
    user: login.data.user,
    session: login.data.session,
    client,
  }
}

async function loadEvent(client: ReturnType<typeof createClient>): Promise<EventRow> {
  const { data, error } = await client
    .from('event')
    .select('id, title, description, event_date, location, status')
    .eq('id', testEventID)
    .single()

  if (error || !data) {
    throw error ?? new Error(`Event ${testEventID} not found`)
  }

  return data
}

async function loadEventMembers(client: ReturnType<typeof createClient>): Promise<EventMemberRow[]> {
  const { data, error } = await client
    .from('event_member')
    .select('user_id, role, user(id, name, email)')
    .eq('event_id', testEventID)

  if (error || !data) {
    throw error ?? new Error(`Failed to load members for event ${testEventID}`)
  }

  return data
}

async function sendFakeMessage(accessToken: string, eventId: string, message: string): Promise<ChatResponse> {
  const response = await fetch(`${apiBaseUrl}/api/events/${eventId}/chat/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ message }),
  })

  if (!response.ok) {
    throw new Error(`Chat endpoint failed: ${response.status} ${await response.text()}`)
  }

  return response.json() as Promise<ChatResponse>
}

function loadEnvFile(path: string): Record<string, string> {
  if (!existsSync(path)) return {}
  const text = readFileSync(path, 'utf8')
  const out: Record<string, string> = {}
  for (const line of text.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue
    const eq = trimmed.indexOf('=')
    if (eq < 0) continue
    const key = trimmed.slice(0, eq).trim()
    const rawValue = trimmed.slice(eq + 1).trim()
    out[key] = rawValue.replace(/^['"]|['"]$/g, '')
  }
  return out
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
