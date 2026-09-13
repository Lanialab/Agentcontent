# AgentContent

Personal creator operating system: research YouTube channels, outlier feed, AI ideas, Brand Blueprint (Ikigai / positioning / pillars), and video scripts (Nhanh / Auto / Sâu).

Business logic is authoritative. Code and adapters cannot override domain rules. SQLite is the sole source of truth. Google Sheets is optional sync/export only — **no Apps Script**.

## Run

Requires Go 1.22+ and Node 20+.

```bash
# 1. Frontend
cd web && npm install && npm run build && cd ..

# 2. API + UI (serves web/dist)
go run ./cmd/server
```

Open http://localhost:8080 — that address only works on this machine. To open AgentContent on a phone, deploy it (see **Deploy for phone**).

Optional split-process dev:

```bash
go run ./cmd/server          # :8080
cd web && npm install && npm run dev   # :5173 proxies /api
```

Scheduled scan (CLI / job entry — wire this to launchd, Task Scheduler, or cron every 60 minutes):

```bash
go run ./cmd/scan
```

```
0 * * * * cd /path/to/AgentContent && go run ./cmd/scan
```

Overlapping ticks are skipped with a SQLite lease (`jobs` row `scan`).

## Tests

```bash
go test ./...
```

Covers outlier scoring (`viewsPerDay` vs same-channel average), the scan → feed use case, and `GET /api/health`.

## Environment

Copy `.env.example`. Important variables:

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_PATH` | `./data/agentcontent.db` | SQLite path (created + migrated on boot). Use `/data/agentcontent.db` on Railway/Fly. |
| `OUTLIER_MULTIPLIER` | `2.5` | Outlier threshold (use 2–3) |
| `SEED_ON_EMPTY` | `true` | Load demo channels / videos / blueprint / scripts |
| `YOUTUBE_API_KEY` | empty | Empty → YouTube **stub** |
| `OPENAI_API_KEY` | empty | Empty → AI **stub** |
| `OPENAI_BASE_URL` | OpenAI | Any OpenAI-compatible endpoint |
| `SHEETS_ENABLED` | `false` | Optional sync adapter (no-op unless implemented later) |

No API keys are required to demo. Seed data is enough.

## Deploy for phone

The Go process serves the built Vite UI and `/api` together. Deploy that one container so an iPhone can open a **public HTTPS URL** (not `localhost`).

Stubs stay on unless you add keys. Leave `YOUTUBE_API_KEY` and `OPENAI_API_KEY` empty for the demo.

Health: `GET /api/health` → `{"ok":true,"name":"AgentContent"}`.

### Railway

1. Push this repo to GitHub (this branch or `main` after merge).
2. At [railway.app](https://railway.app) → **New Project** → **Deploy from GitHub repo** → select AgentContent.
3. Railway uses `Dockerfile` + `railway.json` (builder `DOCKERFILE`, healthcheck `/api/health`).
4. **Variables** → paste from `.env.example`, then set:
   - `DATABASE_PATH=/data/agentcontent.db`
   - `SEED_ON_EMPTY=true`
   - `UI_DIR=/app/web/dist` (optional; already the image default)
   - Leave `PORT` unset so Railway injects it
   - Leave YouTube / OpenAI keys empty (stubs)
5. **Settings → Volumes → Add Volume** → mount path **`/data`** (SQLite is not stored in the image; without a volume the DB resets on every deploy).
6. **Settings → Networking → Generate Domain** — this is the public HTTPS URL.
7. On an iPhone, open that domain in Safari. You should see the Architecture Live View and be able to browse Feed / Creators. Confirm health at `https://<your-domain>/api/health`.

`railway.json` cannot attach the volume; the dashboard (or `railway volume add --mount-path /data`) is required once per service.

### Fly.io (optional)

`fly.toml` builds the same Dockerfile, checks `/api/health`, and mounts volume `agentcontent_data` at `/data`.

```bash
fly launch --no-deploy   # accept/edit the app name in fly.toml
fly volumes create agentcontent_data --size 1
fly deploy
fly apps open
```

Open the `https://<app>.fly.dev` URL on the phone.

### Local image check

```bash
docker build -t agentcontent .
docker run --rm -p 8080:8080 -e PORT=8080 -v agentcontent-data:/data agentcontent
```

Then visit http://localhost:8080 — still local only; use Railway/Fly for the phone.

## Architecture map

The in-app **Kiến trúc / Architecture Live View** is a hexagonal layered map (not a node-flow graph). Original reference: `docs/architecture-original.jpg`.

- **Left** (outside Go): Personal UI Client — Creators & Groups, Outlier Feed, AI Assistant & Ideas, Brand Blueprint, Kịch bản Video.
- **Center**: one hex — GO APPLICATION BACKEND / HEXAGONAL MODULAR MONOLITH — concentric layers Driving adapters → Inbound ports → Use cases → Domain → Outbound ports → Driven adapters.
- **Right**: YouTube Data API, Transcript provider, AI providers, Google Sheets API, local files, native OS scheduler.

On a phone the three columns stack and the hex scales to the viewport; pinch or use **+ / −**. Tạm dừng / Chạy lại / Hiện tất cả remain. Scan, generate idea, and save blueprint pulse the matching wires.

Connection kinds (animated on the map):

- cyan — command / query
- pink — async job / progress
- green — local persistence
- orange — external I/O

Actions such as **Scan now**, **Generate idea**, and **Save blueprint** emit SSE pulses so the matching wires glow.

### Outlier contract

A video is an outlier when:

```
viewsPerDay  ≥  multiplier  ×  average(viewsPerDay of recent same-channel videos, excluding itself)
```

Product default multiplier is **2.5x** (configure `OUTLIER_MULTIPLIER` in the 2–3x band so seed data still shows outliers).

The architecture diagram annotates `views/day + avg(channel views/day) * 100`. That `* 100` is a **visual exaggeration from early seed sketches** so spikes read on a dense map. It is **not** the runtime threshold.

Five feed signals: outlier score, velocity (views/day), recency, baseline gap, hook heuristic.

## Stubs vs real

| Port | Stub (no key / default) | Real |
|---|---|---|
| `VideoSourcePort` | Catalog of seed handles + synthetic recent videos | YouTube Data API v3 (`YOUTUBE_API_KEY`) |
| `AIProviderPort` | Vietnamese markdown draft that cites domain rules | OpenAI-compatible chat completions |
| `TranscriptPort` | Note that provider is unresolved (Supadata vs GoClaw) | Non-goal for MVP |
| `SyncExportPort` | No-op; Sheets never becomes source of truth | Optional later; still no Apps Script |
| `SecretPort` | Environment variables | OS Keychain is a non-goal for MVP |
| File export | `.md` / `.html` / `.pdf.txt` under `data/exports` | Same adapter; PDF is a text stand-in |

## Layout

```
cmd/server          HTTP + UI
cmd/scan            CLI / 60-minute job
internal/domain     Research, Intelligence, Brand, Script
internal/ports      inbound + outbound interfaces
internal/application  use cases
internal/adapters/http
internal/adapters/sqlite   migrations + seed
internal/adapters/youtube  live + stub
internal/adapters/ai
internal/adapters/{transcript,sheets,secrets,export}
web                 Vite + React Architecture Live View
```
