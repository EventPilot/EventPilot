"use server";

import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";
import { randomUUID } from "crypto";

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
  const location = (formData.get("location") as string) ?? "";

  // Insert the event
  const { error: eventError } = await supabase
    .from("event")
    .insert({ id: eventId, title, description, event_date, location });

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

  redirect(`/dashboard/events/${eventId}`);
}
