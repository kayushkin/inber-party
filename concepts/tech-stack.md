# Tech Stack — Implementation Approach

## Overview

This document outlines the proposed technical architecture for building Míl Party — transforming the concept into a working application.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Míl Party UI                       │
│          (React + PixiJS/Canvas)                    │
└──────────────┬──────────────────────────────────────┘
               │ WebSocket / HTTP
┌──────────────▼──────────────────────────────────────┐
│            Míl Party Server                         │
│       (Node.js / Go / Python)                       │
│  - WebSocket server (real-time updates)             │
│  - REST API (CRUD operations)                       │
│  - Event processor (task events → UI state)         │
└──────────────┬──────────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────────┐
│         Inber Core / Session DB                     │
│    (existing multi-agent system)                    │
│  - Agent sessions (running tasks)                   │
│  - Task queue                                       │
│  - Agent state & history                            │
└─────────────────────────────────────────────────────┘
```

## Frontend Stack

### Framework: **React**

**Why React?**
- Component-based architecture fits agent cards, quest cards, etc.
- Strong ecosystem (state management, routing)
- Easy to integrate with Canvas/PixiJS
- Wide developer familiarity

**Alternatives considered:**
- **Vue** — simpler, lighter, but smaller ecosystem
- **Svelte** — fastest, but less mature tooling
- **Vanilla JS** — too much boilerplate for complex UI

### Rendering: **PixiJS** (or HTML5 Canvas)

**Why PixiJS?**
- High-performance 2D rendering (handles many sprites)
- WebGL acceleration
- Great for pixel art games
- Animation and particle effects built-in
- Large sprite count (camp scene with many agents)

**Alternatives:**
- **Phaser** — full game framework, might be overkill
- **Three.js** — 3D focus, unnecessary complexity
- **Plain Canvas** — more control, but more work

**Decision:** Start with PixiJS for camp view, character sprites, and battle animations. Use React for UI panels (quest board, character sheets).

### State Management: **Zustand** or **Redux Toolkit**

**Why Zustand?**
- Lightweight, simple API
- No boilerplate
- Works great with React hooks
- Good for real-time updates

**Why Redux Toolkit?**
- More structure (good for large state)
- DevTools for debugging
- Middleware for WebSocket integration

**Decision:** Start with Zustand for simplicity. Migrate to Redux if state grows complex.

### Styling: **TailwindCSS** + **CSS Modules**

- **Tailwind** for rapid prototyping and utility classes
- **CSS Modules** for component-specific styles (pixel art effects)
- Custom theme for RPG aesthetic (medieval fonts, parchment textures)

### Real-Time Updates: **WebSocket** (Socket.io or native WebSocket)

- Subscribe to agent events (task started, progress update, task complete)
- Push updates to UI instantly
- Fallback to polling if WebSocket unavailable

## Backend Stack

### Server: **Node.js + Express** (or Go/Python)

**Why Node.js?**
- Same language as frontend (TypeScript across the stack)
- Excellent WebSocket support (Socket.io)
- Fast prototyping
- Large ecosystem

**Alternatives:**
- **Go** — faster, better concurrency, but different language
- **Python** — integrates easily with AI tools, but slower

**Decision:** Node.js + TypeScript for initial prototype. Consider Go for production if performance becomes an issue.

### Database: **PostgreSQL** (or SQLite for prototype)

**Schema:**

#### Tables

**agents**
```sql
id, name, level, xp, class, energy, created_at, updated_at
```

**tasks** (quests)
```sql
id, name, description, difficulty, xp_reward, status, assigned_agent_id, created_at, completed_at
```

**task_history**
```sql
id, task_id, agent_id, action, timestamp, details (jsonb)
```

**skills**
```sql
id, agent_id, skill_name, level, task_count, last_used
```

**achievements**
```sql
id, agent_id, achievement_name, unlocked_at
```

**party_quests**
```sql
id, task_id, agent_ids (array), synergy_bonus, status
```

### Data Model

#### Agent State
```typescript
interface Agent {
  id: string;
  name: string;
  level: number;
  xp: number;
  xpToNextLevel: number;
  class: 'wizard' | 'warrior' | 'ranger' | 'cleric';
  energy: number; // 0-100
  status: 'idle' | 'working' | 'on_quest' | 'stuck' | 'resting';
  currentTaskId?: string;
  avatarUrl: string;
  stats: {
    tasksCompleted: number;
    successRate: number;
    avgTime: number;
    linesOfCode: number;
    testsWritten: number;
    bugsFixed: number;
  };
  skills: Skill[];
  achievements: Achievement[];
}
```

#### Task State
```typescript
interface Task {
  id: string;
  name: string;
  description: string;
  difficulty: 'easy' | 'medium' | 'hard' | 'very_hard';
  xpReward: number;
  estimatedTime: number; // hours
  status: 'available' | 'in_progress' | 'completed' | 'failed';
  assignedAgentId?: string;
  startedAt?: Date;
  completedAt?: Date;
  progress: number; // 0-100
  errorMessage?: string;
}
```

#### Skill
```typescript
interface Skill {
  name: string; // 'TypeScript', 'React', etc.
  level: 1 | 2 | 3 | 4;
  taskCount: number;
  tasksToNextLevel: number;
}
```

## Integration with Inber

### Option 1: Direct Database Access
- Míl Party reads from Inber's session database
- Maps sessions → agents, tasks → quests
- Pros: Real-time, accurate
- Cons: Tight coupling, DB schema dependency

### Option 2: Event Stream
- Inber emits events (task started, completed, etc.)
- Míl Party listens and updates its own database
- Pros: Loose coupling, can add features without touching Inber
- Cons: Eventual consistency, event replay needed

### Option 3: API Layer
- Inber exposes REST/GraphQL API
- Míl Party polls or subscribes
- Pros: Clean separation, versioned API
- Cons: Latency, more infrastructure

**Decision:** Start with **Option 2 (Event Stream)** using WebSocket or message queue (Redis Pub/Sub).

### Event Examples

```typescript
// Task started
{
  type: 'task.started',
  agentId: 'bran',
  taskId: 'task-123',
  taskName: 'Fix login bug',
  difficulty: 'medium',
  timestamp: '2026-02-27T12:00:00Z'
}

// Task progress
{
  type: 'task.progress',
  agentId: 'bran',
  taskId: 'task-123',
  action: 'read_file',
  file: 'auth.ts',
  progress: 25,
  timestamp: '2026-02-27T12:05:00Z'
}

// Task completed
{
  type: 'task.completed',
  agentId: 'bran',
  taskId: 'task-123',
  xpAwarded: 50,
  duration: 3600, // seconds
  timestamp: '2026-02-27T13:00:00Z'
}

// Task failed
{
  type: 'task.failed',
  agentId: 'bran',
  taskId: 'task-123',
  error: 'Database connection failed',
  timestamp: '2026-02-27T12:30:00Z'
}

// Agent leveled up
{
  type: 'agent.level_up',
  agentId: 'bran',
  newLevel: 13,
  xpGained: 50,
  timestamp: '2026-02-27T13:00:05Z'
}
```

## Pixel Art Pipeline

### Asset Creation

1. **DALL-E Generation**
   - Prompt: "64x64 pixel art character, fantasy RPG, Celtic warrior, idle stance"
   - Output: PNG image

2. **Manual Editing** (optional)
   - Aseprite or Piskel for touch-ups
   - Create animation frames (idle, walking, working, etc.)

3. **Sprite Sheets**
   - Combine frames into sprite sheets
   - Use TexturePacker or manual grid

4. **PixiJS Loading**
   - Load sprite sheets as textures
   - Animate using PixiJS AnimatedSprite

### Example Sprite Sheet Layout

```
┌────┬────┬────┬────┐
│Idle│Walk│Work│Wave│  ← Bran
│ 1  │ 2  │ 3  │ 4  │
├────┼────┼────┼────┤
│Idle│Walk│Work│Wave│  ← Scáthach
│ 1  │ 2  │ 3  │ 4  │
├────┼────┼────┼────┤
...
```

Each cell: 64x64 pixels

### Color Palette

Celtic/Irish theme:
- **Greens** — #2d5016, #4a7c2e, #76a34d (forest, grass)
- **Golds** — #d4af37, #ffdf00, #ffd700 (armor, accents)
- **Browns** — #5c4033, #8b7355, #a0826d (leather, wood)
- **Blues** — #1e3a5f, #2e5984, #4a90c4 (magic, water)
- **Reds** — #8b0000, #b22222, #dc143c (critical, errors)

## Deployment

### Development
- **Frontend**: Vite dev server (`npm run dev`)
- **Backend**: Nodemon (`npm run dev`)
- **Database**: Docker Compose (PostgreSQL + Redis)

### Production
- **Frontend**: Static build → CDN (Vercel, Netlify, Cloudflare Pages)
- **Backend**: Node.js server on VPS or serverless (AWS Lambda, Fly.io)
- **Database**: Managed PostgreSQL (Supabase, Railway, Render)
- **WebSocket**: Separate WebSocket server or Socket.io with sticky sessions

### Monitoring
- **Sentry** for error tracking
- **Prometheus + Grafana** for metrics
- **WebSocket health checks** (reconnect on disconnect)

## Milestones

### Phase 1: MVP (Camp View)
- [ ] Static camp scene with pixel art placeholders
- [ ] 3 agents (Bran, Scáthach, Aoife) with idle/working states
- [ ] Real-time updates from mock task events
- [ ] Click agent → show basic stats

### Phase 2: Quest Board
- [ ] Quest cards (available, in progress, completed)
- [ ] Drag-and-drop quest assignment
- [ ] Task creation form

### Phase 3: Character Sheets
- [ ] Full agent stats and history
- [ ] Skill tree visualization
- [ ] Quest log

### Phase 4: Battle View
- [ ] Live action feed (battle log)
- [ ] HP bars for agent and task
- [ ] Animations for actions

### Phase 5: Progression
- [ ] XP and leveling system
- [ ] Achievements
- [ ] Leaderboards

### Phase 6: Party Management
- [ ] Party composition UI
- [ ] Synergy calculation
- [ ] Party quest execution

### Phase 7: Polish
- [ ] Actual pixel art (replace placeholders)
- [ ] Sound effects and music
- [ ] Animations and transitions
- [ ] Mobile responsive layout

## Open Questions

1. **Inber integration details** — Does Inber already emit events? Or do we need to add hooks?
2. **Multi-user support** — One Míl Party per user or shared team view?
3. **Authentication** — How do users log in? OAuth? Magic link?
4. **Permissions** — Can anyone assign quests? Or role-based access?
5. **Scalability** — How many agents per user? 5? 20? 100?

## Next Steps

1. **Prototype camp view** (static HTML/CSS) ✓ (this repo)
2. **Prototype character sheet** (static HTML/CSS) ✓ (this repo)
3. **Spike: PixiJS + React integration** — test sprite rendering
4. **Spike: Inber event stream** — figure out event format
5. **Build MVP** — camp view + real-time agent status
6. **User testing** — get feedback, iterate

---

The tech stack is the foundation. Keep it simple, keep it fun, keep it working.
