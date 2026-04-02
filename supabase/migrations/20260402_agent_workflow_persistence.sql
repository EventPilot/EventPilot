create or replace function public.set_updated_at()
returns trigger
language plpgsql
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

alter table public.event
  add column if not exists context jsonb not null default '{}'::jsonb,
  add column if not exists context_updated_at timestamptz null;

create table if not exists public.agent_run (
  id uuid primary key,
  chat_id uuid not null references public.chat(id) on delete cascade,
  event_id uuid not null references public.event(id) on delete cascade,
  requested_by_user_id uuid not null references public."user"(id),
  status text not null,
  plan_summary text not null default '',
  blocked_on_chat_id uuid null references public.chat(id) on delete set null,
  current_task_index integer not null default 0,
  context_snapshot jsonb not null default '{}'::jsonb,
  planner_response jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  started_at timestamptz null,
  completed_at timestamptz null,
  failed_at timestamptz null,
  constraint agent_run_status_check check (status in ('planning', 'awaiting_approval', 'running', 'waiting_on_member', 'completed', 'failed', 'cancelled'))
);

create unique index if not exists agent_run_one_active_per_chat_idx
  on public.agent_run (chat_id)
  where status in ('planning', 'awaiting_approval', 'running', 'waiting_on_member');

create index if not exists agent_run_event_created_idx
  on public.agent_run (event_id, created_at desc);

create index if not exists agent_run_chat_created_idx
  on public.agent_run (chat_id, created_at desc);

create index if not exists agent_run_active_status_updated_idx
  on public.agent_run (status, updated_at desc)
  where status in ('planning', 'awaiting_approval', 'running', 'waiting_on_member');

create table if not exists public.agent_task (
  id uuid primary key,
  run_id uuid not null references public.agent_run(id) on delete cascade,
  position integer not null,
  title text not null,
  kind text not null,
  status text not null,
  target_user_id uuid null references public."user"(id),
  target_chat_id uuid null references public.chat(id) on delete set null,
  instructions text not null default '',
  completion_signal text not null default '',
  result text not null default '',
  task_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  started_at timestamptz null,
  completed_at timestamptz null,
  failed_at timestamptz null,
  constraint agent_task_run_position_unique unique (run_id, position),
  constraint agent_task_kind_check check (kind in ('internal', 'ask_member', 'generate_post', 'wait_for_member', 'publish_post')),
  constraint agent_task_status_check check (status in ('pending', 'in_progress', 'waiting', 'completed', 'failed', 'skipped'))
);

create index if not exists agent_task_run_position_idx
  on public.agent_task (run_id, position);

create index if not exists agent_task_target_chat_idx
  on public.agent_task (target_chat_id);

create index if not exists agent_task_status_updated_idx
  on public.agent_task (status, updated_at desc);

alter table public.chat_message
  add column if not exists agent_run_id uuid null references public.agent_run(id) on delete set null,
  add column if not exists agent_task_id uuid null references public.agent_task(id) on delete set null,
  add column if not exists message_type text not null default 'message',
  add column if not exists metadata jsonb not null default '{}'::jsonb;

alter table public.chat_message
  drop constraint if exists chat_message_type_check;

alter table public.chat_message
  add constraint chat_message_type_check check (message_type in ('message', 'plan_proposal', 'approval_request', 'task_update', 'member_request', 'system'));

create index if not exists chat_message_chat_created_idx
  on public.chat_message (chat_id, created_at);

create index if not exists chat_message_agent_run_created_idx
  on public.chat_message (agent_run_id, created_at);

create index if not exists chat_message_agent_task_created_idx
  on public.chat_message (agent_task_id, created_at);

drop trigger if exists set_agent_run_updated_at on public.agent_run;
create trigger set_agent_run_updated_at
before update on public.agent_run
for each row execute function public.set_updated_at();

drop trigger if exists set_agent_task_updated_at on public.agent_task;
create trigger set_agent_task_updated_at
before update on public.agent_task
for each row execute function public.set_updated_at();
