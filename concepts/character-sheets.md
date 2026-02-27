# Character Sheets — Agent Stats & History

## Concept

Every agent (Míl) has a **Character Sheet** — a detailed view of their stats, history, skills, and progression. This is where you see what makes each agent unique.

Think:
- **D&D character sheets** with stats and abilities
- **Final Fantasy job/license boards**
- **Pokémon stat screens** with move history
- **GitHub profile pages** but for AI agents

## Visual Layout

```
╔══════════════════════════════════════════════════════════╗
║  [Back to Camp]         BRAN THE METHODICAL             ║
╠══════════════════════════════════════════════════════════╣
║                                                           ║
║   ┌────────┐   Level 12                  ⚡ Energy: 78%  ║
║   │ 64x64  │   Wizard (Coder)           ════════░░       ║
║   │ PIXEL  │   XP: 4,820 / 5,000                        ║
║   │  ART   │   ████████████░ (96%)                      ║
║   │ SPRITE │                                             ║
║   └────────┘   "Methodical, thorough, never rushes."    ║
║                                                           ║
╠══════════════════════════════════════════════════════════╣
║ STATS                                                     ║
╠══════════════════════════════════════════════════════════╣
║  📊 Tasks Completed: 127        ⏱️ Avg Time: 2.3h       ║
║  ✅ Success Rate: 94%           💻 Lines of Code: 8,432  ║
║  🧪 Tests Written: 234          🐛 Bugs Fixed: 89        ║
║  📝 Files Modified: 412         ⚠️ Errors: 8             ║
╠══════════════════════════════════════════════════════════╣
║ SKILLS & SPECIALIZATIONS                                 ║
╠══════════════════════════════════════════════════════════╣
║  ⭐ TypeScript Specialist (Lvl 3) — +15% speed on .ts    ║
║  ⭐ Test Master (Lvl 2) — writes better test coverage    ║
║  ⭐ Git Wizard (Lvl 1) — handles merge conflicts         ║
║  🔒 React Expert (unlock at 50 React tasks)              ║
║  🔒 DevOps Pro (unlock at 30 deployment tasks)           ║
╠══════════════════════════════════════════════════════════╣
║ EQUIPMENT (Tools)                                        ║
╠══════════════════════════════════════════════════════════╣
║  🛠️ VSCode       🔧 ESLint      🧪 Jest                  ║
║  📦 npm          🐙 git         🌐 browser               ║
║  📡 API calls    📁 file system 🔍 web search            ║
╠══════════════════════════════════════════════════════════╣
║ QUEST LOG (Recent)                                       ║
╠══════════════════════════════════════════════════════════╣
║  ✅ Fix login bug · 2h 14m · +50 XP · 2 hours ago       ║
║  ✅ Add dark mode · 3h 45m · +80 XP · 5 hours ago       ║
║  ❌ Deploy to prod · FAILED (timeout) · 1 day ago        ║
║  ✅ Write tests for auth · 1h 30m · +45 XP · 2 days ago ║
║  [View Full History]                                     ║
╠══════════════════════════════════════════════════════════╣
║ ACHIEVEMENTS                                              ║
╠══════════════════════════════════════════════════════════╣
║  🏆 Century Club (100 tasks)                             ║
║  🏆 Flawless Week (7 days, 0 errors)                     ║
║  🏆 Speed Runner (task in < 30 min)                      ║
║  🏆 Bug Slayer (50 bugs fixed)                           ║
║  [View All 12 Achievements]                              ║
╠══════════════════════════════════════════════════════════╣
║ RELATIONSHIPS (Party Synergies)                          ║
╠══════════════════════════════════════════════════════════╣
║  💚 Scáthach the Swift — 23 quests together · +12% speed ║
║  💙 Aoife the Bold — 15 quests together · +8% accuracy   ║
║  💛 Fionn the Wise — 8 quests together · +5% XP          ║
╚══════════════════════════════════════════════════════════╝
```

## Sections

### 1. **Header**
- **Avatar** — large pixel art sprite (128x128 or 64x64)
- **Name** — agent's name + title/personality
- **Level & Class** — e.g., "Level 12 Wizard (Coder)"
- **XP Bar** — progress to next level
- **Energy** — current stamina (low energy = needs rest)
- **Quote** — personality flavor text

### 2. **Stats**
Raw metrics from task history:

- **Tasks Completed** — total quests finished
- **Success Rate** — % of tasks completed without errors
- **Avg Time** — mean time per task
- **Lines of Code** — total lines written/modified
- **Tests Written** — count of test cases added
- **Bugs Fixed** — count of fixes
- **Files Modified** — total files touched
- **Errors** — count of failures/errors encountered

### 3. **Skills & Specializations**
Unlocked through experience:

#### How Skills Work
- Agent completes tasks involving specific technologies (TypeScript, React, Testing, etc.)
- After X tasks in a category, unlock Skill Level 1
- Higher levels = better performance (faster, fewer errors, better quality)

#### Example Skill Tree
```
TypeScript Specialist
├─ Level 1 (10 TS tasks): +5% speed
├─ Level 2 (30 TS tasks): +10% speed, fewer type errors
├─ Level 3 (75 TS tasks): +15% speed, suggests best practices
└─ Level 4 (150 TS tasks): +20% speed, auto-refactors

Test Master
├─ Level 1 (20 test tasks): writes basic tests
├─ Level 2 (50 test tasks): better coverage, edge cases
├─ Level 3 (100 test tasks): generates test suites automatically
└─ Level 4 (200 test tasks): finds untested code, suggests tests

Git Wizard
├─ Level 1 (15 git tasks): handles commits, branches
├─ Level 2 (40 git tasks): resolves merge conflicts
├─ Level 3 (80 git tasks): optimizes branch strategy
└─ Level 4 (150 git tasks): rebases like a pro

React Expert, DevOps Pro, API Architect, etc.
```

### 4. **Equipment (Tools)**
What tools/APIs the agent has access to:

- **Built-in**: file system, git, terminal
- **Unlocked**: browser automation, image generation, TTS, etc.
- **Restricted**: production database, deployment keys (require permission)

Visual: icons for each tool, grayed out if locked

### 5. **Quest Log**
Scrollable history of completed/failed tasks:

```
✅ Task Name · Duration · +XP · Time Ago
❌ Task Name · FAILED (reason) · Time Ago
```

Click any quest to see:
- Full description
- Steps taken (log summary)
- Files changed
- Output/errors
- Replay session (if available)

### 6. **Achievements**
Gamification badges:

- **Century Club** — 100 tasks completed
- **Speed Runner** — task in < 30 minutes
- **Marathon** — task taking > 8 hours
- **Flawless Week** — 7 days with 0 errors
- **Bug Slayer** — 50 bugs fixed
- **Test Champion** — 500 tests written
- **Early Bird** — task completed before estimate
- **Clutch Player** — critical bug fixed in < 1 hour
- **Team Player** — 50 party quests
- **Perfectionist** — 10 tasks with 100% success + 0 warnings

### 7. **Relationships (Party Synergies)**
Track which agents work well together:

- **Pairs** that have completed tasks together
- **Synergy bonus** — "Bran + Scáthach: +12% speed"
- **Quest count** — how many times they've teamed up
- **Success rate** — win rate when working together

Use this to form optimal parties for hard quests.

## Interactions

### Hover Stats
Tooltips with more context:
- "Success Rate: 94% — 119 successes, 8 failures"
- "Lines of Code: 8,432 — ~66 per task"

### Click Quest Log Entry
Opens detailed quest report (like a mini battle view replay)

### Click Skill
Shows skill tree and progress:
```
TypeScript Specialist (Level 3)
Progress: 75 / 150 tasks → Level 4
Next Unlock: +20% speed, auto-refactoring
```

### Click Achievement
Shows achievement description + unlock date + rarity

### Edit Button (Optional)
Admins can:
- Rename agent
- Change avatar
- Adjust personality quote
- Reset stats (for testing)

## Comparison View

View multiple character sheets side-by-side to compare agents:

```
╔═══════════════════╦═══════════════════╦═══════════════════╗
║      BRAN         ║    SCÁTHACH       ║      AOIFE        ║
╠═══════════════════╬═══════════════════╬═══════════════════╣
║ Level 12          ║ Level 15          ║ Level 9           ║
║ 127 tasks (94%)   ║ 203 tasks (97%)   ║ 68 tasks (88%)    ║
║ 2.3h avg          ║ 1.8h avg          ║ 3.1h avg          ║
║ TypeScript Spec   ║ Speed Demon       ║ DevOps Pro        ║
╚═══════════════════╩═══════════════════╩═══════════════════╝
```

Useful for deciding who to assign a quest to.

## Why This Works

1. **Personality** — each agent feels unique
2. **Transparency** — see exactly what they've done
3. **Trust** — success rate helps you assign critical tasks
4. **Progression** — watch agents grow over time
5. **Strategy** — use stats to optimize task assignment

## Future Enhancements

- **Skill Respec** — reset skill tree to try different builds
- **Prestige System** — reset level for permanent bonuses
- **Custom Attributes** — user-defined stats (e.g., "code quality score")
- **Personality AI** — agent behavior adapts based on history
- **Trading Cards** — export agent stats as shareable cards
- **Hall of Fame** — retired agents with legendary stats

---

Know your Míls. They're not just processes — they're your party.
