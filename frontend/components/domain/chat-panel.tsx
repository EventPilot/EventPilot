'use client'

import { useEffect, useMemo, useRef, useState, useTransition } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/cn'
import { approveAgentRunAction, sendChatMessageAction, type AgentRun } from '@/app/dashboard/events/[id]/post/actions'
import { createClient } from '@/lib/supabase/client'

type Msg = { role: 'user' | 'assistant'; text: string }

function apiBaseUrl() {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080'
}

type RunEvent = {
  type: string
  run: AgentRun
}

export function ChatPanel({
  eventId,
  initialMessages,
}: {
  eventId: string
  initialMessages: Msg[]
}) {
  const [messages, setMessages] = useState<Msg[]>(initialMessages)
  const [input, setInput] = useState('')
  const [isPending, startTransition] = useTransition()
  const [run, setRun] = useState<AgentRun | null>(null)
  const [runError, setRunError] = useState<string | null>(null)
  const streamRef = useRef<EventSource | null>(null)

  const chips = useMemo(() => ['Shorten', 'More technical', 'Add CTA', 'Make it hype', 'LinkedIn version'], [])

  useEffect(() => {
    return () => {
      streamRef.current?.close()
    }
  }, [])

  useEffect(() => {
    let cancelled = false

    async function hydrateRun() {
      const supabase = createClient()
      const { data } = await supabase.auth.getSession()
      const token = data.session?.access_token
      if (!token) return

      const response = await fetch(
        `${apiBaseUrl()}/api/events/${eventId}/chat/run`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
          cache: 'no-store',
        }
      )

      if (response.status === 204 || !response.ok) return
      const payload = await response.json() as AgentRun
      if (!cancelled) {
        setRun(payload)
      }
    }

    void hydrateRun()
    return () => {
      cancelled = true
    }
  }, [eventId])

  useEffect(() => {
    if (!run) return
    if (run.status === 'completed' || run.status === 'failed' || run.status === 'cancelled') {
      streamRef.current?.close()
      return
    }
    void openRunStream(run.id)
  }, [run?.id])

  async function openRunStream(runId: string) {
    streamRef.current?.close()
    const supabase = createClient()
    const { data } = await supabase.auth.getSession()
    const token = data.session?.access_token
    if (!token) return

    const source = new EventSource(
      `${apiBaseUrl()}/api/agent-runs/${runId}/stream?access_token=${encodeURIComponent(token)}`
    )
    streamRef.current = source

    const handleEvent = (event: MessageEvent<string>) => {
      const payload = JSON.parse(event.data) as RunEvent
      setRun(payload.run)
      if (payload.run.status === 'completed' || payload.run.status === 'failed') {
        source.close()
      }
    }

    source.addEventListener('run_state', handleEvent)
    source.addEventListener('task_state', handleEvent)
    source.addEventListener('blocked', handleEvent)
    source.addEventListener('completed', handleEvent)
    source.onerror = () => {
      setRunError('Lost connection to the run stream.')
      source.close()
    }
  }

  async function approveRun() {
    if (!run) return
    setRunError(null)
    const approved = await approveAgentRunAction(run.id)
    setRun(approved)
  }

  function send(text: string) {
    const t = text.trim()
    if (!t) return
    setMessages((m) => [...m, { role: 'user', text: t }])
    setInput('')
    startTransition(async () => {
      try {
        setRunError(null)
        const result = await sendChatMessageAction(eventId, t)
        if (result.run) {
          setRun(result.run)
          setMessages((m) => [
            ...m,
            {
              role: 'assistant',
              text: result.run?.plan_summary || 'Plan ready for approval.',
            },
          ])
        }
      } catch (error) {
        setRunError(error instanceof Error ? error.message : 'Failed to send message')
      }
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

      {run ? (
        <div className="mt-4 rounded-2xl border border-gray-200 bg-white p-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <div className="text-sm font-semibold">Agent plan</div>
              <div className="text-xs text-gray-500 capitalize">{run.status.replaceAll('_', ' ')}</div>
            </div>
            {run.status === 'awaiting_approval' ? (
              <Button onClick={approveRun} disabled={isPending}>Approve</Button>
            ) : null}
          </div>
          <div className="mt-3 text-sm text-gray-700">{run.plan_summary}</div>
          <div className="mt-3 space-y-2">
            {run.tasks.map((task, idx) => (
              <div key={task.id} className="rounded-xl border border-gray-200 px-3 py-2 text-xs">
                <div className="font-medium">
                  {idx + 1}. {task.title}
                </div>
                <div className="mt-1 text-gray-500 capitalize">{task.status.replaceAll('_', ' ')}</div>
                {task.target_user_name ? (
                  <div className="mt-1 text-gray-500">
                    Waiting on {task.target_user_name}{task.target_role ? ` (${task.target_role})` : ''}
                  </div>
                ) : null}
                {task.result ? <div className="mt-1 text-gray-600">{task.result}</div> : null}
              </div>
            ))}
          </div>
        </div>
      ) : null}

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

      {runError ? (
        <div className="mt-3 text-xs text-red-600">{runError}</div>
      ) : null}

      <div className="mt-4 rounded-full border border-gray-200 bg-indigo-50 px-4 py-2 text-center text-xs text-gray-600">
        Tip: Changes can auto-sync to the post preview.
      </div>
    </Card>
  )
}
