# Battle View — Real-Time Task Execution

## Concept

When an agent is working on a task, you can watch it happen in **Battle View** — a live, gamified representation of the agent executing steps.

The task is the "enemy." The agent's actions are "attacks." Errors are "damage taken." Success is "victory."

Think:
- **Final Fantasy turn-based battles** with attack animations
- **Pokémon battle screens** with move names and effects
- **Slay the Spire** combat with cards and actions
- But showing **real dev work** in a fun, informative way

## Visual Layout

```
╔══════════════════════════════════════════════════════════╗
║               BRAN vs. "FIX LOGIN BUG"                   ║
╠══════════════════════════════════════════════════════════╣
║                                                           ║
║   TASK: Fix Login Bug                   TIME: 1h 12m     ║
║   ⚔️⚔️ Medium Difficulty               XP Reward: +50    ║
║   Progress: ████████░░░░░░░░░░ 40%                      ║
║                                                           ║
║   ┌─────────────────┐                 ┌─────────────┐   ║
║   │                 │                 │   [TASK]    │   ║
║   │  [BRAN]         │                 │   🐛 Bug    │   ║
║   │  Level 12       │ →  💻 Writing   │   HP ██████ │   ║
║   │  Wizard         │     Code...     │   60/100    │   ║
║   │                 │                 │             │   ║
║   │  HP ████████░   │                 │  Resistance:│   ║
║   │  Energy 78%     │                 │  Testing: ★ │   ║
║   └─────────────────┘                 └─────────────┘   ║
║                                                           ║
╠══════════════════════════════════════════════════════════╣
║ BATTLE LOG                                                ║
╠══════════════════════════════════════════════════════════╣
║  💬 Bran: "Analyzing the auth flow..."                   ║
║  🔍 Bran uses Read File! (auth.ts)                       ║
║  📖 Bran learns about the session handling.              ║
║  🤔 Bran: "Found it — cookie expiry issue."              ║
║  ✏️  Bran casts Write File! (auth.ts)                    ║
║     → Critical Hit! Bug HP -40                           ║
║  🧪 Bran uses Test Suite!                                ║
║     → Super Effective! All tests pass!                   ║
║     → Bug HP -30                                         ║
║  ⚠️  Bug uses Error! (lint warning)                      ║
║     → Bran takes 5 damage!                               ║
║  🛠️  Bran uses Fix Lint!                                 ║
║     → Bug HP -20                                         ║
║  🎉 Bug defeated! Quest Complete!                        ║
║  🏆 Bran gains +50 XP!                                   ║
╚══════════════════════════════════════════════════════════╝
```

## Elements

### Agent Panel (Left)
- **Avatar** — pixel art sprite, animated based on action
- **Name & Level**
- **HP Bar** — agent's "health" (errors reduce it)
  - If HP reaches 0, task fails and needs intervention
- **Energy** — stamina (drains slowly, affects speed)
- **Status effects** — buffs (TypeScript Specialist active!) or debuffs (rate limited, waiting...)

### Task Panel (Right)
- **Task name** — the "enemy"
- **Difficulty** — represented as enemy strength/level
- **HP Bar** — task "health" representing completeness
  - Each successful action reduces task HP
  - When HP = 0, task is complete
- **Resistances** — task characteristics:
  - Testing: ★★★ (needs lots of tests)
  - Complexity: ★★ (moderate difficulty)
  - Urgency: ★ (not time-sensitive)

### Progress Bar (Top)
- Overall task completion percentage
- Time elapsed
- XP reward

### Battle Log (Bottom)
Scrollable action feed showing what's happening in real-time:

#### Action Format
```
[Icon] [Agent] [Action] ([Target/Details])
→ [Effect/Result]
```

#### Example Actions

**Agent Actions:**
- 🔍 **Read File** — agent examines code
- ✏️ **Write File** — agent edits code
- 🧪 **Run Tests** — agent executes test suite
- 🐛 **Debug** — agent investigates error
- 🔧 **Refactor** — agent improves structure
- 📦 **Install Package** — agent adds dependency
- 🌐 **Search Web** — agent looks up docs
- 💬 **Comment** — agent explains reasoning
- 🚀 **Deploy** — agent pushes changes

**Task/Error Actions:**
- ⚠️ **Error** — task fights back (syntax error, test failure)
- ⏸️ **Wait** — external dependency (API response, user input)
- 🛡️ **Resist** — task is harder than expected
- 💥 **Critical Error** — major blocker

**Results:**
- **Critical Hit!** — perfect action, big progress
- **Super Effective!** — action well-suited to task (e.g., TypeScript specialist editing .ts file)
- **Miss!** — action didn't help (wrong approach)
- **Resisted!** — task pushes back
- **Victory!** — task complete

## Animations

### Agent Animations
- **Idle** — gentle sway, breathing
- **Attacking** — strike/cast motion (matches action type)
- **Damaged** — recoil, flash red
- **Victorious** — jump/cheer
- **Defeated** — slump, question mark

### Task Animations
- **Idle** — ominous presence, subtle movement
- **Damaged** — flash white, shake
- **Defeated** — dissolve/fade with particle effects

### Special Effects
- **Critical Hit** — screen flash, bold text
- **Super Effective** — glowing aura around agent
- **Level Up** — burst of light, fanfare sound

## Interactions

### Pause/Resume
- **Pause button** — freeze the task
- Useful for inspecting current state or intervening

### Intervene
- **Chat button** — send message to agent
  - "Try using library X instead"
  - "Skip the deployment step for now"
- Agent reads message and adjusts approach

### Cancel
- **Stop button** — abort the task
- Moves quest to Failed column
- Agent returns to idle

### Speed Controls
- **1x / 2x / 5x** — playback speed
- Or **instant** — skip to result (for impatient users)

## Status Effects

### Buffs (Positive)
- 🌟 **Specialist Bonus** — agent using a mastered skill
- ⚡ **Energized** — agent just rested, +20% speed
- 🤝 **Party Assist** — another agent helping in background
- 🎯 **Focused** — agent in "the zone," +accuracy

### Debuffs (Negative)
- 😴 **Tired** — low energy, -20% speed
- ⏳ **Rate Limited** — API throttle, must wait
- 🔥 **On Fire** — multiple errors, -accuracy
- 🧊 **Blocked** — waiting for external input

## Detailed Battle Example

### Scenario: Bran Fixes a Login Bug

```
00:00 — Quest Start
💬 Bran: "Let's see what we're dealing with."
🔍 Bran uses Read File! (pages/login.tsx)
📖 Bran learns about the login form structure.
🔍 Bran uses Read File! (lib/auth.ts)
📖 Bran discovers the authentication logic.

00:03 — Analysis Phase
🤔 Bran: "The session cookie expires too quickly."
💬 Bran: "I'll update the maxAge setting."

00:05 — Implementation
✏️ Bran casts Write File! (lib/auth.ts)
→ Critical Hit! Bug HP -40 (60/100 remaining)
📝 Bran modified 3 lines of code.

00:07 — Testing
🧪 Bran uses Test Suite!
⏳ Running tests...
✅ All 12 tests pass!
→ Super Effective! Bug HP -30 (30/100 remaining)

00:10 — Linting
⚠️ Bug uses Error! (ESLint: missing semicolon)
→ Bran takes 5 damage! (HP: 95/100)
🛠️ Bran uses Fix Lint!
→ Bug HP -10 (20/100 remaining)

00:12 — Final Check
🔍 Bran uses Read File! (double-check changes)
✅ Looks good.
🧪 Bran uses Test Suite! (one more time)
✅ Perfect score!
→ Bug HP -20 (0/100 remaining)

00:14 — Victory!
🎉 Bug defeated! Quest Complete!
🏆 Bran gains +50 XP!
⭐ Bran leveled up! (Level 12 → 13)
💬 Bran: "Cookie expiry fixed. Users should stay logged in now."
```

## Why This Works

1. **Entertaining** — watching work happen is actually fun
2. **Informative** — see exactly what the agent is doing
3. **Real-time** — live updates, not post-mortem logs
4. **Gamified** — feels like a game, but represents real work
5. **Transparent** — builds trust (you see every step)

## Future Enhancements

- **Replays** — watch completed battles after the fact
- **Spectator mode** — watch multiple agents battle at once
- **Power-ups** — give agent temporary boosts mid-battle
- **Boss battles** — extra-hard tasks with special mechanics
- **Co-op battles** — multiple agents working together, synchronized actions
- **Battle stats** — detailed breakdown of time per action, efficiency score
- **Commentary mode** — optional narrator (like Pokémon Stadium)

---

Work is battle. Make it epic.
