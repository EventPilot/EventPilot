export function Topbar({ title }: { title: string }) {
  return (
    <div className="h-[72px] border-b border-gray-200 bg-white">
      <div className="flex h-full items-center justify-between px-6">
        <div className="text-lg font-semibold">{title}</div>

        <div className="flex items-center gap-3">
          <div className="h-9 w-[340px] rounded-xl border border-gray-200 bg-gray-50 px-4 text-xs text-gray-500 flex items-center">
            Search events, drafts…
          </div>
          <div className="h-9 w-9 rounded-xl border border-gray-200 bg-gray-50 flex items-center justify-center text-sm">
            🔔
          </div>
          <div className="h-9 rounded-full border border-gray-200 bg-gray-50 px-4 text-xs text-gray-600 flex items-center">
            Me
          </div>
        </div>
      </div>
    </div>
  )
}
