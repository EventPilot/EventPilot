import { cn } from '@/lib/cn'

const DOT: Record<string, string> = {
  Scheduled: 'bg-sky-500',
  'Awaiting inputs': 'bg-amber-500',
  'Inputs collected': 'bg-purple-500',
  'Draft ready': 'bg-green-600',
  Published: 'bg-green-600',
}

export function StatusPill({ status, className }: { status: string; className?: string }) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-2 rounded-full border border-gray-200 bg-gray-100 px-3 py-1 text-xs text-gray-600',
        className
      )}
    >
      <span className={cn('h-2 w-2 rounded-full', DOT[status])} />
      {status}
    </span>
  )
}
