# Party Management — Team Composition & Synergies

## Concept

Some tasks are too big for one agent. **Party Management** lets you assemble teams of agents to tackle complex quests together.

Think:
- **Final Fantasy Tactics** party composition
- **Pokémon team building** with type advantages
- **Overwatch hero synergies**
- **D&D party balance** (tank, healer, DPS)

But for AI coding agents working on the same codebase.

## Core Mechanics

### What is a Party?
A **party** is 2-4 agents assigned to the same quest:

- They work **simultaneously** on different aspects
- They **coordinate** through shared context
- They gain **synergy bonuses** when their skills complement
- They **share XP** from the quest (split or bonus pool)

### Why Form Parties?

**Hard Quests:**
- Tasks requiring multiple skill sets (frontend + backend + testing)
- Large refactors touching many files
- Time-sensitive tasks needing parallel work

**Learning:**
- Junior agent paired with senior for mentorship
- Specialist teaches generalist a new skill

**Efficiency:**
- Two fast agents better than one slow agent
- Parallelization when task has independent subtasks

## Party Composition UI

```
╔══════════════════════════════════════════════════════════╗
║  PARTY BUILDER — "Refactor Auth System"                  ║
╠══════════════════════════════════════════════════════════╣
║                                                           ║
║  Difficulty: ⚔️⚔️⚔️⚔️ Very Hard  |  Est. Time: 8h        ║
║  XP Reward: +300 (split between party)                   ║
║                                                           ║
║  ┌────────────────────────────────────────────────────┐  ║
║  │ PARTY SLOTS (Drag agents here)                     │  ║
║  ├────────────────────────────────────────────────────┤  ║
║  │                                                     │  ║
║  │  [1] ┌───────┐    [2] ┌───────┐    [3] [EMPTY]    │  ║
║  │      │ BRAN  │        │SCÁTHA │                    │  ║
║  │      │ Lv 12 │        │ Lv 15 │                    │  ║
║  │      │Wizard │        │Ranger │                    │  ║
║  │      └───────┘        └───────┘                    │  ║
║  │      TypeScript      Testing                       │  ║
║  │      +15% TS         +20% speed                    │  ║
║  │                                                     │  ║
║  │  Synergy: Bran + Scáthach                          │  ║
║  │  🤝 "Testing Duo" — +15% test coverage            │  ║
║  │  🌟 Worked together 23 times (94% success rate)   │  ║
║  │                                                     │  ║
║  └────────────────────────────────────────────────────┘  ║
║                                                           ║
║  ┌────────────────────────────────────────────────────┐  ║
║  │ AVAILABLE AGENTS                                   │  ║
║  ├────────────────────────────────────────────────────┤  ║
║  │  ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐     │  ║
║  │  │ AOIFE │  │ FIONN │  │DEIRDRE│  │ CIAN  │     │  ║
║  │  │ Lv 9  │  │ Lv 10 │  │ Lv 14 │  │ Lv 7  │     │  ║
║  │  │DevOps │  │ Mage  │  │Fighter│  │Cleric │     │  ║
║  │  └───────┘  └───────┘  └───────┘  └───────┘     │  ║
║  │  (Energy: 85%) (Energy: 60%) (Energy: 90%)        │  ║
║  └────────────────────────────────────────────────────┘  ║
║                                                           ║
║  📊 PARTY STATS                                          ║
║  Combined Success Rate: 95%                              ║
║  Est. Completion Time: 5h 20m (with synergy)             ║
║  XP per Agent: +100 (300 ÷ 3)                           ║
║                                                           ║
║  [Start Quest] [Cancel] [Save as Template]               ║
╚══════════════════════════════════════════════════════════╝
```

## Synergy System

### How Synergies Work

When certain agents work together, they get **bonuses**:

1. **Skill Synergies** — complementary skills boost effectiveness
   - "Frontend + Backend" → +10% integration success
   - "Coder + Tester" → +15% test coverage
   - "DevOps + Backend" → +20% deployment reliability

2. **Relationship Synergies** — agents who've worked together before
   - 10-20 quests together → +5% speed
   - 21-50 quests together → +10% speed
   - 51+ quests together → +15% speed

3. **Level Synergies** — mixing experience levels
   - Junior + Senior → junior gains +50% XP (mentorship)
   - Balanced party (all similar levels) → +5% success rate
   - One high-level carry → -10% XP for others (carried)

4. **Personality Synergies** (future)
   - "Methodical + Bold" → balanced approach
   - "Swift + Wise" → speed + quality
   - "Two Methodical" → slow but thorough

### Example Synergies

```
🤝 Testing Duo (Bran + Scáthach)
  +15% test coverage
  "They always catch edge cases together."

⚡ Speed Team (Scáthach + Aoife)
  +20% completion speed
  "Fast and furious. Don't blink."

🎯 Full Stack (Bran + Fionn + Aoife)
  +10% integration success
  "Frontend, backend, and deploy — covered."

🛡️ Safe Deploy (Aoife + Deirdre)
  +25% deployment reliability
  "They triple-check everything."

🧠 Brain Trust (Fionn + Deirdre)
  +15% code quality
  "Two wise heads, excellent refactors."
```

### Anti-Synergies

Some combos **don't work well**:

```
⚠️ Lone Wolves (Cian + Niamh)
  -10% coordination
  "They prefer working solo."

⚠️ Skill Overlap (Bran + Oisín — both TypeScript specialists)
  -5% efficiency
  "Redundant skills, wasted potential."

⚠️ Energy Mismatch (Tired + Energized)
  -10% speed
  "One's dragging the other down."
```

## Party Roles

Agents can take on roles in a party:

### 1. **Leader** 👑
- Coordinates the party
- Makes final decisions on approach
- Typically the highest level or most experienced

### 2. **Specialist** ⭐
- Focuses on their area of expertise
- e.g., TypeScript specialist handles all .ts files

### 3. **Support** 🛠️
- Assists others (writes tests, fixes lint, reviews code)
- Often a generalist or junior learning from seniors

### 4. **Reviewer** 🔍
- Quality control — reviews changes before commit
- Usually a senior or detail-oriented agent

## Party Quest Execution

### Phase 1: Planning
- Leader analyzes quest
- Splits into subtasks
- Assigns each agent a subtask
- Sets coordination rules (who touches what files)

### Phase 2: Execution
- Agents work **in parallel** on subtasks
- Shared context: all agents see each other's progress
- Coordination: agents avoid merge conflicts by claiming files

### Phase 3: Integration
- Agents merge their work
- Reviewer checks for issues
- Party runs full test suite together

### Phase 4: Completion
- If success → all agents celebrate, XP awarded
- If failure → party regroups, identifies issue, retries

## Battle View (Party Mode)

```
╔══════════════════════════════════════════════════════════╗
║         PARTY QUEST: "Refactor Auth System"              ║
╠══════════════════════════════════════════════════════════╣
║  Task HP: ████████████████░░░░ 75/100                   ║
║  Time: 2h 45m  |  XP: +300 (split)                       ║
╠══════════════════════════════════════════════════════════╣
║                                                           ║
║  [BRAN]      💻 Writing types...     ⚡ Synergy Active   ║
║  Lv 12       HP ████████░  80%                           ║
║  Leader      "Updating auth types"                       ║
║                                                           ║
║  [SCÁTHACH]  🧪 Writing tests...     ⚡ Synergy Active   ║
║  Lv 15       HP ██████████ 100%                          ║
║  Specialist  "Testing login flow"                        ║
║                                                           ║
║  [AOIFE]     🚀 Updating deploy...                       ║
║  Lv 9        HP ████████░  85%                           ║
║  Support     "Configuring env vars"                      ║
║                                                           ║
╠══════════════════════════════════════════════════════════╣
║ PARTY LOG                                                 ║
╠══════════════════════════════════════════════════════════╣
║  💬 Bran: "I'll handle the TypeScript types."            ║
║  💬 Scáthach: "I'll write comprehensive tests."          ║
║  💬 Aoife: "I'll update the deployment config."          ║
║  🤝 Synergy Activated: Testing Duo (+15% coverage)       ║
║  ✏️  Bran casts Write File! (auth.types.ts)              ║
║  🧪 Scáthach uses Test Suite! (auth.test.ts)             ║
║  🔧 Aoife uses Edit Config! (.env.production)            ║
║  ✅ Bran's subtask complete! (types updated)             ║
║  ✅ Scáthach's subtask complete! (tests passing)         ║
║  ⚙️  Aoife still working... (deploying)                  ║
║  ✅ Aoife's subtask complete! (deploy successful)        ║
║  🎉 QUEST COMPLETE! +300 XP (100 each)                   ║
║  🏆 Party achievement: "Full Stack Success"              ║
╚══════════════════════════════════════════════════════════╝
```

## Party Templates

Save successful party compositions for reuse:

```
╔══════════════════════════════════════╗
║ 📋 SAVED PARTY TEMPLATES             ║
╠══════════════════════════════════════╣
║ ⭐ Full Stack Squad                  ║
║    Bran + Scáthach + Aoife           ║
║    Use for: complex features         ║
║    Win rate: 96% (25 quests)         ║
║                                      ║
║ ⚡ Speed Demons                      ║
║    Scáthach + Aoife                  ║
║    Use for: urgent hotfixes          ║
║    Avg time: 1.2h (15 quests)        ║
║                                      ║
║ 🛡️ Safe Deploy Team                 ║
║    Aoife + Deirdre + Fionn           ║
║    Use for: production deploys       ║
║    Error rate: 2% (50 quests)        ║
║                                      ║
║ [Create New Template]                ║
╚══════════════════════════════════════╝
```

## Energy & Rotation

Agents get tired. Party management includes rotation:

### Energy System
- Each agent has **energy** (0-100%)
- Working reduces energy
- Resting restores energy
- Low energy = slower, more errors

### Auto-Rotation
- When an agent's energy drops below 30%, suggest rotation
- Swap in a rested agent mid-quest (if possible)
- Or pause quest to let agents rest

### Best Practices
- Don't overwork the same agents
- Rotate parties to keep synergies fresh
- Balance hard quests with easy quests for energy management

## Why This Works

1. **Tackle big tasks** — parallelization for complex work
2. **Synergies matter** — rewarding team composition strategy
3. **Relationship building** — agents grow together
4. **Mentorship** — junior agents learn from seniors
5. **Fun team dynamics** — feels like managing an RPG party

## Future Enhancements

- **Party chat** — agents communicate during quests
- **Dynamic roles** — agents switch roles mid-quest
- **Rivalry system** — some agents compete for better performance
- **Friendship levels** — unlock unique synergies at high relationship
- **Party achievements** — special rewards for legendary teams
- **Party formations** — defensive (fewer errors) vs. offensive (faster)

---

Great quests need great parties. Choose wisely. Adventure together.
