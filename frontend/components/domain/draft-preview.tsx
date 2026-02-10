import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import type { Draft, RoleInput } from '@/lib/data/events'

export function DraftPreview({ draft, inputs }: { draft: Draft; inputs: RoleInput[] }) {
  return (
    <Card className="p-6">
      <div className="text-lg font-semibold">Draft preview</div>
      <div className="text-xs text-gray-500 mt-1">Generated text (editable)</div>

      <div className="mt-4 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-sm leading-6 whitespace-pre-line">
        {draft.text || 'Draft will appear here once generated.'}
      </div>

      <div className="mt-4 text-xs text-gray-500">Media</div>
      <div className="mt-2 grid grid-cols-3 gap-3">
        {(draft.media.length ? draft.media : [{ id: 'placeholder', label: 'IMG' }]).slice(0, 3).map((m) => (
          <div
            key={m.id}
            className="h-20 rounded-2xl border border-gray-200 bg-gray-200 flex items-center justify-center text-xs text-gray-600"
          >
            {m.label}
          </div>
        ))}
      </div>

      <div className="mt-4 text-xs text-gray-500">Hashtags & mentions</div>
      <div className="mt-2 rounded-2xl border border-gray-200 bg-gray-50 p-3 text-sm">
        {draft.hashtags || '#launch #aerospace'}
      </div>

      <div className="mt-4 flex items-center gap-3">
        <Button variant="secondary" size="sm">
          Regenerate
        </Button>
        <Button size="sm">Approve & Publish</Button>
      </div>

      <div className="mt-5 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-xs text-gray-600">
        <div className="font-medium text-gray-900">Generated from inputs</div>
        <div className="mt-2 space-y-1">
          {inputs.map((r) => (
            <div key={r.role}>
              {r.role}: {r.completeCount === r.totalCount ? 'complete' : '(pending)'}
            </div>
          ))}
        </div>
        <div className="mt-3">Tip: publishing is locked until required roles are complete.</div>
      </div>
    </Card>
  )
}
