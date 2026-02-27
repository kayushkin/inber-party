# Progression System — Leveling, Skills & Rewards

## Concept

Agents **grow** as they complete tasks. They gain XP, level up, unlock skills, earn achievements, and climb leaderboards. This makes long-term agent management feel rewarding.

Think:
- **RPG leveling systems** (Final Fantasy, Pokémon, Elder Scrolls)
- **Skill trees** (Diablo, Path of Exile, Borderlands)
- **Battle passes** (Fortnite, Dota 2)
- **Achievement systems** (Steam, Xbox)

But for AI coding agents completing real work.

## Core Loop

```
Complete Task → Gain XP → Level Up → Unlock Skills/Gear → Become More Effective → Take Harder Tasks → Repeat
```

## XP & Leveling

### How XP is Earned

**Base XP** from task difficulty:
- ⚔️ Easy: 20-40 XP
- ⚔️⚔️ Medium: 50-100 XP
- ⚔️⚔️⚔️ Hard: 120-200 XP
- ⚔️⚔️⚔️⚔️ Very Hard: 250-500 XP

**Bonus XP** from performance:
- **Early Finish** (+10-25%): completed before estimate
- **Perfect Execution** (+20%): zero errors, all tests pass
- **Fast Completion** (+15%): finished in record time
- **First Try Success** (+10%): no retries needed
- **Party Bonus** (+5-10%): working in a team

**Penalties:**
- **Over Time** (-5-10%): took longer than estimate
- **Multiple Retries** (-10%): failed and retried
- **User Intervention** (-15%): needed help to complete

### Level Curve

```
Level 1: 0 XP
Level 2: 100 XP
Level 3: 250 XP
Level 4: 450 XP
Level 5: 700 XP
...
Level 10: 3,000 XP
Level 20: 15,000 XP
Level 50: 100,000 XP
Level 100: 500,000 XP
```

Formula: `XP_needed = 50 * level^2`

### What Leveling Does

- **Stats improve** — +1% success rate, +0.5% speed per level
- **Unlock skills** — every 5 levels, unlock a new skill slot
- **Avatar upgrades** — better gear, visual effects
- **Increased XP capacity** — earn more per task
- **Prestige** — bragging rights, leaderboard position

## Skill System

### Skill Categories

**Languages:**
- JavaScript, TypeScript, Python, Go, Rust, etc.

**Frameworks:**
- React, Vue, Angular, Next.js, Express, Django, etc.

**Specializations:**
- Testing (Jest, Pytest, Cypress)
- DevOps (Docker, CI/CD, Kubernetes)
- Databases (SQL, MongoDB, Redis)
- APIs (REST, GraphQL, WebSocket)

**Soft Skills:**
- Debugging (error handling, troubleshooting)
- Refactoring (code quality, cleanup)
- Documentation (writing clear docs)
- Communication (explaining changes)

### Skill Levels

Each skill has 4 levels:

```
Level 1: Novice (10 tasks)
  - Basic familiarity
  - +5% speed in this skill
  - Fewer errors

Level 2: Competent (30 tasks)
  - Solid understanding
  - +10% speed
  - Suggests improvements

Level 3: Expert (75 tasks)
  - Deep knowledge
  - +15% speed
  - Follows best practices
  - Teaches others

Level 4: Master (150 tasks)
  - Peak performance
  - +20% speed
  - Innovates solutions
  - Legendary status
```

### Skill Tree Visualization

```
TypeScript
├─ Level 1: Novice (10 tasks) ✓
├─ Level 2: Competent (30 tasks) ✓
├─ Level 3: Expert (75 tasks) ← 45/75
└─ Level 4: Master (150 tasks) 🔒

React
├─ Level 1: Novice (10 tasks) ✓
└─ Level 2: Competent (30 tasks) ← 12/30

Testing (Jest)
├─ Level 1: Novice (10 tasks) ✓
├─ Level 2: Competent (30 tasks) ✓
├─ Level 3: Expert (75 tasks) ✓
└─ Level 4: Master (150 tasks) ✓ ⭐

DevOps
├─ Level 1: Novice (10 tasks) 🔒
```

### Skill Benefits

- **Speed**: complete tasks faster
- **Quality**: fewer errors, better code
- **Suggestions**: agent proposes optimizations
- **Teaching**: mentors junior agents (XP bonus for both)
- **Specialization**: becomes "the React expert" of your team

## Equipment (Tools)

Agents unlock tools as they progress:

### Starting Equipment (Level 1)
- 📁 File System (read/write files)
- 🔧 Git (basic commits, branches)
- 💻 Code Editor (basic editing)
- 🧪 Testing (run tests)

### Unlockable Tools
- 🌐 **Web Browser** (Level 5) — automate browser, test UIs
- 📡 **API Client** (Level 7) — make HTTP requests
- 🎨 **Image Generation** (Level 10) — create assets with DALL-E
- 🔍 **Web Search** (Level 12) — research solutions
- 🗣️ **Text-to-Speech** (Level 15) — generate audio
- 🤖 **Sub-Agents** (Level 20) — spawn helper agents
- ☁️ **Cloud Access** (Level 25) — deploy to AWS/Azure/GCP
- 🗄️ **Database** (Level 30) — direct DB queries (with permission)

### Equipment Restrictions

Some tools require **admin approval**:
- Production database access
- Deployment keys
- Payment APIs
- Destructive operations (delete files, drop tables)

Agents can **request access** when needed.

## Avatar Progression

As agents level up, their pixel art sprite **evolves**:

### Level 1-9: Novice
- Basic equipment (simple clothes, wooden staff)
- Neutral colors

### Level 10-19: Adept
- Better gear (leather armor, iron sword)
- Slightly glowing aura

### Level 20-29: Expert
- Advanced gear (chainmail, magic staff)
- Glowing effects

### Level 30-49: Master
- Legendary gear (plate armor, enchanted weapon)
- Strong aura, particle effects

### Level 50+: Grandmaster
- Epic gear (glowing armor, mythic weapon)
- Full particle effects, aura, custom animations

### Customization (Future)
- Unlock cosmetic items (hats, capes, colors)
- Choose class appearance (wizard, warrior, ranger)
- Seasonal skins (Halloween, winter, etc.)

## Achievements

Unlock badges for special accomplishments:

### Task-Based
- 🏆 **First Steps** — Complete your first task
- 🏆 **Century Club** — Complete 100 tasks
- 🏆 **Legendary** — Complete 1,000 tasks

### Performance-Based
- 🏆 **Speed Runner** — Task completed in < 30 minutes
- 🏆 **Marathon Runner** — Task completed in > 8 hours
- 🏆 **Perfectionist** — 10 tasks with 100% success, 0 warnings
- 🏆 **Flawless Week** — 7 days without an error

### Skill-Based
- 🏆 **TypeScript Master** — Level 4 in TypeScript
- 🏆 **Polyglot** — Level 2+ in 5 languages
- 🏆 **Jack of All Trades** — Level 1+ in 10 skills

### Team-Based
- 🏆 **Team Player** — 50 party quests completed
- 🏆 **Leader** — Led 20 party quests
- 🏆 **Mentor** — Helped 5 junior agents level up

### Special
- 🏆 **Bug Slayer** — Fixed 50 bugs
- 🏆 **Architect** — Refactored 100+ files in one task
- 🏆 **Clutch** — Fixed critical bug in < 1 hour
- 🏆 **Comeback Kid** — Completed task after 3+ failures

### Secret Achievements
- 🏆 **Ghost** — Completed task with 0 files committed (wat?)
- 🏆 **Overkill** — Used sub-agents for an Easy task
- 🏆 **Night Owl** — Completed task at 3 AM
- 🏆 **Lucky** — Completed task on 1st try with < 10 min

## Leaderboards

Compete for the top spot:

### Overall Ranking
```
╔══════════════════════════════════════╗
║ 🏆 LEADERBOARD (All Time)            ║
╠══════════════════════════════════════╣
║ 1. Scáthach    Level 47   250k XP    ║
║ 2. Bran        Level 42   198k XP    ║
║ 3. Deirdre     Level 38   145k XP    ║
║ 4. Aoife       Level 35   122k XP    ║
║ 5. Fionn       Level 32   98k XP     ║
║ ...                                  ║
╚══════════════════════════════════════╝
```

### Category Leaderboards
- **Most Tasks Completed**
- **Highest Success Rate**
- **Fastest Average Time**
- **Most XP Earned (This Week)**
- **Most Party Quests**
- **Most Achievements**

### Personal Bests
Track individual agent records:
- Fastest task completion
- Longest task (marathon)
- Biggest XP gain (single task)
- Longest success streak

## Prestige System

**What is Prestige?**

At Level 50, agents can **prestige** — reset to Level 1 but keep:
- All unlocked skills (but need to re-level them)
- All achievements
- Avatar customizations
- A **Prestige Star** ⭐ badge

**Prestige Bonuses:**
- +10% XP gain (permanent)
- Unlock exclusive avatar styles
- Access to "Prestige Quests" (special challenges)
- Bragging rights

**Prestige Levels:**
- Prestige 1: ⭐ (1 star)
- Prestige 2: ⭐⭐ (2 stars)
- Prestige 10: ⭐⭐⭐⭐⭐⭐⭐⭐⭐⭐ (legend)

## Daily/Weekly Challenges

Bonus XP for completing objectives:

### Daily Quests
- Complete 3 tasks → +100 XP
- Finish 1 party quest → +150 XP
- Zero errors for 24 hours → +200 XP

### Weekly Quests
- Complete 20 tasks → +500 XP
- Earn 3 achievements → +750 XP
- Level up a skill → +1,000 XP

### Seasonal Events
- **Summer of Code** — 2x XP for all tasks
- **Bug Hunt** — 3x XP for bug fixes
- **Speedrun Week** — Bonus XP for fast completions

## Why This Works

1. **Motivation** — agents (and you) feel progress
2. **Long-term engagement** — reason to keep using the system
3. **Specialization** — agents develop unique strengths
4. **Recognition** — achievements celebrate milestones
5. **Fun** — gamification makes work enjoyable

## Future Enhancements

- **Skill talent trees** — choose specializations (e.g., TypeScript → React or Node.js)
- **Equipment forging** — combine tools to make custom ones
- **Guilds** — teams of users compete globally
- **Trading** — swap agents with other Míl Party users
- **Seasons** — reset leaderboards every 3 months, give rewards
- **Legendary quests** — ultra-hard challenges for max-level agents
- **Cosmetic shop** — spend earned currency on avatar items

---

Every task is XP. Every level is progress. Every agent is a legend in the making.
