export type EventStatus =
  | "Scheduled"
  | "Awaiting inputs"
  | "Draft ready"
  | "Published";

export type EventRole = "Owner" | "Photographer" | "Customer/Partner";

export type TimelineItem = {
  title: string;
  meta: string;
};

export type RoleInput = {
  role: EventRole;
  completeCount: number;
  totalCount: number;
  hint: string;
};

export type Draft = {
  text: string;
  hashtags: string;
  media: Array<{ id: string; label: string }>;
};

export type Event = {
  id: string;
  title: string;
  start: string;
  end: string;
  timezone: string;
  location: string;
  client?: string;
  status: EventStatus;
  description?: string;
  tags?: string[];
  timeline: TimelineItem[];
  roleInputs: RoleInput[];
  draft: Draft;
};

const EVENTS: Event[] = [
  {
    id: "evt_001",
    title: "Customer Rocket Launch — Artemis Demo",
    start: "Feb 18, 2026 2:00 PM",
    end: "Feb 18, 2026 4:00 PM",
    timezone: "ET",
    location: "Wallops Flight Facility",
    client: "Artemis Aerospace",
    status: "Awaiting inputs",
    description: "Customer-facing demo launch milestone.",
    tags: ["Launch", "Milestone", "Customer"],
    timeline: [
      { title: "Event created", meta: "Feb 2 • 10:11 AM" },
      { title: "Prompts scheduled", meta: "After event end time" },
      { title: "Owner input", meta: "Complete" },
      { title: "Photographer upload", meta: "Pending" },
      { title: "Customer quote", meta: "Pending" },
      { title: "Draft generated", meta: "—" },
      { title: "Published", meta: "—" },
    ],
    roleInputs: [
      {
        role: "Owner",
        completeCount: 3,
        totalCount: 3,
        hint: "What happened? • Key result • Who to credit",
      },
      {
        role: "Photographer",
        completeCount: 0,
        totalCount: 2,
        hint: "Upload media • Add captions",
      },
      {
        role: "Customer/Partner",
        completeCount: 0,
        totalCount: 1,
        hint: "Provide quote • Approve usage",
      },
    ],
    draft: {
      text: [
        "Successful customer launch demo completed today.",
        "Key highlight: stable ascent + clean separation.",
        "Huge thanks to the team and partners at Artemis Aerospace.",
      ].join("\n"),
      hashtags: "#launch #aerospace @ArtemisAero",
      media: [
        { id: "m1", label: "IMG" },
        { id: "m2", label: "IMG" },
        { id: "m3", label: "IMG" },
      ],
    },
  },
  {
    id: "evt_002",
    title: "Payload Integration",
    start: "Feb 10, 2026 11:00 AM",
    end: "Feb 10, 2026 12:00 PM",
    timezone: "ET",
    location: "Assembly Bay",
    status: "Scheduled",
    timeline: [],
    roleInputs: [],
    draft: { text: "", hashtags: "", media: [] },
  },
  {
    id: "evt_003",
    title: "Engine Static Fire",
    start: "Feb 14, 2026 9:30 AM",
    end: "Feb 14, 2026 10:30 AM",
    timezone: "ET",
    location: "Test Stand 2",
    status: "Scheduled",
    timeline: [],
    roleInputs: [],
    draft: { text: "", hashtags: "", media: [] },
  },
  {
    id: "evt_005",
    title: "Press Kit Review",
    start: "Feb 25, 2026 4:00 PM",
    end: "Feb 25, 2026 5:00 PM",
    timezone: "ET",
    location: "Conference Room",
    status: "Draft ready",
    timeline: [],
    roleInputs: [],
    draft: { text: "", hashtags: "", media: [] },
  },
];

export type EventSummary = Pick<
  Event,
  "id" | "title" | "start" | "location" | "status"
>;

export async function listEvents(): Promise<EventSummary[]> {
  return EVENTS.map(({ id, title, start, location, status }) => ({
    id,
    title,
    start,
    location,
    status,
  }));
}

export async function getNextEvent(): Promise<Event | null> {
  // For now, just use the first event as the “current” one.
  return EVENTS[0] ?? null;
}

export async function getEventById(id: string): Promise<Event | null> {
  return EVENTS.find((e) => e.id === id) ?? null;
}
