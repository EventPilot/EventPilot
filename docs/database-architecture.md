# Database Design Document

This document describes the relational database schema for EventPilot using Supabase (PostgreSQL). The database stores user information, event milestones, chat-based agent interactions, media assets, and generated posts.

## Sql vs Nosql

|                | Sql (relational)        | Nosql (non-relational)        |
| -------------- | ----------------------- | ----------------------------- |
| Format         | tables (rows and cols)  | collections                   |
| Schema         | structured and explicit | flexible but implicit         |
| Relationships  | foreign keys & joins    | manual references             |
| Data integrity | enforced by database    | enforced by app code          |
| Best for       | clear relationships     | unstructured or changing data |

## Why we chose Supabase?

Sql is a better fit for EventPilot because our core data model is relational. Users, events, chats, and posts have explicit relationships, and Supabase uses PostgreSql, which ensures data integrity and enables simple queries. On the other hand, Nosql approach would shift these guarantees into application code and thus increase overall design complexity. Even for the chats table, Nosql json may be at a disadvantage when it comes to pagination, ordering, and filtering data. However, sql can simply achieve through query like this:
`SELECT * FROM chat_messages WHERE chat_id = 1 ORDER BY created_at ASC;`

## Tables

### users

| column | type | notes                         |
| ------ | ---- | ----------------------------- |
| id     | uuid | primary key                   |
| name   | text | user name                     |
| role   | text | owner, photographer, customer |

### events

| column      | type | notes             |
| ----------- | ---- | ----------------- |
| id          | uuid | primary key       |
| title       | text | event title       |
| description | text | event discription |
| event_date  | date | event date        |

### event_owners (many-to-many)

- allows multiple owners/members per event

| column   | type | notes       |
| -------- | ---- | ----------- |
| user_id  | uuid | foreign key |
| event_id | uuid | foreign key |

### chats

- each event has one chat

| column     | type        | notes         |
| ---------- | ----------- | ------------- |
| id         | uuid        | primary key   |
| event_id   | uuid        | foreign key   |
| created_at | timestamptz | default now() |

### chat_messages

- stores the actual conversation between users and chat agent
- all messages for all events are in this table

| column      | type        | notes              |
| ----------- | ----------- | ------------------ |
| id          | uuid        | primary key        |
| chat_id     | uuid        | foreign key        |
| sender_type | text        | user/agent         |
| sender_id   | uuid        | null if it's agent |
| message     | text        | message            |
| created_at  | timestamptz | default now()      |

### posts

| column     | type        | notes                            |
| ---------- | ----------- | -------------------------------- |
| id         | uuid        | primary key                      |
| event_id   | uuid        | foreign key                      |
| content    | text        | generated post text              |
| status     | text        | posted, failed                   |
| url        | text        | link to live post (verification) |
| created_at | timestamptz | default now()                    |

### media

- tracks uploaded images/videos stored in Supabase storage

| column       | type        | notes                  |
| ------------ | ----------- | ---------------------- |
| id           | uuid        | primary key            |
| event_id     | uuid        | foreign key            |
| uploaded_by  | uuid        | foreign key (users.id) |
| media_type   | text        | image, video           |
| storage_path | text        | path/filename          |
| metadata     | jsonb       | media properties       |
| created_at   | timestamptz | default now()          |

## Blob storage (Supabase storage)

- stores actual media files
- database only stores file path since it's bad at large binaries
