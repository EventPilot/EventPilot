'use server'

import { createClient } from '@/lib/supabase/server'
import { redirect } from 'next/navigation'
import { randomUUID } from 'crypto'

export async function sendChatMessageAction(eventId: string, message: string) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) redirect('/login')

  // Look up the chat for this event, or create one if it doesn't exist
  let { data: chat } = await supabase
    .from('chat')
    .select('id')
    .eq('event_id', eventId)
    .single()

  if (!chat) {
    const { data: newChat, error: createError } = await supabase
      .from('chat')
      .insert({ id: randomUUID(), event_id: eventId })
      .select('id')
      .single()
    if (createError) throw new Error(createError.message)
    chat = newChat
  }

  const { error } = await supabase.from('chat_message').insert({
    id: randomUUID(),
    chat_id: chat!.id,
    sender_type: 'user',
    sender_id: user.id,
    message,
  })

  if (error) throw new Error(error.message)
}
