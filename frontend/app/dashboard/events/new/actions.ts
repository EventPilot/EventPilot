"use server";

import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";
import { randomUUID } from "crypto";

type RoleAssignment = {
  email: string;
  role: string;
};

export async function createEvent(formData: FormData) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) redirect("/login");

  const eventId = randomUUID();
  const title = formData.get("title") as string;
  const description = formData.get("description") as string;
  const event_date = formData.get("event_date") as string;
  const event_date_iso = new Date(event_date).toISOString(); // convert event_date to supabase time format
  const location = (formData.get("location") as string) ?? "";
  const tone = (formData.get("tone") as string) ?? "Professional";
  const rolesRaw = formData.get("roles") as string | null;
  const roleAssignments: RoleAssignment[] = rolesRaw
    ? JSON.parse(rolesRaw)
    : [];

  // Insert the event
  const { error: eventError } = await supabase.from("event").insert({
    id: eventId,
    title,
    description,
    event_date: event_date_iso,
    location,
    tone,
  });

  if (eventError) throw new Error(eventError?.message);

  // Add creator as owner
  const { error: memberError } = await supabase.from("event_member").insert({
    event_id: eventId,
    user_id: user.id,
    role: "Owner",
  });

  if (memberError) {
    // Cleanup
    await supabase.from("event").delete().eq("id", eventId);
    throw new Error(memberError.message);
  }

  // Add additional role members by email lookup
  if (roleAssignments.length > 0) {
    const emails = roleAssignments.map((r) => r.email);

    const { data: matchedUsers, error: lookupError } = await supabase
      .from("user")
      .select("id, email")
      .in("email", emails);

    if (!lookupError && matchedUsers && matchedUsers.length > 0) {
      const emailToId = Object.fromEntries(
        matchedUsers.map((u) => [u.email, u.id]),
      );
      const memberRows = roleAssignments
        .filter((r) => emailToId[r.email])
        .map((r) => ({
          event_id: eventId,
          user_id: emailToId[r.email],
          role: r.role,
        }));
      if (memberRows.length > 0) {
        const { error: rolesError } = await supabase
          .from("event_member")
          .insert(memberRows);
        if (rolesError)
          console.error("Failed to insert role members:", rolesError.message);
      }
    }
  }

  redirect(`/dashboard/events/${eventId}`);
}
