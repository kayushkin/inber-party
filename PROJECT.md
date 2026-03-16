# PROJECT.md — Míl Party

## What This Is

Pixel-art RPG UI for visualizing AI coding agents (from the [inber](https://github.com/kayushkin/inber) framework) as Celtic mythology characters around a campfire. Observability, not control — watch your agents quest, level up, and (occasionally) get stuck.

## Architecture Decisions Already Made

**Backend: Go + Postgres + WebSocket**
- Go `net/http` (stdlib, no framework) with gorilla/websocket for real-time
- PostgreSQL via `lib/pq` — relational schema for agents, tasks, skills, achievements
- Auto-migrations on startup (CREATE TABLE IF NOT EXISTS), seed data for 3 default agents
- CORS middleware baked into main.go (allow-all for dev)
- WebSocket hub pattern: broadcast channel, client register/unregister, goroutine per client

**Frontend: React 19 + Vite + TypeScript + Zustand**
- Zustand for global state (agents, tasks, stats, WS connection)
- React Router for 3 pages: Camp View, Character Sheet, Quest Board
- WebSocket auto-reconnect (3s interval)
- API helper object (`api.getAgents()`, etc.) co-located with store
- CSS-per-page (no Tailwind, no CSS-in-JS) — hand-written pixel-art theme

**Data Model**
- `agents`: id, name, title, class, level, xp, energy, status, avatar_emoji
- `tasks`: id, name, description, difficulty, xp_reward, status, assigned_agent_id, progress
- `skills`: agent_id + skill_name, level, task_count (unique constraint)
- `achievements`: agent_id + achievement_name, unlocked_at (unique constraint)

**Design Aesthetic**
- Dark backgrounds (#0a0a0a, #1a1a1a), gold accents (#ffd700, #d4af37), forest greens (#2d5016, #4a7c2e)
- Monospace font (Courier New)
- CSS animations: breathing idle, pulsing work, waving stuck, flickering campfire
- Emoji avatars (🧙🏹⚔️) — pixel-art sprites planned but not implemented

## What's Working

### Backend ✅
- PostgreSQL connection, auto-migration, seed data (Bran, Scáthach, Aoife)
- Full REST API: CRUD agents, CRUD tasks, GET stats
- Dynamic PATCH endpoints (partial updates with query builder)
- WebSocket hub with broadcast on every mutation
- Static file serving for production builds
- Binary builds to `bin/milparty`

### Frontend ✅
- **Camp View**: Agents arranged in a circle around animated campfire, click-to-navigate
- **Character Sheet**: Stats grid, XP progress bar, energy bar, skills list, achievements list, quest log
- **Quest Board**: Create quests via modal, assign agents via dropdown, mark complete, status filtering
- Zustand store with WebSocket integration (auto-reconnect, live updates)
- Full RPG visual theme with animations

### Infra ✅
- Makefile: dev, build, install, clean, test targets
- .gitignore covers binaries, node_modules, dist, .env, IDE files
- Vite dev proxy-ready (not configured yet)

## What's Missing / Broken

### No inber integration
This is the big one. The app is currently a standalone RPG dashboard with manual task creation. It doesn't read from inber's session logs, session DB, or memory store at all. The agents are seeded statically.

### No real data pipeline
- No session log parser (inber writes JSONL to `.inber/workspace/` and `logs/`)
- No session DB reader (inber uses SQLite at `.inber/sessions.db`)
- No file watcher or polling for live session updates
- Agent status is manual (PATCH via API), not derived from active inber sessions

### No pixel-art sprites
Using emoji avatars. The `agents/avatars/` directory in the inber repo has a generation script but no actual sprites.

### No XP/leveling logic
The schema has level + xp fields but there's no server-side logic to award XP on task completion, calculate level-ups, or unlock achievements.

### No tests
Zero tests on backend or frontend. The Makefile has a `test` target but nothing to run.

### Frontend gaps
- Vite proxy not configured (CORS reliance instead)
- No error boundaries or loading states (minimal)
- No responsive design
- Agent cards overlap at certain viewport sizes (CSS circle layout is fragile)
- No dark mode toggle (it's always dark)

### Backend gaps
- No input validation or sanitization
- No pagination on list endpoints
- No authentication (fine for local, bad for anything else)
- Dynamic PATCH query builder is SQL-injection-adjacent (uses parameterized queries but hand-rolled)
- No graceful shutdown
- No health check endpoint

## Phased Roadmap

### Phase 1: Connect to inber (the reason this exists)
1. **Session DB reader** — Read inber's SQLite `sessions.db` (SessionRow, TurnRow tables). Map sessions → quests, turns → progress updates.
2. **Agent sync** — On startup, read inber's `agents.json` to populate/update the agents table. Map agent configs to RPG characters (class, title, emoji from agent identity files).
3. **Active session poller** — Watch `.inber/active/` directory for active session JSON files. Update agent status (idle → working → idle) in real-time.
4. **JSONL log tailer** — Tail active session JSONL files. Extract tool calls, token counts, errors. Push events through WebSocket.

### Phase 2: Make it real
5. **XP engine** — Award XP based on: tokens used, tool calls made, tasks completed. Auto-level-up with configurable thresholds.
6. **Skill tracking** — Map tool usage to skills (shell → "Command Mastery", edit_file → "Code Weaving", etc.). Level up skills based on usage frequency.
7. **Achievement system** — First quest, 10 quests completed, 100k tokens used, survived an error, etc. Trigger on relevant events.
8. **Cost tracking** — Pull input/output token counts from session logs, calculate costs per agent/quest using model pricing.

### Phase 3: Polish the UI
9. **Pixel-art sprites** — Replace emoji with actual 32×32 or 64×64 pixel art. Idle/working/stuck animations per character.
10. **Camp scene overhaul** — Proper 2D scene with ground, trees, sky. Agents positioned around fire with proper layering.
11. **Quest timeline** — Visual timeline per quest showing tool calls as "actions" (spell cast, sword swing, etc.).
12. **Responsive layout** — Make it usable on mobile/tablet.
13. **Sound effects** — Optional subtle RPG sounds on events (quest complete fanfare, level-up chime).

### Phase 4: Advanced features
14. **SSE/streaming** — Replace polling with server-sent events for log tailing.
15. **Historical dashboard** — Charts: tokens over time, cost over time, quests per agent, XP curves.
16. **Multi-project support** — Track agents across multiple inber repos.
17. **Drag-and-drop task assignment** — Drop a task description onto an agent to trigger an inber session.
18. **Party composition view** — Which agents work well together (based on sub-agent spawn patterns from session data).

## Quick Reference

```bash
# Dev
make install            # Go modules + npm install
make dev                # Backend on :8080
cd frontend && npm run dev  # Frontend on :5173

# Build
make build              # Both → bin/milparty + frontend/dist/

# Prod
DATABASE_URL=... ./bin/milparty  # Serves API + static frontend on :8080
```

**Stack**: Go 1.24 / PostgreSQL / React 19 / Vite 7 / TypeScript 5.9 / Zustand 5 / gorilla/websocket
