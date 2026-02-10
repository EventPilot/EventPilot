import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

export function TaskCard({ title, meta }: { title: string; meta: string }) {
  return (
    <Card className="bg-gray-50 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-sm font-medium">{title}</div>
          <div className="text-xs text-gray-500 mt-1">{meta}</div>
        </div>
        <Button variant="secondary" size="sm">
          Open
        </Button>
      </div>
    </Card>
  )
}
