import { createClient } from "../supabase/server";

export async function listUpcomingEvents() {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) return [];
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
        location,
        status
      )
    `,
    )
    .eq("user_id", user.id)
    .order("event_date", { referencedTable: "event", ascending: true })
    .limit(8); // fetch more to account for past events being filtered out

  const now = new Date().toISOString();

  return (
    memberRows
      ?.map((row: any) => ({ ...row.event, role: row.role }))
      .filter((e: any) => e.event_date >= now)
      .sort((a: any, b: any) => a.event_date.localeCompare(b.event_date))
      .slice(0, 4) ?? []
  );
}
