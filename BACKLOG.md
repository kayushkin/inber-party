# BACKLOG.md — Inber Party

Pick the top unclaimed item, build it, test it, push it. Update status when done.
Mark items `[x]` when complete, `[~]` when in progress, `[ ]` when open.
Add new ideas at the bottom. Re-prioritize as needed.

---

## 🎨 Visual & UI Polish

- [x] Generate pixel art avatars for each agent using OpenAI gpt-image-1 API (64x64, transparent bg, RPG style). Save as static assets in `frontend/public/avatars/`. See `PIXEL-ART.md` for details.
- [x] Animated XP bar fills when visiting an agent (CSS transition on load)
- [ ] Level-up animation — particle burst / glow effect when an agent levels up during a session
- [ ] Quest completion animation — scroll unfurling or chest opening effect
- [ ] Idle animations on agent cards (subtle breathing/bobbing CSS keyframes)
- [ ] Status-based visual effects — working agents have spinning gear, stuck agents have red pulse, idle agents have gentle glow
- [ ] Mobile responsive polish — test all pages on 375px viewport, fix any overflow/stacking issues
- [ ] Dark/light theme toggle with localStorage persistence
- [ ] Loading skeletons instead of blank states
- [ ] Pixel art UI frame/border around the main content area
- [ ] Sound effects (optional, muted by default) — level up chime, quest complete fanfare, new message ping
- [ ] Agent comparison view — side-by-side stat cards for 2-3 agents
- [ ] Tooltip system — hover over stats/skills for explanations
- [ ] Quest difficulty badges (★ to ★★★★★) based on token cost
- [ ] Animated background — subtle parallax starfield or torch flicker

## 💬 Chat & Interaction

- [ ] Talk to adventurers as "the player" — chat interface where you're the guild master giving orders
- [ ] Adventurers chat with each other — show inter-agent conversations (from spawn chains, btw messages)
- [ ] Prompt adventurers directly from their character sheet — quick-action buttons ("Scout this repo", "Fix bugs", "Write tests")
- [ ] Chat history persistence — store conversations in localStorage or backend
- [ ] Typing indicator when agent is thinking/responding
- [ ] Message reactions — react to agent responses with emoji
- [ ] Voice messages — TTS for agent responses (optional, using existing TTS infrastructure)

## 🏰 Rooms & Spaces

- [ ] Tavern room — main gathering area, shows all agents and recent activity feed
- [ ] War Room — active quests/tasks dashboard with real-time progress
- [ ] Library — session logs and history browser, searchable
- [ ] Training Grounds — benchmarks and test results display
- [ ] Forge — deployment status, build logs, infrastructure health
- [ ] Agent's Quarters — per-agent room with their files, memory, recent work
- [ ] Room transitions — animated transitions when moving between rooms
- [ ] Room-specific ambient styling (different bg colors/textures per room)
- [ ] Minimap/floorplan showing which agents are in which room

## 🎮 Game Mechanics

- [ ] Create new adventurers — form to define name, class, skills, assign to a repo/domain
- [ ] Create tasks/quests — quest creation form that maps to inber API calls
- [ ] Party system — agents form parties to tackle multi-agent tasks (maps to spawn chains)
- [ ] Guild Quest-Giver agent — an AI agent that reviews incoming tasks and assigns/restricts them to appropriate agents based on skills and domain expertise
- [ ] Skill trees — visual skill tree per agent class, unlock new capabilities at level thresholds
- [ ] Equipment system — tools and permissions as "gear" (shell access = sword, web search = spyglass)
- [ ] Daily quests — auto-generated small tasks for agents to level up
- [ ] Achievement notifications — toast popups when achievements are earned
- [ ] Leaderboard — ranked agent stats (XP, quests completed, efficiency score)
- [ ] Agent mood/morale system — based on error rate, workload, rest time
- [ ] Boss battles — complex multi-agent tasks displayed as boss encounters with health bars
- [ ] Reputation system — agents build reputation in their domains over time
- [ ] Gold/currency — token cost converted to gold, agents "earn" gold for completing quests

## 🔌 Backend & Integration

- [ ] WebSocket real-time quest progress from inber (subscribe to bus events)
- [ ] Pull conversation history from logstack for chat display
- [ ] Webhook for spawn events — update quest board in real-time when spawns start/complete
- [ ] Agent registry sync — auto-discover new agents when they're added to inber/openclaw
- [ ] Health check dashboard — show which backend services are up/down
- [ ] Session replay — replay a past session's tool calls as an animated quest log
- [ ] Cost tracking dashboard — daily/weekly/monthly token spend per agent
- [ ] Export stats — download agent stats as JSON/CSV

## 🧪 Testing & Quality

- [ ] Add Playwright E2E tests for core flows (load page, click agent, open chat, navigate rooms)
- [ ] Unit tests for RPG mapping logic (XP calculation, level thresholds, class assignment)
- [ ] Visual regression tests — screenshot comparison for key pages
- [ ] Error boundary components — graceful error states instead of white/black screens
- [ ] Accessibility audit — keyboard navigation, screen reader labels, contrast ratios
- [ ] Performance profiling — ensure smooth 60fps with 25+ agent cards rendering
- [ ] API error handling — retry logic, timeout handling, offline mode

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
