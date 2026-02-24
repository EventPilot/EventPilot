import { Card } from '@/components/ui/card'
import type { TimelineItem } from '@/lib/data/events'

export function Timeline({ items }: { items: TimelineItem[] }) {
  return (
    <Card className="p-6">
      <div className="text-lg font-semibold">Timeline</div>
      <div className="mt-5 space-y-3">
        {items.map((it) => (
          <div key={it.title} className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
            <div className="text-sm font-medium">{it.title}</div>
            <div className="text-xs text-gray-500 mt-1">{it.meta}</div>
          </div>
        ))}
      </div>
    </Card>
  )
}
