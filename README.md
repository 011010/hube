# hube

Personal self-hosted hub — notes, tasks, finance, projects, and an AI agent, all in one place.

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.26, Chi, SQLite |
| Frontend | React 19, TypeScript, Vite, Tailwind |
| Desktop | Tauri 2.0 (WIP) |
| Infra | Docker, GitHub Actions, ghcr.io |

Architecture follows **Hexagonal/Clean**: `domain → application → infrastructure`. The HTTP layer and external clients live in `infrastructure`; business logic lives in `application`; models and repository interfaces live in `domain`.

## Project structure

```
hube/
├── backend/
│   ├── cmd/
│   │   ├── api/        # HTTP server entrypoint
│   │   └── cli/        # hube CLI
│   └── internal/
│       ├── domain/     # models + repository interfaces
│       ├── application/# use cases / services
│       └── infrastructure/
│           ├── http/   # Chi router + handlers
│           ├── sqlite/ # SQLite repositories
│           └── external/ # MoneyMonkey, PayPinga, Claude, OpenAI clients
├── frontend/           # React SPA (Vite + Tailwind)
├── desktop/            # Tauri 2.0 desktop app (WIP)
├── docker-compose.yml
└── Makefile
```

## Running locally

### Prerequisites

- Go 1.22+
- Node.js 22 + pnpm
- (Optional) Docker

### Dev mode

```bash
make dev          # starts API on :8090 + frontend on :5173 in parallel
make dev-api      # API only
make dev-ui       # frontend only
```

### Docker

```bash
cp .env.example .env   # fill in secrets
make docker-up         # starts api + web containers
make docker-down
```

The web container exposes port `80`. The API is internal (only accessible by the web container via Docker network).

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8090` | API listen port |
| `DB_PATH` | `./hube.db` | SQLite database path |
| `MONKEYAPI_URL` | — | MoneyMonkey base URL (finance) |
| `MONKEYAPI_KEY` | — | MoneyMonkey API key |
| `PAYPINGA_URL` | — | PayPinga base URL (card tracker) |
| `PAYPINGA_KEY` | — | PayPinga API key |
| `CLAUDE_API_KEY` | — | Anthropic API key (AI Playground) |
| `OPENAI_BASE_URL` | — | OpenAI-compatible base URL (Ollama, OpenRouter, OpenCode) |
| `OPENAI_API_KEY` | — | API key for the OpenAI-compatible provider |

## CLI

Install once:

```bash
make install-cli
```

This runs `go install ./cmd/cli` and places the `cli` binary in your `$GOPATH/bin`. Rename it to `hube` or add an alias:

```bash
alias hube='HUBE_URL=http://your-server:8090 cli'
```

### Commands

```bash
hube tasks                   # all tasks
hube tasks --status todo     # filter by status (todo | in_progress | done)
hube notes                   # list notes with preview
hube transactions            # finance summary + recent transactions
```

`HUBE_URL` defaults to `http://localhost:8090`. Point it at your Pi or VPS:

```bash
HUBE_URL=http://tu-pi:8090 hube tasks --status todo
```

## API reference

Base path: `/api/v1`

| Method | Path | Description |
|---|---|---|
| `GET` | `/tasks` | List tasks (optional `?status=`) |
| `POST` | `/tasks` | Create task |
| `PUT` | `/tasks/{id}` | Update task |
| `DELETE` | `/tasks/{id}` | Delete task |
| `GET` | `/notes` | List notes |
| `POST` | `/notes` | Create note |
| `PUT` | `/notes/{id}` | Update note |
| `DELETE` | `/notes/{id}` | Delete note |
| `GET` | `/folders` | List folders |
| `GET` | `/events` | List calendar events |
| `GET` | `/projects` | List projects |
| `GET` | `/apps` | List launcher apps |
| `GET` | `/settings` | Get all settings |
| `PUT` | `/settings` | Update settings |
| `GET` | `/finance/summary` | Balance + recent transactions |
| `GET` | `/cards/summary` | Card tracker summary |
| `POST` | `/ai/chat` | AI Playground — send message to agent |
| `GET` | `/health` | Health check |

## CI/CD

`.github/workflows/ci.yml` runs on every push to `main`:

1. **Backend** — `go build` + `go vet`
2. **Frontend** — `pnpm install --frozen-lockfile` + `pnpm run build`
3. **Publish images** — builds Docker images and pushes to `ghcr.io/011010/hube-api:latest` and `ghcr.io/011010/hube-web:latest`

Required secrets: `GITHUB_TOKEN` (automatic).

## Deploy

Pull and run the published images on any server:

```bash
docker pull ghcr.io/011010/hube-api:latest
docker pull ghcr.io/011010/hube-web:latest
```

Or use `docker-compose.yml` pointing to the ghcr.io images directly (WIP — deploy automation is a pending task).
