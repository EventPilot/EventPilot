import Link from "next/link";
import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { AppShell } from "@/components/shell/app-shell";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { EventCard } from "@/components/domain/event-card";
import { TaskCard } from "@/components/domain/task-card";

import { formatDateTime } from "@/components/helpers";
import { listUpcomingEvents } from "@/lib/data/upcoming-events";

export default async function DashboardHomePage() {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    redirect("/login");
  }

  const { data: profile } = await supabase
    .from("user")
    .select("name")
    .eq("id", user.id)
    .single();

  const name = profile?.name ?? user.email?.split("@")[0] ?? "Account";

  const events = await listUpcomingEvents();

  return (
    <AppShell title="Home" userName={name} userSubline="Owner • Workspace A">
      <div className="grid grid-cols-12 gap-6">
        <div className="col-span-8 space-y-6">
          <Card className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-lg font-semibold">Upcoming Events</div>
                <div className="text-sm text-gray-500 mt-1">
                  All scheduled milestones and drafts.
                </div>
              </div>
              <Link href="/dashboard/events/new">
                <Button>+ Add event</Button>
              </Link>
            </div>

            <div className="mt-5 grid grid-cols-2 gap-4">
              {events.map((e: any) => (
                <EventCard
                  key={e.id}
                  title={e.title}
                  subtitle={
                    e.location
                      ? `${formatDateTime(e.event_date)} • ${e.location}`
                      : `${formatDateTime(e.event_date)}`
                  }
                  status={e.status}
                  role={e.role}
                  href={`/dashboard/events/${e.id}`}
                />
              ))}
            </div>
          </Card>
        </div>

        <div className="col-span-4 space-y-6">
          <Card className="p-6">
            <div className="text-lg font-semibold">Action items</div>
            <div className="text-sm text-gray-500 mt-1">
              Things that block a post from being generated.
            </div>

            <div className="mt-5 space-y-4">
              <TaskCard
                title="Photographer upload pending"
                meta="Customer Rocket Launch — Artemis Demo"
              />
              <TaskCard
                title="Customer quote needed"
                meta="Customer Rocket Launch — Artemis Demo"
              />
              <TaskCard
                title="Review draft for Press Kit"
                meta="Press Kit Review"
              />
            </div>
          </Card>

          <Card className="p-6">
            <div className="text-lg font-semibold">Quick links</div>
            <div className="mt-4 flex flex-col gap-3">
              <Link
                href="/dashboard/events/new"
                className="text-sm text-blue-700 hover:underline"
              >
                Create a new event
              </Link>
              <Link
                href="/dashboard/drafts"
                className="text-sm text-blue-700 hover:underline"
              >
                Review drafts
              </Link>
              <Link
                href="/dashboard/settings"
                className="text-sm text-blue-700 hover:underline"
              >
                Workspace settings
              </Link>
            </div>
          </Card>
        </div>
      </div>
    </AppShell>
  );
}
