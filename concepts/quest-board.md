# Quest Board — Task Management

## Concept

The **Quest Board** is where tasks become quests. Instead of a sterile task queue, you get a medieval bulletin board covered in quest cards.

Think:
- **Notice boards in RPG towns** where you pick up side quests
- **Monster Hunter quest counter** with difficulty ratings
- **Trello/Kanban boards** but with personality

## Visual Layout

```
╔══════════════════════════════════════════════════════╗
║             🗡️  QUEST BOARD  🗡️                     ║
╠══════════════════════════════════════════════════════╣
║                                                       ║
║  [AVAILABLE]     [IN PROGRESS]     [COMPLETED]       ║
║                                                       ║
║  ┌─────────┐     ┌─────────┐       ┌─────────┐      ║
║  │ Fix Bug │     │ Deploy  │       │ Tests   │      ║
║  │ ⚔️ Med  │     │ ⚔️⚔️ Hard│       │ ⚔️ Easy │      ║
║  │ +50 XP  │     │ +120 XP │       │ +30 XP  │ ✓    ║
║  │ ~2h     │     │ Bran    │       │ Scátha  │      ║
║  └─────────┘     └─────────┘       └─────────┘      ║
║                                                       ║
║  ┌─────────┐     ┌─────────┐       ┌─────────┐      ║
║  │ Refactor│     │ Feature │       │ Docs    │      ║
║  │ ⚔️⚔️⚔️ V.Hard│ ⚔️⚔️ Hard│       │ ⚔️ Easy │      ║
║  │ +200 XP │     │ +150 XP │       │ +20 XP  │ ✓    ║
║  │ ~8h     │     │ Aoife   │       │ Fionn   │      ║
║  └─────────┘     └─────────┘       └─────────┘      ║
║                                                       ║
║  ┌─────────┐                        [FAILED]         ║
║  │ New Repo│                                         ║
║  │ ⚔️ Med  │                        ┌─────────┐      ║
║  │ +80 XP  │                        │ Hotfix  │      ║
║  │ ~3h     │                        │ ⚔️⚔️ Hard│ ✗    ║
║  └─────────┘                        │ Deirdre │      ║
║                                     └─────────┘      ║
╚══════════════════════════════════════════════════════╝
```

## Quest Card Components

Each quest card is a small panel with:

### Header
- **Quest name** (the task description, shortened)
- **Priority icon** (⭐ urgent, ⏰ deadline, 💡 idea)

### Difficulty
- **Swords** represent difficulty:
  - ⚔️ Easy (< 1 hour, low risk)
  - ⚔️⚔️ Medium (1-4 hours, moderate complexity)
  - ⚔️⚔️⚔️ Hard (4-8 hours, high complexity)
  - ⚔️⚔️⚔️⚔️ Very Hard (8+ hours, critical/risky)
- Auto-calculated based on estimated time + keywords (refactor, deploy, database, etc.)

### Rewards
- **+X XP** — how much experience the agent gains
- Higher difficulty = more XP
- Bonus XP for completing ahead of time or with zero errors

### Time Estimate
- **~Xh** — estimated completion time
- Updates in real-time as the agent works
- Shows elapsed time for in-progress quests

### Agent Assignment
- **Agent avatar** (small pixel sprite) + name
- Drag agents onto cards to assign
- Or click card → "Assign to..." menu

### Tags/Skills
- **Labels** showing required skills: `TypeScript`, `Testing`, `DevOps`, etc.
- Agents with matching skills highlighted in assignment menu

## Columns

### 📋 Available
- Tasks that **haven't been started**
- Sorted by priority (urgent → normal → low)
- Agents can "claim" tasks (if auto-assign mode is on)

### ⚙️ In Progress
- Tasks **currently being worked on**
- Shows agent avatar, progress bar, elapsed time
- Live updates from the agent's session

### ✅ Completed
- Tasks **successfully finished**
- Shows completion time, XP awarded
- Celebrating animation when first moved here
- Auto-archives after 24 hours (or keep for review)

### ❌ Failed
- Tasks that **errored out** or were abandoned
- Shows failure reason (error message, timeout, user canceled)
- Can be retried (sends back to Available)

## Interactions

### Click a Quest Card
Opens **quest details panel**:

```
╔═══════════════════════════════════════╗
║ QUEST: Fix Login Bug                  ║
╠═══════════════════════════════════════╣
║ Status: In Progress                   ║
║ Assigned to: Bran the Methodical      ║
║ Started: 14:23                        ║
║ Elapsed: 1h 12m                       ║
║                                        ║
║ Description:                          ║
║ Users report login errors on Safari.  ║
║ Check auth flow and session handling. ║
║                                        ║
║ Steps Completed:                      ║
║ ✓ Reproduce bug                       ║
║ ✓ Identify root cause (cookie issue)  ║
║ ⚙️ Write fix                          ║
║ ⏳ Test fix                           ║
║ ⏳ Deploy                             ║
║                                        ║
║ [View Battle] [Cancel] [Reassign]     ║
╚═══════════════════════════════════════╝
```

### Drag & Drop
- **Drag quest card** onto an agent in Camp View → assigns task
- **Drag quest card** between columns (if manual mode)
- **Drag multiple agents** onto a hard quest → creates a party quest

### Create New Quest
- **+ button** at top of Available column
- Opens form:
  - Quest name (required)
  - Description (markdown supported)
  - Estimated time (optional, auto-calculates difficulty)
  - Skills required (tags)
  - Priority (normal, high, urgent)
  - Auto-assign? (let agents claim it vs. manual assignment)

### Filters & Search
- **Filter by**:
  - Difficulty (easy, medium, hard, very hard)
  - Skills (TypeScript, Testing, etc.)
  - Agent (show only Bran's quests)
  - Date range
- **Search**: fuzzy search across quest names and descriptions

## Quest States (Detailed)

### Available → In Progress
- User assigns quest to agent (drag or menu)
- OR agent auto-claims quest (if enabled)
- Card moves to "In Progress" column
- Agent state changes to "Working" or "On Quest"

### In Progress → Completed
- Agent finishes successfully
- Card moves to "Completed" with ✓
- XP awarded, agent celebrates
- If agent leveled up, show level-up notification

### In Progress → Failed
- Agent encounters error it can't recover from
- OR user cancels task
- OR timeout exceeded
- Card moves to "Failed" with ✗ and error summary

### Failed → Available (Retry)
- User clicks "Retry" on failed quest
- Card moves back to Available
- Can be assigned to same or different agent

### Completed → Archive
- After 24 hours (configurable)
- OR user manually archives
- Removed from board, stored in quest history
- Still visible in agent character sheets

## Auto-Assignment Modes

### Manual
- User assigns all quests by drag/drop or menu
- Full control, good for critical tasks

### Auto-Claim
- Available quests shown to idle agents
- Agents "claim" quests matching their skills
- First-come, first-served

### Smart Assignment
- System suggests best agent for each quest
- Based on: skills, success rate, current workload, energy level
- User can accept or override

### Party Mode
- Hard quests automatically look for 2+ agents with complementary skills
- User confirms party composition

## Notifications

- **New quest posted** → notify idle agents
- **Quest completed** → celebrate + show XP
- **Quest failed** → alert user, highlight card
- **Quest stuck** → agent waves from battle view, card pulses

## Why This Works

1. **Familiar metaphor** — everyone knows quest boards
2. **Visual task management** — see status at a glance
3. **Gamification** — tasks feel like quests, not chores
4. **Flexibility** — supports manual, auto, and hybrid workflows
5. **Progress visibility** — know exactly what's happening

## Future Enhancements

- **Quest chains** — completing Quest A unlocks Quest B
- **Daily/weekly quests** — recurring tasks for XP farming
- **Bounties** — high-reward quests with bonuses
- **Community board** — share quests with other Míl Party users
- **Quest templates** — save common task patterns
- **Voice of the quest giver** — flavor text when posting quests

---

Every task is an adventure. Post it. Watch it happen. Celebrate the win.
