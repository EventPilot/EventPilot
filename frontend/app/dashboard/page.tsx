import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { AppShell } from "@/components/shell/app-shell";
import { Card } from "@/components/ui/card";
import { EventCard } from "@/components/domain/event-card";

export default async function DashboardHomePage() {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    redirect("/login");
  }

  const { data: profile } = await supabase.from('user').select('name').eq('id', user.id).single()
  const { data: memberRows } = await supabase
    .from("event_member")
    .select(
      `
      role,
      event (
        id,
        title,
        description,
        event_date,
        created_at,
        location,
        status
      )
    `,
    )
    .eq("user_id", user.id);

  const events =
    memberRows
      ?.map((row: any) => ({ ...row.event, role: row.role }))
      .sort((a: any, b: any) => a.event_date.localeCompare(b.event_date)) ?? [];

  return (
    <AppShell title="Events" userName={profile?.name ?? user.email?.split('@')[0] ?? 'Account'} userSubline={user.email ?? ''}>
      <Card className="p-6">
        <div className="flex items-center justify-between gap-4">
          <div>
            <div className="text-lg font-semibold">All events</div>
            <div className="mt-1 text-sm text-gray-500 dark:text-slate-400">
              {events.length} event{events.length !== 1 ? "s" : ""}
            </div>
          </div>
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {events.map((e: any) => (
            <EventCard
              key={e.id}
              title={e.title}
              eventDate={e.event_date}
              location={e.location}
              status={e.status}
              role={e.role}
              href={`/dashboard/events/${e.id}`}
            />
          ))}
        </div>
      </Card>
    </AppShell>
  );
}
