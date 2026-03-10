"use server";

import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";

export async function deleteEventAction(eventId: string) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) redirect("/login");

  // Verify the user is the owner
  const { data: membership } = await supabase
    .from("event_member")
    .select("role")
    .eq("event_id", eventId)
    .eq("user_id", user.id)
    .single();

  if (membership?.role !== "Owner") {
    throw new Error("Only the event owner can delete this event");
  }

  // Delete the event
  const { error } = await supabase.from("event").delete().eq("id", eventId);
  if (error) throw new Error(error.message);

  redirect("/dashboard/events");
}

export async function markEventFinishedAction(eventId: string) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) redirect("/login");

  const { data: membership } = await supabase
    .from("event_member")
    .select("role")
    .eq("event_id", eventId)
    .eq("user_id", user.id)
    .single();

  if (membership?.role !== "Owner") {
    throw new Error("Only the event owner can update this event");
  }

  const { error } = await supabase
    .from("event")
    .update({ status: "Awaiting inputs" })
    .eq("id", eventId);

  if (error) throw new Error(error.message);

  redirect(`/dashboard/events/${eventId}`);
}

export async function updateEventAction(eventId: string, formData: FormData) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) redirect("/login");

  const title = formData.get("title") as string;
  const description = formData.get("description") as string;
  const event_date = formData.get("event_date") as string;
  const location = (formData.get("location") as string) ?? "";

  const { error } = await supabase
    .from("event")
    .update({ title, description, event_date, location })
    .eq("id", eventId);

  if (error) throw new Error(error.message);

  redirect(`/dashboard/events/${eventId}`);
}
