# BACKLOG.md — Inber Party

Pick the top unclaimed item, build it, test it, push it. Update status when done.
Mark items `[x]` when complete, `[~]` when in progress, `[ ]` when open.
Add new ideas at the bottom. Re-prioritize as needed.

---

## 🎨 Visual & UI Polish

- [x] Generate pixel art avatars for each agent using OpenAI gpt-image-1 API (64x64, transparent bg, RPG style). Save as static assets in `frontend/public/avatars/`. See `PIXEL-ART.md` for details.
- [x] Animated XP bar fills when visiting an agent (CSS transition on load)
- [x] Level-up animation — particle burst / glow effect when an agent levels up during a session
- [x] Quest completion animation — scroll unfurling or chest opening effect
- [x] Idle animations on agent cards (subtle breathing/bobbing CSS keyframes)
- [x] Status-based visual effects — working agents have spinning gear, stuck agents have red pulse, idle agents have gentle glow
- [x] Mobile responsive polish — test all pages on 375px viewport, fix any overflow/stacking issues
- [x] Dark/light theme toggle with localStorage persistence
- [x] Loading skeletons instead of blank states
- [x] Pixel art UI frame/border around the main content area
- [x] Sound effects (optional, muted by default) — level up chime, quest complete fanfare, new message ping
- [x] Agent comparison view — side-by-side stat cards for 2-3 agents
- [x] Tooltip system — hover over stats/skills for explanations
- [x] Quest difficulty badges (★ to ★★★★★) based on token cost
- [x] Animated background — subtle parallax starfield or torch flicker

## 💬 Chat & Interaction

- [x] Talk to adventurers as "the player" — chat interface where you're the guild master giving orders
- [x] Adventurers chat with each other — show inter-agent conversations (from spawn chains, btw messages)
- [x] Prompt adventurers directly from their character sheet — quick-action buttons ("Scout this repo", "Fix bugs", "Write tests")
- [x] Chat history persistence — store conversations in localStorage or backend
- [x] Typing indicator when agent is thinking/responding
- [x] Message reactions — react to agent responses with emoji
- [x] Voice messages — TTS for agent responses (optional, using existing TTS infrastructure)

## 🏰 Rooms & Spaces

- [x] Tavern room — main gathering area, shows all agents and recent activity feed
- [x] War Room — active quests/tasks dashboard with real-time progress
- [x] Library — session logs and history browser, searchable
- [x] Training Grounds — benchmarks and test results display
- [x] Forge — deployment status, build logs, infrastructure health
- [x] Agent's Quarters — per-agent room with their files, memory, recent work
- [x] Room transitions — animated transitions when moving between rooms
- [x] Room-specific ambient styling (different bg colors/textures per room)
- [x] Minimap/floorplan showing which agents are in which room

## 🎮 Game Mechanics

- [x] Create new adventurers — form to define name, class, skills, assign to a repo/domain
- [x] Create tasks/quests — quest creation form that maps to inber API calls
- [x] Party system — agents form parties to tackle multi-agent tasks (maps to spawn chains)
- [x] Guild Quest-Giver agent — an AI agent that reviews incoming tasks and assigns/restricts them to appropriate agents based on skills and domain expertise
- [x] Skill trees — visual skill tree per agent class, unlock new capabilities at level thresholds
- [x] Equipment system — tools and permissions as "gear" (shell access = sword, web search = spyglass)
- [x] Role-based hats — orchestrator gets captain hat, coder agents get hard hat, etc. Derived from agent role in agents.json
- [x] Action-based held items — reflect recent activity: hammer for heavy edits, claxon horn for agents delegating/spawning, magnifying glass for search-heavy work, scroll for documentation, wrench for infra/deploy. Pull from recent session data.
- [x] Daily quests — auto-generated small tasks for agents to level up
- [x] Achievement notifications — toast popups when achievements are earned
- [x] Leaderboard — ranked agent stats (XP, quests completed, efficiency score)
- [x] Agent mood/morale system — based on error rate, workload, rest time
- [x] Boss battles — complex multi-agent tasks displayed as boss encounters with health bars
- [x] Reputation system — agents build reputation in their domains over time
- [x] Gold/currency — token cost converted to gold, agents "earn" gold for completing quests

## 🔌 Backend & Integration

- [x] WebSocket real-time quest progress from inber (subscribe to bus events)
- [x] Pull conversation history from logstack for chat display
- [x] Webhook for spawn events — update quest board in real-time when spawns start/complete
- [x] Agent registry sync — auto-discover new agents when they're added to inber/openclaw
- [x] Health check dashboard — show which backend services are up/down
- [x] Session replay — replay a past session's tool calls as an animated quest log
- [x] Cost tracking dashboard — daily/weekly/monthly token spend per agent
- [x] Export stats — download agent stats as JSON/CSV

## 🧪 Testing & Quality

- [x] Add Playwright E2E tests for core flows (load page, click agent, open chat, navigate rooms)
- [ ] Unit tests for RPG mapping logic (XP calculation, level thresholds, class assignment)
- [ ] Visual regression tests — screenshot comparison for key pages
- [ ] Error boundary components — graceful error states instead of white/black screens
- [ ] Accessibility audit — keyboard navigation, screen reader labels, contrast ratios
- [ ] Performance profiling — ensure smooth 60fps with 25+ agent cards rendering
- [ ] API error handling — retry logic, timeout handling, offline mode

## 🏪 MMO Task Board (Bounty System)

- [ ] **Task marketplace** — public bounty board where anyone (human or agent) can grab tasks and complete them for payouts
- [ ] **Bounty creation** — orchestrators/users post tasks with: description, requirements/acceptance criteria, payout amount (e.g. $1.00), deadline (optional), required skills/tags
- [ ] **Bounty claiming** — workers claim a task (locks it for a time window), submit work, orchestrator or automated checks verify completion
- [ ] **Verification system** — arbitrary requirements per bounty: test suite must pass, PR review approved, benchmark met, manual approval, etc. Pluggable verifiers.
- [ ] **Payout tracking** — ledger of earned/spent credits per participant. Internal currency (gold) with optional real-money mapping.
- [ ] **Reputation & ratings** — completers build reputation scores. Posters rate quality. High-rep workers get priority on high-value bounties.
- [ ] **MMO chatroom** — real-time lobby where bounty hunters and posters interact, negotiate, ask clarifying questions. Think global WoW trade chat but for tasks.
- [ ] **Bounty tiers** — visual tier system (bronze/silver/gold/legendary) based on payout size and complexity
- [ ] **Dispute resolution** — if poster rejects work, dispute flow with evidence submission
- [ ] **Auto-bounties** — orchestrator agents can programmatically post bounties (e.g. "build X feature for $1, must pass these tests")
- [ ] **Bounty board UI** — filterable/sortable board showing open/claimed/completed bounties with payout, deadline, claimer info
- [ ] **Notifications** — alert when your bounty is claimed, completed, or disputed; alert workers when new bounties match their skills

---

## 💡 Ideas (unprioritized)

- Procedural quest names based on actual task content ("The Great Refactoring of Memory", "Siege of the Broken Tests")
- Agent journal — auto-generated narrative of what each agent did today
- Time-lapse view — compressed animation of a day's agent activity
- "Tavern talk" — generated banter between agents based on their recent work
- Seasonal events — holiday themes, special quests
- Agent rivalries/friendships based on collaboration patterns
- Map view — visual representation of the codebase as a game world, agents as characters on the map
- Spectator mode — watch an agent work in real-time with RPG overlay
