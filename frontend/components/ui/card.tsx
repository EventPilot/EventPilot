import { cn } from '@/lib/cn'

export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('rounded-2xl border border-gray-200 bg-white dark:border-slate-800 dark:bg-slate-950', className)} {...props} />
}
