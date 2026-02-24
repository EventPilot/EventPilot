# EventPilot

EventPilot is a calendar and marketing agent for small engineering startups and student orgs. It schedules major milestones/events, prompts event stakeholders (owner, photographer, customers) for information and media after the event, uses an LLM to generate social posts, and publishes to X (Twitter) first, with LinkedIn and Instagram planned later.

## Architecture

- **Backend**: Go API running serverless on Vercel
- **Frontend**: Next.js with App Router
- **Database**: Supabase (PostgreSQL)
- **LLM**: Claude (Anthropic) with OpenAI fallback
- **Social**: X API v2 (Twitter)

See [docs/backend-architecture.md](docs/backend-architecture.md) and [docs/database-architecture.md](docs/database-architecture.md) for detailed architecture documentation.

## Project Structure

```
EventPilot/
├── api/                 # Go backend API
│   ├── handlers/        # HTTP handlers for endpoints
│   ├── middleware/      # HTTP middleware (auth, CORS, logging)
│   ├── models/          # Data models
│   ├── db/              # Database connection
│   └── main.go          # Entry point
├── frontend/            # Next.js frontend
│   ├── app/             # Next.js App Router
│   ├── components/      # React components
│   └── lib/             # Utilities and API clients
└── docs/                # Documentation
```

## Development

### Backend (Go API)

```bash
cd api
go mod download
set -a && source ../.env && set +a && go run main.go
```

### Frontend (Next.js)

```bash
cd frontend
npm install
npm run dev
```

## Environment Variables

See `.env.example` files in respective directories for required environment variables.

## License

MIT
