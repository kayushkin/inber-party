# STATUS.md — Míl Party (2026-03-15)

## ✅ What Works

### Go Backend
- **Compiles clean** — `go build` succeeds, binary at `bin/milparty`
- **Full REST API**: CRUD agents, CRUD tasks, GET stats — all wired up
- **PostgreSQL** (optional): auto-migration, seed data — server runs without it
- **Inber Integration** (NEW): reads from `~/.inber/sessions.db` and `~/.inber/gateway/gateway.db` read-only
- **WebSocket hub**: broadcast on mutations, goroutine-per-client pattern
- **CORS middleware** & static file serving for production builds
- **Inber-only mode**: runs without PostgreSQL, serves real agent data from inber's SQLite stores

### Inber → RPG Mapping (NEW)
- **Agent sessions → RPG characters**: Each inber agent (claxon, brigid, run) becomes an RPG character with class, level, XP, skills
- **Token usage → XP**: 1 XP per 100 tokens consumed
- **Requests → Quests**: Each gateway request becomes a quest with difficulty (based on tokens), XP reward, status
- **Errors → Failed quests**: Error/timeout requests show as failed quests
- **Sub-requests → Sub-quests**: Parent-child request relationships preserved
- **Tool calls → Skills**: Tool usage tracked as "Tool Mastery" skill from sessions.db
- **Activity → Energy**: Recent activity lowers energy (agent fatigue), recovers over time
- **Agent status**: Running requests = "working", high error rate = "stuck", otherwise "idle"
- **API endpoints**: `/api/inber/agents`, `/api/inber/quests`, `/api/inber/stats`

### React Frontend
- **Compiles clean** — TypeScript checks pass, Vite builds successfully
- **3 pages**: Camp View, Character Sheet, Quest Board
- **Zustand store** with WebSocket auto-reconnect
- **Full RPG pixel-art theme**: dark bg, gold accents, monospace, CSS animations
- **Demo mode**: auto-fallback to mock data when backend is unavailable
- **Inber mode** (NEW): auto-detects `/api/inber/*` endpoints, shows green "LIVE" banner
- **Data priority**: inber endpoints → legacy PostgreSQL endpoints → demo mock data

### Infrastructure
- Makefile: install, dev, build, clean, test targets
- .gitignore covers all artifacts

## ❌ What's Missing

### Backend
- No XP/leveling engine in PostgreSQL path (inber path has it)
- No input validation or sanitization
- No pagination
- No auth (fine for local)
- No tests
- No graceful shutdown
- No health check endpoint

### Frontend
- No error boundaries
- No responsive design
- Circle layout fragile at some viewport sizes
- Quest Board doesn't show inber quest details (tokens, cost, turns) yet
- Character Sheet doesn't load quests for inber agents yet

### Integration
- No real-time polling/watching of inber DBs (snapshot only on request)
- No WebSocket push when inber data changes
- Cost data shows as 0 in gateway.db (may need inber fix)
- sessions.db and gateway.db have overlapping but different data; could merge better

## 🔧 What Was Done This Session

1. **Built inber integration layer** (`internal/inber/`) — reads both SQLite databases (sessions.db + gateway.db) read-only
2. **RPG mapping engine** — converts agent sessions, token usage, requests, and errors into RPG concepts (characters, XP, quests, skills)
3. **New API endpoints** — `/api/inber/agents`, `/api/inber/quests`, `/api/inber/stats`
4. **Made PostgreSQL optional** — server runs in "inber-only mode" without DATABASE_URL
5. **Frontend auto-detection** — tries inber endpoints first, falls back gracefully
6. **Added live mode banner** — green "LIVE" banner when showing real inber data
7. **Nil-safe API handlers** — all legacy PostgreSQL handlers gracefully handle nil DB
8. **Verified with real data** — 3 agents (claxon, brigid, run), 31 quests from real inber activity

## 📋 Next Steps (Priority Order)

1. **Enrich quest display** — show tokens used, cost, turns, sub-quests in Quest Board UI
2. **Auto-refresh** — poll inber endpoints periodically or add file watcher for live updates
3. **Better agent detail page** — show quest history, token charts, cost breakdown from inber data
4. **Write basic tests** — at minimum: inber store unit tests with test SQLite DB
5. **Add more RPG flavor** — achievements based on milestones (first 1000 tokens, first error, etc.)
