import { Card } from '@/components/ui/card'

export function XPostCard({
  accountName = 'Event Pilot',
  handle = '@eventpilot.app',
  content,
}: {
  accountName?: string
  handle?: string
  content: string
}) {
  const lines = content.split('\n').filter(Boolean)

  return (
    <Card className="h-full rounded-[30px] border border-sky-200/80 bg-[linear-gradient(180deg,#f8fbff_0%,#ffffff_100%)] p-5 shadow-[0_20px_60px_rgba(14,30,37,0.08)] dark:border-slate-700 dark:bg-[linear-gradient(180deg,#0b1220_0%,#111827_100%)] dark:shadow-[0_20px_60px_rgba(2,6,23,0.45)]">
      <div className="flex items-start gap-3">
        <div className="flex h-11 w-11 items-center justify-center rounded-full border border-sky-200 bg-sky-50 text-sm font-semibold text-sky-700 dark:border-sky-900 dark:bg-sky-950/50 dark:text-sky-200">
          EP
        </div>
        <div className="flex-1">
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <div className="text-sm font-semibold text-slate-950 dark:text-slate-100">{accountName}</div>
              <div className="text-xs text-slate-500 dark:text-slate-400">{handle}</div>
            </div>
          </div>

          <div className="mt-4 space-y-2 text-[15px] leading-7 text-slate-900 dark:text-slate-100">
            {(lines.length > 0 ? lines : ['The draft will appear here when the agent or post generator creates one.']).map((l, i) => (
              <div key={i}>{l}</div>
            ))}
          </div>

          <div className="mt-5 rounded-[24px] border border-dashed border-sky-200 bg-sky-50/60 p-4 dark:border-slate-700 dark:bg-slate-900/70">
            <div className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">
              Media preview
            </div>
            <div className="mt-3 grid grid-cols-3 gap-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <div
                  key={i}
                  className="flex h-32 items-center justify-center rounded-[18px] border border-sky-200 bg-white text-xs font-medium text-slate-500 dark:border-slate-700 dark:bg-slate-950 dark:text-slate-400"
                >
                  Image slot
                </div>
              ))}
            </div>
          </div>

          <div className="mt-5 flex items-center justify-between border-t border-slate-200 pt-3 text-xs text-slate-500 dark:border-slate-700 dark:text-slate-400">
            <span>Reply</span>
            <span>Repost</span>
            <span>Like</span>
            <span>Share</span>
          </div>
        </div>
      </div>
    </Card>
  )
}
