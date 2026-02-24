import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import type { RoleInput } from '@/lib/data/events'

export function RoleInputs({ inputs }: { inputs: RoleInput[] }) {
  return (
    <Card className="p-6">
      <div className="text-lg font-semibold">Inputs by role</div>

      <div className="mt-5 space-y-4">
        {inputs.map((r) => {
          const pct = r.totalCount === 0 ? 0 : Math.round((r.completeCount / r.totalCount) * 100)
          const barClass = pct === 100 ? 'bg-green-600' : 'bg-amber-500'

          return (
            <div key={r.role} className="rounded-2xl border border-gray-200 bg-gray-50 p-4">
              <div className="text-sm font-medium">{r.role}</div>
              <div className="text-xs text-gray-500 mt-1">
                {r.completeCount}/{r.totalCount} complete
              </div>
              <div className="mt-3 h-2 rounded-full bg-gray-200 overflow-hidden">
                <div className={`h-full ${barClass}`} style={{ width: `${Math.max(10, pct)}%` }} />
              </div>
              <div className="mt-3 text-xs text-gray-500">{r.hint}</div>
              <div className="mt-4 flex items-center gap-3">
                <Button variant={r.role === 'Photographer' ? 'primary' : 'secondary'} size="sm">
                  Complete
                </Button>
                <Button variant="secondary" size="sm">
                  Request edit
                </Button>
              </div>
            </div>
          )
        })}
      </div>
    </Card>
  )
}
