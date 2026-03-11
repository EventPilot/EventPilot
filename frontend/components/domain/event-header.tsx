"use client";

import { Card } from "@/components/ui/card";
import { StatusPill } from "@/components/ui/status-pill";
import { EventActions } from "./event-actions";
import { formatDateTime } from "../helpers";

interface EventData {
  id: string;
  title: string;
  description: string;
  event_date: string;
  location: string;
  status: string;
  created_at: string;
}

export function EventHeader({
  event,
  role,
  rightSlot,
}: {
  event: EventData;
  role: string;
  rightSlot?: React.ReactNode;
}) {
  return (
    <Card className="p-6">
      <div className="flex items-start justify-between gap-6">
        <div>
          <div className="text-lg font-semibold">{event.title}</div>
          <div className="text-sm text-gray-500 mt-1">
            {formatDateTime(event.event_date)}{" "}
            {event.location && `• ${event.location}`}
          </div>
          <div className="mt-3 border-l-2 border-gray-200 pl-3 text-sm text-gray-600">
            {event.description}
          </div>
          <div className="inline-flex mt-3 gap-2">
            <span className="items-center rounded-full border border-gray-200 bg-gray-100 px-3 py-1 text-xs text-gray-600">
              {role}
            </span>
            <StatusPill status={event.status} />
          </div>
        </div>
        {role === "Owner" &&
          (rightSlot ?? (
            <EventActions eventId={event.id} status={event.status} />
          ))}
      </div>
    </Card>
  );
}
