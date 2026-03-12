'use client'

import { useMemo, useState, useTransition } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/cn'
import { sendChatMessageAction } from '@/app/dashboard/events/[id]/post/actions'

type Msg = { role: 'user' | 'assistant'; text: string }

export function ChatPanel({ eventId }: { eventId: string }) {
  const [messages, setMessages] = useState<Msg[]>([
    {
      role: 'assistant',
      text: 'I can help refine this post. Want it shorter, more technical, or with a stronger CTA?',
    },
    { role: 'user', text: 'Make it more technical and mention the key metric.' },
    {
      role: 'assistant',
      text: 'Draft option: Completed customer demo with clean stage separation and stable ascent. Metric: 98% nominal trajectory match. Want a 1-sentence version too?',
    },
    { role: 'user', text: 'Yes, 1 sentence.' },
    { role: 'assistant', text: 'Customer demo complete: stable ascent, clean separation, 98% trajectory match.' },
  ])
  const [input, setInput] = useState('')
  const [isPending, startTransition] = useTransition()

  const chips = useMemo(() => ['Shorten', 'More technical', 'Add CTA', 'Make it hype', 'LinkedIn version'], [])

  function send(text: string) {
    const t = text.trim()
    if (!t) return
    setMessages((m) => [...m, { role: 'user', text: t }])
    setInput('')
    startTransition(async () => {
      await sendChatMessageAction(eventId, t)
      setMessages((m) => [...m, { role: 'assistant', text: 'Got it — working on it. (LLM response placeholder)' }])
    })
  }

  return (
    <Card className="p-5 flex flex-col h-full min-h-0">
      <div className="text-lg font-semibold">Chat with AI</div>
      <div className="text-sm text-gray-500 mt-1">Refine the post by talking to the model.</div>

      <div className="mt-4 flex flex-wrap gap-2">
        {chips.map((c) => (
          <button
            key={c}
            onClick={() => send(c)}
            className="rounded-full border border-gray-200 bg-gray-50 px-3 py-2 text-xs text-gray-600 hover:bg-gray-100"
          >
            {c}
          </button>
        ))}
      </div>

      <div className="mt-4 flex-1 min-h-0 overflow-auto rounded-2xl border border-gray-200 bg-gray-50 p-4 space-y-3">
        {messages.map((m, idx) => (
          <div
            key={idx}
            className={cn(
              'max-w-[90%] rounded-2xl border border-gray-200 p-3 text-xs leading-5',
              m.role === 'user' ? 'ml-auto bg-blue-100' : 'bg-white'
            )}
          >
            {m.text}
          </div>
        ))}
      </div>

      <div className="mt-4 flex items-center gap-2">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && send(input)}
          placeholder="Ask: rewrite, add CTA, generate LinkedIn..."
          className="h-12 flex-1 rounded-2xl border border-gray-200 bg-white px-4 text-sm outline-none"
        />
        <Button onClick={() => send(input)} disabled={isPending}>Send</Button>
      </div>

      <div className="mt-4 rounded-full border border-gray-200 bg-indigo-50 px-4 py-2 text-center text-xs text-gray-600">
        Tip: Changes can auto-sync to the post preview.
      </div>
    </Card>
  )
}
