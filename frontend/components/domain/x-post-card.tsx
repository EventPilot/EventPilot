import { Card } from '@/components/ui/card'

export function XPostCard({
  accountName = 'EventPilot',
  handle = '@eventpilot',
  content,
  hashtags,
}: {
  accountName?: string
  handle?: string
  content: string
  hashtags: string
}) {
  const lines = content.split('\n').filter(Boolean)

  return (
    <Card className="p-5">
      <div className="flex items-start gap-3">
        <div className="h-9 w-9 rounded-full bg-gray-200 border border-gray-200" />
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <div className="text-sm font-medium">{accountName}</div>
            <div className="text-xs text-gray-500">{handle} • 2h</div>
          </div>

          <div className="mt-3 space-y-1 text-sm">
            {lines.map((l, i) => (
              <div key={i}>{l}</div>
            ))}
            {hashtags && <div className="pt-2 text-gray-700">{hashtags}</div>}
          </div>

          <div className="mt-4 rounded-2xl border border-gray-200 bg-gray-100 p-3">
            <div className="grid grid-cols-3 gap-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <div
                  key={i}
                  className="h-40 rounded-xl border border-gray-200 bg-gray-200 flex items-center justify-center text-xs text-gray-600"
                >
                  IMG
                </div>
              ))}
            </div>
          </div>

          <div className="mt-4 border-t border-gray-200 pt-3 flex items-center justify-between text-xs text-gray-500">
            <span>💬 12</span>
            <span>🔁 4</span>
            <span>❤️ 38</span>
            <span>📤 Share</span>
          </div>
        </div>
      </div>
    </Card>
  )
}
