'use client'

import { useEffect, useRef, useState, useTransition } from 'react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/cn'
import {
  approveAgentRunAction,
  getLatestAgentRunAction,
  sendChatMessageAction,
  type ChatMessageEntry,
  type AgentRun,
} from '@/app/dashboard/events/[id]/post/actions'
import { createClient } from '@/lib/supabase/client'

type RunEvent = {
  type: string
  run: AgentRun
}

type TimelineItem =
  | { kind: 'message'; createdAt: string; id: string; entry: ChatMessageEntry }
  | { kind: 'run'; createdAt: string; id: string; run: AgentRun }

const activeRunStatuses = new Set(['planning', 'awaiting_approval', 'running', 'waiting_on_member'])
const terminalRunStatuses = new Set(['completed', 'failed', 'cancelled'])

function apiBaseUrl() {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080'
}

function formatTimestamp(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const month = months[date.getUTCMonth()]
  const day = date.getUTCDate()
  const rawHour = date.getUTCHours()
  const minute = String(date.getUTCMinutes()).padStart(2, '0')
  const suffix = rawHour >= 12 ? 'PM' : 'AM'
  const hour = rawHour % 12 || 12

  return `${month} ${day}, ${hour}:${minute} ${suffix}`
}

function formatStatus(status: string) {
  return status.replaceAll('_', ' ')
}

function isActiveRun(status: string) {
  return activeRunStatuses.has(status)
}

function taskTone(status: string) {
  switch (status) {
    case 'completed':
      return 'border-emerald-200 bg-emerald-50 text-emerald-950 dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-200'
    case 'in_progress':
      return 'border-blue-200 bg-blue-50 text-blue-950 dark:border-blue-900 dark:bg-blue-950/40 dark:text-blue-200'
    case 'waiting':
      return 'border-amber-200 bg-amber-50 text-amber-950 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-200'
    case 'failed':
      return 'border-red-200 bg-red-50 text-red-950 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200'
    default:
      return 'border-slate-200 bg-white/80 text-slate-900 dark:border-slate-700 dark:bg-slate-900/80 dark:text-slate-100'
  }
}

function statusTone(status: string) {
  switch (status) {
    case 'awaiting_approval':
      return 'border-sky-200 bg-sky-50 text-sky-900 dark:border-sky-900 dark:bg-sky-950/40 dark:text-sky-200'
    case 'running':
      return 'border-blue-200 bg-blue-50 text-blue-900 dark:border-blue-900 dark:bg-blue-950/40 dark:text-blue-200'
    case 'waiting_on_member':
      return 'border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-200'
    case 'completed':
      return 'border-emerald-200 bg-emerald-50 text-emerald-900 dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-200'
    case 'failed':
      return 'border-red-200 bg-red-50 text-red-900 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200'
    default:
      return 'border-slate-200 bg-slate-100 text-slate-800 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200'
  }
}

function currentTaskTitle(run: AgentRun) {
  const activeTask = run.tasks.find((task) => task.status === 'in_progress' || task.status === 'waiting')
  return activeTask?.title ?? run.tasks[run.current_task_index]?.title ?? ''
}

function RunArtifact({
  run,
  isPending,
  onApprove,
}: {
  run: AgentRun
  isPending: boolean
  onApprove: () => void
}) {
  const headline = currentTaskTitle(run)

  return (
    <div className="rounded-[28px] border border-slate-200 bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(239,246,255,0.92))] p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)] dark:border-slate-700 dark:bg-[linear-gradient(180deg,rgba(15,23,42,0.96),rgba(2,6,23,0.98))] dark:shadow-[0_18px_50px_rgba(2,6,23,0.45)]">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="text-[11px] font-semibold uppercase tracking-[0.24em] text-slate-500 dark:text-slate-400">Agent run</div>
          <div className="mt-2 text-xl font-semibold text-slate-950 dark:text-slate-100">
            {run.plan_summary || 'Working through the next publishing steps.'}
          </div>
          {headline ? (
            <div className="mt-2 text-sm text-slate-600 dark:text-slate-300">Current focus: {headline}</div>
          ) : null}
        </div>
        <div className={cn('rounded-full border px-3 py-1 text-xs font-semibold capitalize', statusTone(run.status))}>
          {formatStatus(run.status)}
        </div>
      </div>

      <div className="mt-5 grid gap-3">
        {run.tasks.map((task, idx) => (
          <div key={task.id} className={cn('rounded-2xl border px-4 py-3', taskTone(task.status))}>
            <div className="flex items-start justify-between gap-3">
              <div>
                <div className="text-sm font-semibold">
                  {idx + 1}. {task.title}
                </div>
                {task.instructions ? (
                  <div className="mt-1 text-xs leading-5 text-slate-600 dark:text-slate-300">{task.instructions}</div>
                ) : null}
              </div>
              <div className="text-[11px] font-semibold uppercase tracking-[0.18em] opacity-75">
                {formatStatus(task.status)}
              </div>
            </div>
            {task.target_user_name ? (
              <div className="mt-2 text-xs text-slate-600 dark:text-slate-300">
                Waiting on {task.target_user_name}
                {task.target_role ? ` • ${task.target_role}` : ''}
              </div>
            ) : null}
            {task.result ? <div className="mt-2 text-xs text-slate-700 dark:text-slate-200">{task.result}</div> : null}
          </div>
        ))}
      </div>

      {run.status === 'awaiting_approval' ? (
        <div className="mt-5 flex items-center justify-between gap-3 rounded-2xl border border-sky-200 bg-sky-50/80 px-4 py-3 dark:border-sky-900 dark:bg-sky-950/30">
          <div className="text-sm text-sky-950 dark:text-sky-200">This plan is ready to start.</div>
          <Button onClick={onApprove} disabled={isPending}>Approve plan</Button>
        </div>
      ) : null}
    </div>
  )
}

export function AgentWorkspace({
  eventId,
  initialEntries,
  initialRun,
}: {
  eventId: string
  initialEntries: ChatMessageEntry[]
  initialRun: AgentRun | null
}) {
  const [entries, setEntries] = useState(initialEntries)
  const [input, setInput] = useState('')
  const [run, setRun] = useState<AgentRun | null>(initialRun)
  const [runError, setRunError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()
  const streamRef = useRef<EventSource | null>(null)
  const scrollRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    return () => {
      streamRef.current?.close()
    }
  }, [])

  useEffect(() => {
    if (!run || !isActiveRun(run.status)) {
      streamRef.current?.close()
      return
    }

    void openRunStream(run.id)
  }, [run?.id, run?.status])

  async function refreshLatestRun() {
    try {
      const latestRun = await getLatestAgentRunAction(eventId)
      setRun(latestRun)
    } catch (error) {
      setRunError(error instanceof Error ? error.message : 'Failed to load the latest agent state')
    }
  }

  async function openRunStream(runId: string) {
    streamRef.current?.close()

    const supabase = createClient()
    const { data } = await supabase.auth.getSession()
    const token = data.session?.access_token
    if (!token) {
      return
    }

    const source = new EventSource(`${apiBaseUrl()}/api/agent-runs/${runId}/stream?access_token=${encodeURIComponent(token)}`)
    streamRef.current = source

    const handleEvent = (event: MessageEvent<string>) => {
      const payload = JSON.parse(event.data) as RunEvent
      setRun(payload.run)
      if (terminalRunStatuses.has(payload.run.status)) {
        source.close()
        window.setTimeout(() => {
          void refreshLatestRun()
        }, 150)
      }
    }

    source.addEventListener('run_state', handleEvent)
    source.addEventListener('task_state', handleEvent)
    source.addEventListener('blocked', handleEvent)
    source.addEventListener('completed', handleEvent)
    source.onerror = () => {
      source.close()
      void refreshLatestRun()
    }
  }

  async function approveRun() {
    if (!run) {
      return
    }

    try {
      setRunError(null)
      const approved = await approveAgentRunAction(run.id)
      setRun(approved)
    } catch (error) {
      setRunError(error instanceof Error ? error.message : 'Failed to approve the plan')
    }
  }

  function submitInstruction(text: string) {
    const nextText = text.trim()
    if (!nextText) {
      return
    }

    const tempId = `local-${Date.now()}`
    const optimisticEntry: ChatMessageEntry = {
      id: tempId,
      role: 'user',
      text: nextText,
      createdAt: new Date().toISOString(),
      messageType: 'message',
      metadata: {},
    }

    setEntries((current) => [...current, optimisticEntry])
    setInput('')

    startTransition(async () => {
      try {
        setRunError(null)
        const result = await sendChatMessageAction(eventId, nextText)
        setEntries((current) =>
          current.map((entry) =>
            entry.id === tempId
              ? {
                  id: result.message.id,
                  role: 'user',
                  text: result.message.message,
                  createdAt: result.message.created_at,
                  messageType: result.message.message_type,
                  metadata: result.message.metadata ?? {},
                }
              : entry
          )
        )
        if (result.run) {
          setRun(result.run)
          return
        }

        await refreshLatestRun()
      } catch (error) {
        setEntries((current) => current.filter((entry) => entry.id !== tempId))
        setRunError(error instanceof Error ? error.message : 'Failed to send instruction')
      }
    })
  }

  const visibleEntries = entries.filter((entry) => entry.messageType !== 'approval_request')
  const timeline: TimelineItem[] = visibleEntries.map((entry) => ({
    kind: 'message',
    createdAt: entry.createdAt,
    id: `message-${entry.id}`,
    entry,
  }))
  if (run) {
    timeline.push({
      kind: 'run',
      createdAt: run.created_at,
      id: `run-${run.id}`,
      run,
    })
  }
  timeline.sort((a, b) => {
    const timeDiff = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
    if (timeDiff !== 0) {
      return timeDiff
    }
    return a.kind === 'message' ? -1 : 1
  })

  function isQueuedMessage(entry: ChatMessageEntry) {
    const state = entry.metadata?.workflow_state
    return typeof state === 'string' && state === 'queued'
  }

  useEffect(() => {
    const container = scrollRef.current
    if (!container) {
      return
    }

    container.scrollTop = container.scrollHeight
  }, [timeline.map((item) => item.id).join('|')])

  return (
    <section className="flex h-full min-h-0 flex-col overflow-hidden rounded-[36px] bg-[radial-gradient(circle_at_top_left,rgba(219,234,254,0.85),transparent_38%),linear-gradient(180deg,#ffffff_0%,#f8fafc_100%)] p-5 shadow-[0_24px_80px_rgba(15,23,42,0.08)] ring-1 ring-slate-200/80 dark:bg-[radial-gradient(circle_at_top_left,rgba(30,64,175,0.24),transparent_38%),linear-gradient(180deg,#0f172a_0%,#020617_100%)] dark:shadow-[0_24px_80px_rgba(2,6,23,0.5)] dark:ring-slate-800">
      <div ref={scrollRef} className="flex-1 overflow-y-auto pr-1">
        <div className="space-y-3">
          {timeline.map((item) => {
            if (item.kind === 'run') {
              return <RunArtifact key={item.id} run={item.run} isPending={isPending} onApprove={approveRun} />
            }

            const queued = isQueuedMessage(item.entry)
            return (
              <div
                key={item.id}
                className={cn(
                  'w-full rounded-[26px] border px-4 py-3 shadow-sm',
                  item.entry.role === 'user'
                    ? queued
                      ? 'border-violet-200 bg-violet-50/90 text-slate-950 dark:border-violet-900 dark:bg-violet-950/40 dark:text-slate-100'
                      : 'border-blue-200 bg-blue-50/90 text-slate-950 dark:border-blue-900 dark:bg-blue-950/40 dark:text-slate-100'
                    : 'border-amber-200 bg-amber-50/95 text-slate-950 dark:border-amber-900 dark:bg-amber-950/40 dark:text-slate-100'
                )}
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
                    {item.entry.role === 'user' ? (queued ? 'Queued' : 'Instruction') : 'Agent update'}
                  </div>
                  <div className="text-[11px] text-slate-500 dark:text-slate-400">{formatTimestamp(item.entry.createdAt)}</div>
                </div>
                <div className="mt-2 text-sm leading-6 whitespace-pre-line">{item.entry.text}</div>
              </div>
            )
          })}

          {!run && visibleEntries.length === 0 ? (
            <div className="rounded-[28px] border border-dashed border-slate-300 bg-white/70 px-5 py-8 text-center text-sm text-slate-500 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-400">
              No saved agent activity yet. Add an instruction below to kick off the next workflow.
            </div>
          ) : null}
        </div>
      </div>

      <div className="mt-5 border-t border-slate-200/80 pt-4 dark:border-slate-800">
        <input
          value={input}
          disabled={isPending}
          onChange={(event) => setInput(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter' && !event.shiftKey) {
              event.preventDefault()
              submitInstruction(input)
            }
          }}
          placeholder="Give the agent a new instruction and press Enter"
          className="h-14 w-full rounded-[22px] border border-slate-200 bg-white/90 px-5 text-sm text-slate-900 outline-none transition focus:border-blue-300 focus:ring-4 focus:ring-blue-100 dark:border-slate-700 dark:bg-slate-900/90 dark:text-slate-100 dark:placeholder:text-slate-500 dark:focus:border-blue-700 dark:focus:ring-blue-950/60"
        />

        {runError ? <div className="mt-3 text-xs text-red-600 dark:text-red-400">{runError}</div> : null}
      </div>
    </section>
  )
}
