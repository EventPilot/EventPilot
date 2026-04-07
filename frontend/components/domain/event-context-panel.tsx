import { Card } from '@/components/ui/card'

type ContextValue = string | number | boolean | null | ContextValue[] | { [key: string]: ContextValue }

function formatLabel(value: string) {
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, (match) => match.toUpperCase())
}

function renderValue(value: ContextValue, keyPrefix: string): React.ReactNode {
  if (value == null) {
    return <div className="text-sm text-slate-500 dark:text-slate-400">No value yet</div>
  }

  if (Array.isArray(value)) {
    if (value.length === 0) {
      return <div className="text-sm text-slate-500 dark:text-slate-400">No items yet</div>
    }

    return (
      <div className="space-y-2">
        {value.map((item, index) => (
          <div
            key={`${keyPrefix}-${index}`}
            className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2 dark:border-slate-700 dark:bg-slate-900"
          >
            {typeof item === 'object' && item !== null ? (
              <div className="space-y-2">
                {Object.entries(item).map(([nestedKey, nestedValue]) => (
                  <div key={`${keyPrefix}-${index}-${nestedKey}`}>
                    <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
                      {formatLabel(nestedKey)}
                    </div>
                    <div className="mt-1 text-sm text-slate-700 dark:text-slate-200">
                      {typeof nestedValue === 'object' && nestedValue !== null
                        ? JSON.stringify(nestedValue)
                        : String(nestedValue)}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-sm text-slate-700 dark:text-slate-200">{String(item)}</div>
            )}
          </div>
        ))}
      </div>
    )
  }

  if (typeof value === 'object') {
    const entries = Object.entries(value)
    if (entries.length === 0) {
      return <div className="text-sm text-slate-500 dark:text-slate-400">No details yet</div>
    }

    return (
      <div className="space-y-3">
        {entries.map(([nestedKey, nestedValue]) => (
          <div
            key={`${keyPrefix}-${nestedKey}`}
            className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2 dark:border-slate-700 dark:bg-slate-900"
          >
            <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
              {formatLabel(nestedKey)}
            </div>
            <div className="mt-1 text-sm text-slate-700 dark:text-slate-200">
              {typeof nestedValue === 'object' && nestedValue !== null ? JSON.stringify(nestedValue) : String(nestedValue)}
            </div>
          </div>
        ))}
      </div>
    )
  }

  return <div className="text-sm leading-6 text-slate-700 dark:text-slate-200">{String(value)}</div>
}

export function EventContextPanel({ context, mediaUrls = [] }: { context?: Record<string, ContextValue> | null; mediaUrls?: string[] }) {
  const entries = context ? Object.entries(context) : []

  return (
    <div className="grid gap-6 xl:grid-cols-2">
      <Card className="p-6">
        <div className="text-lg font-semibold">Event context ideas</div>
        <div className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          Structured notes the post workflow can use while collecting inputs and drafting copy.
        </div>

        {entries.length === 0 ? (
          <div className="mt-5 rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-5 py-8 text-sm text-slate-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-400">
            No event context has been saved yet.
          </div>
        ) : (
          <div className="mt-5 grid gap-4 md:grid-cols-2">
            {entries.map(([key, value]) => (
              <div
                key={key}
                className="rounded-[28px] border border-slate-200 bg-[linear-gradient(180deg,#ffffff_0%,#f8fafc_100%)] p-4 shadow-sm dark:border-slate-700 dark:bg-[linear-gradient(180deg,#0f172a_0%,#020617_100%)]"
              >
                <div className="text-[11px] font-semibold uppercase tracking-[0.24em] text-slate-500 dark:text-slate-400">
                  {formatLabel(key)}
                </div>
                <div className="mt-3">{renderValue(value, key)}</div>
              </div>
            ))}
          </div>
        )}
      </Card>

      <Card className="p-6">
        <div className="text-lg font-semibold">Uploaded Media</div>
        <div className="mt-1 text-sm text-slate-500 dark:text-slate-400">
          Images uploaded for this event.
        </div>

        {mediaUrls.length === 0 ? (
          <div className="mt-5 rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-5 py-8 text-sm text-slate-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-400">
            No media has been uploaded yet.
          </div>
        ) : (
          <div className="mt-5 flex flex-wrap justify-center gap-4">
            {mediaUrls.map((url, i) => (
              <div
                key={i}
                className="w-52 overflow-hidden rounded-2xl border border-slate-200 dark:border-slate-700"
              >
                <img src={url} alt={`Event media ${i + 1}`} className="h-48 w-full object-cover" />
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
