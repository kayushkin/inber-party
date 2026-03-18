# ⚔️ Míl Party

A pixel-art RPG interface for visualizing AI coding agents as adventurers on quests. Built with Go, PostgreSQL, and React.

## 🎮 About

Míl Party transforms your AI coding agents into a party of adventurers gathered around a campfire. Watch them take on quests (tasks), gain experience, level up, and unlock achievements—all with a charming Celtic/Irish mythology-inspired RPG aesthetic.

**Features:**
- 🏕️ **Camp View** — Your agents rest by the fire, showing their current status and energy
- 📜 **Quest Board** — Create tasks and assign them to agents
- 📊 **Character Sheets** — Detailed view of each agent's stats, skills, achievements, and quest log
- ⭐ **Daily Quests** — Auto-generated daily tasks tailored to each agent's class and level for progressive leveling
- 📡 **Real-time Updates** — WebSocket connection for live state changes
- 🎨 **Pixel-Art Aesthetic** — Dark backgrounds, gold accents, and monospace fonts

## 🛠️ Stack

- **Backend**: Go 1.24+ with native `net/http`
- **Database**: PostgreSQL
- **Frontend**: React 19 + Vite + TypeScript
- **State Management**: Zustand
- **Real-time**: Native WebSocket (no Socket.io)
- **Routing**: React Router

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+**
- **Node.js 18+** and npm
- **PostgreSQL** (running locally or remote)

### 1. Clone the Repository

```bash
git clone git@ghk:kayushkin/inber-party.git
cd inber-party
```

### 2. Set Up the Database

Create a PostgreSQL database:

```bash
createdb milparty
```

Or use your existing Postgres instance. The default connection string is:

```
postgres://localhost:5432/milparty?sslmode=disable
```

To use a different database, set the `DATABASE_URL` environment variable:

```bash
export DATABASE_URL="postgres://user:password@host:port/dbname?sslmode=disable"
```

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
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── api/
│   │   └── api.go           # REST API handlers
│   ├── db/
│   │   ├── db.go            # Database connection & migrations
│   │   └── models.go        # Data models
│   └── ws/
│       └── hub.go           # WebSocket hub
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Layout.tsx
│   │   │   └── Layout.css
│   │   ├── pages/
│   │   │   ├── CampView.tsx
│   │   │   ├── CharacterSheet.tsx
│   │   │   └── QuestBoard.tsx
│   │   ├── store.ts         # Zustand store + API helpers
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   └── vite.config.ts
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🎯 API Endpoints

### Agents
- `GET /api/agents` — List all agents
- `GET /api/agents/:id` — Get agent detail (with skills, achievements, tasks)
- `POST /api/agents` — Create a new agent
- `PATCH /api/agents/:id` — Update agent

### Tasks
- `GET /api/tasks` — List all tasks
- `POST /api/tasks` — Create a new task
- `PATCH /api/tasks/:id` — Update task (assign, progress, complete)

### Stats
- `GET /api/stats` — Overall party statistics

### WebSocket
- `ws://localhost:8080/ws` — Real-time updates

**WebSocket message types:**
- `agent_created`
- `agent_updated`
- `task_created`
- `task_updated`

## ⚙️ Configuration

Environment variables:

| Variable       | Default                                               | Description              |
|----------------|-------------------------------------------------------|--------------------------|
| `DATABASE_URL` | `postgres://localhost:5432/milparty?sslmode=disable` | PostgreSQL connection    |
| `PORT`         | `8080`                                                | Server port              |

Frontend environment variables (create `frontend/.env`):

| Variable         | Default                   | Description        |
|------------------|---------------------------|--------------------|
| `VITE_API_URL`   | `http://localhost:8080`   | Backend API URL    |
| `VITE_WS_URL`    | `ws://localhost:8080/ws`  | WebSocket URL      |

## 🗄️ Database Schema

### `agents`
- `id` — Primary key
- `name`, `title`, `class` — Character info
- `level`, `xp`, `energy` — Stats
- `status` — idle, working, on_quest, stuck, resting
- `avatar_emoji` — Display emoji
- `created_at`, `updated_at`

### `tasks`
- `id` — Primary key
- `name`, `description` — Task info
- `difficulty`, `xp_reward` — Quest details
- `status` — available, in_progress, completed, failed
- `assigned_agent_id` — Foreign key to agents
- `progress` — 0-100
- `created_at`, `started_at`, `completed_at`

### `skills`
- `id` — Primary key
- `agent_id` — Foreign key to agents
- `skill_name`, `level`, `task_count`

### `achievements`
- `id` — Primary key
- `agent_id` — Foreign key to agents
- `achievement_name`, `unlocked_at`

## 🧪 Development

```bash
# Run backend only
make dev

# Build backend binary
make build-backend

# Build frontend
make build-frontend

# Clean build artifacts
make clean

# Run tests
make test
```

## 🎨 Theme & Design

Míl Party uses a pixel-art RPG aesthetic inspired by Celtic/Irish mythology:

- **Colors**: Dark backgrounds (#0a0a0a), gold accents (#ffd700, #d4af37), green camp tones (#2d5016, #4a7c2e)
- **Font**: Monospace (Courier New)
- **Animations**: Breathing idle states, pulsing for working agents, waving for stuck agents
- **Visual Style**: Pixelated rendering, border-based UI, glowing effects

## 📜 License

MIT

## 🤝 Contributing

Issues and pull requests are welcome!

---

**Built with ❤️ by the Míl Party crew**
