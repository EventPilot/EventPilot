import { Topbar } from '@/components/shell/topbar'

export function AppShell({
  title,
  userName,
  userSubline,
  showCreateAction = true,
  children,
}: {
  title: string
  userName?: string
  userSubline?: string
  showCreateAction?: boolean
  children: React.ReactNode
}) {
  return (
    <div className="flex h-screen overflow-hidden bg-[#f7f8fa] dark:bg-slate-950">
      <main className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <Topbar
          title={title}
          userName={userName}
          userSubline={userSubline}
          showCreateAction={showCreateAction}
        />
        <div className="flex-1 overflow-auto p-6">{children}</div>
      </main>
    </div>
  )
}
