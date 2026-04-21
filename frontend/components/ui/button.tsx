import { cn } from '@/lib/cn'

type Props = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'danger' | 'success'
  size?: 'sm' | 'md'
}

export function Button({ className, variant = 'primary', size = 'md', ...props }: Props) {
  const base =
    'inline-flex items-center justify-center rounded-xl border text-sm font-medium transition disabled:opacity-50 disabled:cursor-not-allowed'

  const variants: Record<string, string> = {
    primary: 'bg-blue-600 border-blue-600 text-white hover:bg-blue-700 dark:border-blue-500 dark:bg-blue-500 dark:hover:bg-blue-400',
    secondary: 'bg-white border-gray-200 text-gray-900 hover:bg-gray-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:hover:bg-slate-800',
    success: 'bg-green-600 border-green-600 text-white hover:bg-green-700 dark:border-green-500 dark:bg-green-500 dark:hover:bg-green-400',
    danger: 'bg-red-600 border-red-600 text-white hover:bg-red-700 dark:border-red-500 dark:bg-red-500 dark:hover:bg-red-400',
  }

  const sizes: Record<string, string> = {
    sm: 'h-9 px-3',
    md: 'h-10 px-4',
  }

  return <button className={cn(base, variants[variant], sizes[size], className)} {...props} />
}
