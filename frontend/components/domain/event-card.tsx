"use client";

import Link from "next/link";
import { Card } from "@/components/ui/card";
import { StatusPill } from "@/components/ui/status-pill";
import { Button } from "@/components/ui/button";
import { formatDateTime } from "../helpers";

export function EventCard({
  title,
  eventDate,
  location,
  status,
  role,
  href,
}: {
  title: string;
  eventDate: string;
  location?: string;
  status: string;
  role: string;
  href: string;
}) {
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="text-sm font-medium">{title}</div>
          <div className="text-xs text-gray-500 mt-1">
            {location
              ? `${formatDateTime(eventDate)} • ${location}`
              : formatDateTime(eventDate)}
          </div>
          <div className="inline-flex mt-3 gap-2">
            <span className="items-center rounded-full border border-gray-200 bg-gray-100 px-3 py-1 text-xs text-gray-600">
              {role}
            </span>
            <StatusPill status={status} />
          </div>
        </div>

        <Link href={href}>
          <Button variant="secondary" size="sm">
            View
          </Button>
        </Link>
      </div>
    </Card>
  );
}
