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
    <div className="flex h-screen overflow-hidden">
      <Sidebar userName={userName} subline={userSubline} />
      <main className="flex flex-col flex-1 overflow-hidden">
        <Topbar title={title} />
        <div className="flex-1 overflow-auto p-6">{children}</div>
      </main>
    </div>
  )
}
