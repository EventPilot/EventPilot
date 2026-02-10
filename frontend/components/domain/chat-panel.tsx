'use client'

import { useMemo, useState } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/cn'

type Msg = { role: 'user' | 'assistant'; text: string }

export function ChatPanel() {
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

  const chips = useMemo(() => ['Shorten', 'More technical', 'Add CTA', 'Make it hype', 'LinkedIn version'], [])

  function send(text: string) {
    const t = text.trim()
    if (!t) return
    setMessages((m) => [...m, { role: 'user', text: t }])
    setInput('')
    setTimeout(() => {
      setMessages((m) => [...m, { role: 'assistant', text: 'Got it — I’ll rewrite it in that direction. (LLM response placeholder)' }])
    }, 400)
  }

  return (
    <Card className="p-5">
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

      <div className="mt-4 h-[480px] overflow-auto rounded-2xl border border-gray-200 bg-gray-50 p-4 space-y-3">
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
          placeholder="Ask: rewrite, add CTA, generate LinkedIn…"
          className="h-12 flex-1 rounded-2xl border border-gray-200 bg-white px-4 text-sm outline-none"
        />
        <Button onClick={() => send(input)}>Send</Button>
      </div>

      <div className="mt-4 rounded-full border border-gray-200 bg-indigo-50 px-4 py-2 text-center text-xs text-gray-600">
        Tip: Changes can auto-sync to the post preview.
      </div>
    </Card>
  )
}
