import { Sidebar } from '@/components/shell/sidebar'
import { Topbar } from '@/components/shell/topbar'

export function AppShell({
  title,
  userName,
  userSubline,
  children,
}: {
  title: string
  userName?: string
  userSubline?: string
  children: React.ReactNode
}) {
  return (
    <div className="flex min-h-screen">
      <Sidebar userName={userName} subline={userSubline} />
      <main className="flex-1">
        <Topbar title={title} />
        <div className="p-6">{children}</div>
      </main>
    </div>
  )
}
