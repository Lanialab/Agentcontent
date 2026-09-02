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

Open http://localhost:8080

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

Covers outlier scoring (`viewsPerDay` vs same-channel average) and the scan → feed use case.

## Environment

Copy `.env.example`. Important variables:

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_PATH` | `./data/agentcontent.db` | SQLite path (created + migrated on boot) |
| `OUTLIER_MULTIPLIER` | `2.5` | Outlier threshold (use 2–3) |
| `SEED_ON_EMPTY` | `true` | Load demo channels / videos / blueprint / scripts |
| `YOUTUBE_API_KEY` | empty | Empty → YouTube **stub** |
| `OPENAI_API_KEY` | empty | Empty → AI **stub** |
| `OPENAI_BASE_URL` | OpenAI | Any OpenAI-compatible endpoint |
| `SHEETS_ENABLED` | `false` | Optional sync adapter (no-op unless implemented later) |

No API keys are required to demo. Seed data is enough.

## Architecture map

The in-app **Kiến trúc / Architecture Live View** is the hexagonal modular monolith:

```
UI client (outside Go)
  → driving adapters (HTTP handlers + CLI/job)
    → inbound ports
      → application use cases (+ async job runtime)
        → domain (Research, Intelligence, Brand Blueprint, Kịch bản Video)
          → outbound ports
            → driven adapters (SQLite, YouTube, Transcript, AI, Sheets, Keychain, file export)
              → external systems
```

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
