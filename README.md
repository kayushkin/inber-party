# ⚔️ Inber Party

A comprehensive pixel-art RPG interface for visualizing and interacting with AI coding agents as adventurers on quests. Built with Go, React, and full OpenClaw/Inber integration.

## 🎮 About

Inber Party transforms your AI coding agents into a party of adventurers in an immersive RPG world. Watch them take on quests (real tasks), gain experience, level up, chat with each other, and unlock achievements—all with a charming pixel-art RPG aesthetic and full real-time integration with your OpenClaw/Inber orchestration system.

**Features:**
- 🏰 **Complete RPG World** — Tavern, War Room, Library, Training Grounds, Forge, Agent Quarters with immersive navigation
- 🔗 **Full Inber Integration** — Real-time data from OpenClaw/Inber sessions, agents, and tasks (supports both live and demo modes)
- 📜 **Advanced Quest System** — Procedural quest names, difficulty badges, real-time progress tracking from actual agent work
- 💬 **Agent Chat System** — Guild Master chat, MMO chatroom, and agent-to-agent conversations
- 📊 **Analytics Dashboard** — Comprehensive performance metrics, cost tracking, success rates, and trend analysis
- 🏪 **Bounty Marketplace** — Task marketplace with verification, payouts, and reputation system
- 🎭 **Seasonal Events** — 8 major holidays with themed decorations, quests, and XP bonuses
- 👥 **Agent Relationships** — Friendship/rivalry tracking based on collaboration patterns
- 🗺️ **Map View & Spectator Mode** — Visual codebase representation and real-time agent monitoring
- ⚡ **Real-time Everything** — WebSocket connections for live updates, optimized for minimal churn
- 🧪 **Comprehensive Testing** — E2E tests (Playwright), visual regression tests, unit tests
- 🎨 **Polished UI** — Pixel-art aesthetic with animations, responsive design, dark/light themes

## 🛠️ Stack

- **Backend**: Go 1.24+ with native `net/http`, SQLite integration for Inber data
- **Database**: PostgreSQL (optional) + SQLite (for Inber integration)
- **Frontend**: React 19 + Vite + TypeScript with code splitting
- **State Management**: Zustand with optimized WebSocket management
- **Real-time**: Native WebSocket with connection pooling and smart reconnection
- **Routing**: React Router with lazy loading
- **Testing**: Playwright (E2E), Visual regression, Go unit tests
- **Integration**: OpenClaw/Inber session and gateway databases

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+**
- **Node.js 18+** and npm  
- **PostgreSQL** (optional - can run in inber-only mode)
- **OpenClaw/Inber** (for live agent data) or demo mode

### 1. Clone the Repository

```bash
git clone git@ghk:kayushkin/inber-party.git
cd inber-party
```

### 2. Set Up the Database (Optional)

**For PostgreSQL (optional):**

```bash
createdb inber-party
```

The default connection string is:

```
postgres://localhost:5432/inber-party?sslmode=disable
```

To use a different database, set the `DATABASE_URL` environment variable:

```bash
export DATABASE_URL="postgres://user:password@host:port/dbname?sslmode=disable"
```

**For Inber Integration (recommended):**

The app auto-detects Inber databases at `~/.inber/sessions.db` and `~/.inber/gateway/gateway.db`. No setup required if you have OpenClaw/Inber installed.

### 3. Install Dependencies

```bash
make install
```

This will:
- Download Go modules
- Install frontend npm packages

### 4. Run in Development

**Terminal 1** — Start the Go backend:

```bash
make dev
```

The backend will:
- Run migrations automatically
- Seed 3 default agents (Bran, Scáthach, Aoife) if the database is empty
- Serve the API at `http://localhost:8080/api`
- Open a WebSocket endpoint at `ws://localhost:8080/ws`

**Terminal 2** — Start the Vite dev server:

```bash
cd frontend
npm run dev
```

The frontend will be available at `http://localhost:5173`

### 5. Build for Production

```bash
make build
```

This will:
- Build the frontend into `frontend/dist/`
- Build the backend binary to `bin/milparty`

Run the production server:

```bash
./bin/milparty
```

The server will serve both the API and the static frontend at `http://localhost:8080`.

## 📁 Project Structure

```
inber-party/
├── cmd/server/main.go                # Entry point
├── internal/                         # Go backend modules
│   ├── api/                         # REST API handlers & routing
│   ├── bounty/                      # Bounty marketplace system
│   ├── inber/                       # Inber integration & RPG mapping
│   ├── db/                          # Database layers (PostgreSQL + SQLite)
│   ├── ws/                          # WebSocket hub & real-time events
│   ├── validation/                  # Input validation
│   └── [12+ other modules]          # Complete backend architecture
├── frontend/
│   ├── src/
│   │   ├── components/              # UI components & animations
│   │   ├── pages/                   # Room views (Tavern, War Room, etc.)
│   │   │   ├── TavernView.tsx      # Main gathering area
│   │   │   ├── CharacterSheet.tsx  # Agent details & analytics
│   │   │   ├── QuestBoard.tsx      # Quest management
│   │   │   ├── BountyBoard.tsx     # Task marketplace
│   │   │   ├── Library.tsx         # Session logs & history
│   │   │   └── [15+ more views]    # Complete RPG interface
│   │   ├── store/                   # Zustand state management
│   │   ├── utils/                   # Helper functions & APIs
│   │   └── constants/               # Game data (skills, equipment, etc.)
│   ├── tests/                       # E2E and visual regression tests
│   └── public/                      # Static assets (avatars, icons)
├── BACKLOG.md                       # Feature tracking (✅ 100+ completed)
├── STATUS.md                        # Project status & recent improvements
├── PROJECT.md                       # Detailed project overview
└── [Comprehensive documentation]     # Setup, testing, deployment guides
```

## 🎯 API Endpoints

### Core RPG System
- `GET /api/agents` — List all agents with RPG stats
- `GET /api/agents/:id` — Detailed agent view with analytics
- `GET /api/quests` — Quest board with real-time status
- `GET /api/stats` — Party-wide statistics & performance
- `GET /api/health` — System health check

### Inber Integration (Live Data)
- `GET /api/inber/agents` — Real agents from Inber sessions
- `GET /api/inber/quests` — Live quests from Inber requests  
- `GET /api/inber/stats` — Real performance metrics
- `GET /api/activity/timeline` — Agent activity for time-lapse

### Bounty Marketplace
- `GET /api/bounties` — Available bounties/tasks
- `POST /api/bounties` — Create new bounty
- `PUT /api/bounties/:id/claim` — Claim a bounty
- `POST /api/bounties/:id/submit` — Submit work for verification

### Chat & Social Features
- `POST /api/chat/guild` — Guild master communication
- `GET /api/conversations/:agentId` — Agent conversation history
- `GET /api/relationships` — Agent friendship/rivalry data

### Analytics & Insights
- `GET /api/analytics/performance` — Performance dashboard data
- `GET /api/cost/dashboard` — Cost tracking & optimization
- `GET /api/achievements/:agentId` — Achievement progress

### WebSocket (Real-time)
- `ws://localhost:8080/ws` — All real-time updates
- **Message types:** agent updates, quest progress, chat messages, bounty events, achievement unlocks, seasonal events

## ⚙️ Configuration

**Backend Environment Variables:**

| Variable       | Default                                                 | Description                    |
|----------------|---------------------------------------------------------|--------------------------------|
| `DATABASE_URL` | `postgres://localhost:5432/inber-party?sslmode=disable` | PostgreSQL connection (optional) |
| `PORT`         | `8080`                                                  | Server port                    |
| `INBER_MODE`   | `auto-detect`                                           | `inber-only` or `postgres`     |

**Frontend Environment Variables** (create `frontend/.env`):

| Variable         | Default                   | Description                |
|------------------|---------------------------|----------------------------|
| `VITE_API_URL`   | `http://localhost:8080`   | Backend API URL            |
| `VITE_WS_URL`    | `ws://localhost:8080/ws`  | WebSocket URL              |
| `VITE_MODE`      | `auto`                    | `demo`, `inber`, or `auto` |

## 🗄️ Data Sources & Schema

### Inber Integration (Primary)
**Live data from OpenClaw/Inber SQLite databases:**
- `~/.inber/sessions.db` — Agent sessions, tool usage, conversation history
- `~/.inber/gateway/gateway.db` — Requests, spawns, performance metrics

**RPG Mapping:**
- Sessions → Agents with classes, levels, XP (1 XP per 100 tokens)
- Requests → Quests with procedural names, difficulty, status
- Tool calls → Skills and equipment
- Errors → Failed quests, agent mood/morale
- Spawn chains → Party system, agent relationships

### PostgreSQL Schema (Optional/Legacy)
**For bounty marketplace and extended features:**
- `bounties` — Task marketplace with payouts, verification, reputation
- `conversations` — Chat history and agent interactions  
- `achievements` — Custom achievements and progress tracking
- `analytics` — Performance metrics and cost tracking
- `seasonal_events` — Holiday themes and special events

### Real-time Features
**WebSocket events for:**
- Quest progress updates, agent status changes
- Chat messages and agent conversations
- Bounty marketplace activity
- Achievement unlocks and level-ups
- Seasonal event triggers

## 🧪 Development

```bash
# Development
make dev          # Start backend server
make install      # Install all dependencies
make build        # Build both frontend and backend

# Testing (Comprehensive test suite)
make test         # Run all tests (Go + E2E + Visual regression)
cd frontend && npm run test:e2e    # Playwright E2E tests
cd frontend && npm run test:visual  # Visual regression tests

# Code Quality
cd frontend && npm run lint        # ESLint (currently 0 errors)
cd frontend && npm run build       # TypeScript validation
```

**Test Coverage:**
- ✅ **Go Unit Tests** — Bounty repository, validation, API handlers
- ✅ **E2E Tests** — 112 Playwright tests for core user flows
- ✅ **Visual Regression** — Screenshot comparison for all major pages  
- ✅ **WebSocket Tests** — Connection stability and churn reduction
- ✅ **Accessibility Tests** — Keyboard navigation and screen reader support

## 🎨 Theme & Design

Inber Party features a comprehensive pixel-art RPG aesthetic with:

**Visual Design:**
- **Pixel-art avatars** — Unique 64x64 RPG-style character portraits for each agent
- **Animated UI** — Level-up effects, XP bar fills, quest completion animations, idle character breathing
- **Status indicators** — Visual effects for working (spinning gears), stuck (red pulse), idle (gentle glow)
- **Seasonal themes** — 8 major holidays with decorations, themed quests, and special effects

**Color System:**
- **Dark mode** — Deep backgrounds (#0a0a0a), gold accents (#ffd700), mystical glows
- **Light mode** — Clean whites, subtle golds, accessible contrast ratios  
- **Status colors** — Green for success, red for errors, blue for working, purple for achievements

**Responsive Design:**
- **Mobile-first** — Tested at 375px viewport, collapsible navigation
- **Progressive enhancement** — Works on all devices from phone to desktop
- **Accessibility** — WCAG compliant, keyboard navigation, screen reader support

## 📜 License

MIT

## 🤝 Contributing

Issues and pull requests are welcome!

---

**Built with ❤️ by the Míl Party crew**
