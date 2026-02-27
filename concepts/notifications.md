# Notifications — Agent Attention System

## Concept

Agents need to get your attention when something important happens. Instead of boring alerts, they use **in-character notifications** that fit the RPG theme.

Think:
- **Speech bubbles in games** where NPCs talk to you
- **Exclamation marks above heads** when someone needs help
- **Quest markers** showing where to go next
- But for agent status updates

## Notification Priorities

### 1. **Casual Wave** 💬
- **When**: Agent finished a task, leveled up, or has info to share
- **Visual**: Agent waves gently, speech bubble appears
- **Sound**: Soft chime
- **Action**: Optional — check when convenient
- **Example**: "Hey! I finished that refactor. Want to review?"

### 2. **Urgent Flag** 🚩
- **When**: Agent is stuck but not blocked (needs guidance, found issue)
- **Visual**: Agent waves faster, yellow/orange flag above head, pulsing
- **Sound**: Alert beep (repeats every 30 seconds)
- **Action**: Check soon (within a few minutes)
- **Example**: "I found 3 ways to implement this. Which approach do you prefer?"

### 3. **Critical SOS** 🆘
- **When**: Agent is blocked and can't proceed (critical error, permission needed)
- **Visual**: Agent frantically waving, red SOS beacon, screen edge pulses red
- **Sound**: Urgent alarm (repeats every 10 seconds)
- **Action**: Check immediately
- **Example**: "Error: Database connection failed! I need your help!"

## Visual Styles

### Speech Bubbles
Appear above agent avatar in camp view:

```
     ╭─────────────────────╮
     │ Task done! +50 XP! │
     ╰─────┬───────────────╯
           │
        [ BRAN ]
        (waving)
```

**Variations:**
- **Info** (💬): white bubble, normal text
- **Success** (🎉): green bubble, celebratory text
- **Question** (❓): yellow bubble, italic text
- **Error** (⚠️): red bubble, bold text

### Icons & Badges

**Above Agent Avatar:**
- 💬 — has message
- ❓ — needs decision
- 🚩 — stuck
- 🆘 — blocked
- 🎉 — celebrating
- ⏸️ — paused/waiting
- 💤 — resting

**Badge Count:**
If multiple notifications, show number:
```
  ┌─3──┐
  │ 🚩 │
  └────┘
```

### Screen Edge Pulses

For critical notifications, the entire UI draws attention:

- **Yellow pulse** (urgent) — edge of screen glows yellow briefly
- **Red pulse** (critical) — edge of screen pulses red continuously

## Notification Panel

Sidebar or dropdown showing all notifications:

```
╔══════════════════════════════════════╗
║ 🔔 NOTIFICATIONS               [5]   ║
╠══════════════════════════════════════╣
║ 🆘 CRITICAL                          ║
║  ┌────────────────────────────────┐ ║
║  │ Bran — Database error          │ ║
║  │ "Can't connect to prod DB!"    │ ║
║  │ 2 minutes ago                  │ ║
║  └────────────────────────────────┘ ║
║                                      ║
║ 🚩 URGENT                            ║
║  ┌────────────────────────────────┐ ║
║  │ Scáthach — Needs decision      │ ║
║  │ "Use axios or fetch?"          │ ║
║  │ 5 minutes ago                  │ ║
║  └────────────────────────────────┘ ║
║                                      ║
║ 💬 INFO                              ║
║  ┌────────────────────────────────┐ ║
║  │ Aoife — Task complete          │ ║
║  │ "Tests passing! +45 XP"        │ ║
║  │ 10 minutes ago                 │ ║
║  └────────────────────────────────┘ ║
║                                      ║
║  ┌────────────────────────────────┐ ║
║  │ Fionn — Leveled up!            │ ║
║  │ "Now level 10! 🎉"             │ ║
║  │ 15 minutes ago                 │ ║
║  └────────────────────────────────┘ ║
║                                      ║
║  ┌────────────────────────────────┐ ║
║  │ Deirdre — Found optimization   │ ║
║  │ "Can reduce bundle size 20%"   │ ║
║  │ 30 minutes ago                 │ ║
║  └────────────────────────────────┘ ║
║                                      ║
║ [Clear All] [Settings]               ║
╚══════════════════════════════════════╝
```

## Notification Types

### 1. **Task Complete** ✅
```
💬 Bran: "Bug fixed! All tests pass. +50 XP"
[View Changes] [Dismiss]
```

### 2. **Task Failed** ❌
```
⚠️ Scáthach: "Deploy failed — permission denied on server."
[View Error] [Retry] [Cancel Task]
```

### 3. **Stuck/Need Help** 🚩
```
❓ Aoife: "Found 3 solutions for the API issue. Which should I use?"
[Option A] [Option B] [Option C] [Your Call]
```

### 4. **Blocked/Critical** 🆘
```
🆘 Fionn: "CRITICAL: Production database is unreachable!"
[View Details] [Pause All] [Contact DevOps]
```

### 5. **Level Up** 🎉
```
🎉 Deirdre: "Level 15 reached! Unlocked 'React Expert' skill!"
[View Character Sheet] [Nice!]
```

### 6. **Achievement Unlocked** 🏆
```
🏆 Bran: "Achievement unlocked: Bug Slayer (50 bugs fixed)!"
[View Achievements] [Cool!]
```

### 7. **Low Energy** 😴
```
💤 Scáthach: "I'm getting tired... energy at 20%."
[Rest Now] [Push Through] [Dismiss]
```

### 8. **Quest Available** 📋
```
💬 System: "New quest posted: 'Refactor auth system' (⚔️⚔️⚔️ Hard)"
[View Quest] [Auto-Assign] [Dismiss]
```

### 9. **Party Synergy** 🤝
```
💡 System: "Bran + Scáthach haven't worked together lately. Their synergy is strong!"
[Assign Party Quest] [Noted]
```

## Sound Design

### Sound Levels
- **Muted** — no sound, visual only
- **Subtle** — soft chimes and beeps
- **Full** — RPG-style sound effects
- **Voice** (future) — TTS for agent messages

### Example Sounds
- **Task complete**: *ding!* (victory chime)
- **Level up**: *fanfare!* (trumpets)
- **Error**: *buzzer* (wrong answer sound)
- **Urgent**: *beep beep* (alarm)
- **Critical**: *ALARM!* (loud, persistent)
- **Agent wave**: *hey!* (friendly attention sound)

## Notification Settings

User preferences:

```
╔══════════════════════════════════════╗
║ 🔔 NOTIFICATION SETTINGS             ║
╠══════════════════════════════════════╣
║ Enable Notifications: [✓]            ║
║                                      ║
║ Sound:                               ║
║   ○ Muted                            ║
║   ● Subtle                           ║
║   ○ Full                             ║
║   ○ Voice (TTS)                      ║
║                                      ║
║ Show notifications for:              ║
║   [✓] Task complete                  ║
║   [✓] Task failed                    ║
║   [✓] Agent stuck                    ║
║   [✓] Agent blocked                  ║
║   [✓] Level up                       ║
║   [✓] Achievement unlocked           ║
║   [ ] Low energy                     ║
║   [ ] Quest available                ║
║                                      ║
║ Notification persistence:            ║
║   Critical: Until dismissed          ║
║   Urgent: 5 minutes                  ║
║   Info: 1 minute                     ║
║                                      ║
║ Do Not Disturb: [Off]                ║
║   (mute all notifications)           ║
║                                      ║
║ [Save Settings]                      ║
╚══════════════════════════════════════╝
```

## Notification Actions

Users can respond directly from notifications:

### Quick Actions
- **Dismiss** — close notification
- **View Details** — jump to agent/quest
- **Pause Task** — stop agent temporarily
- **Cancel Task** — abort quest
- **Retry** — restart failed task
- **Assign to Another Agent** — reassign stuck task

### Contextual Actions
Based on notification type:
- **Make Decision** — choose from options (for stuck agents)
- **Approve Permission** — grant access (for blocked agents)
- **View Changes** — see diff (for completed tasks)
- **View Error** — see logs (for failed tasks)
- **Rest Agent** — put on break (for tired agents)

## Desktop Notifications

For users with the app minimized:

```
╔══════════════════════════════════════╗
║ 🆘 Míl Party — Bran Blocked          ║
╠══════════════════════════════════════╣
║ "Database connection failed!"        ║
║ Click to view details                ║
╚══════════════════════════════════════╝
```

**Priorities:**
- Critical → always show desktop notification
- Urgent → show if app inactive > 2 minutes
- Info → only show if enabled in settings

## Mobile Notifications

For mobile/remote monitoring:

- Push notifications via web push API
- Summary: "3 agents need attention"
- Tap to open Míl Party

## Notification History

Keep a log of past notifications:

```
╔══════════════════════════════════════╗
║ 📜 NOTIFICATION HISTORY              ║
╠══════════════════════════════════════╣
║ Today                                ║
║  🆘 Bran — Database error (14:23)    ║
║  🚩 Scáthach — Stuck (13:45)         ║
║  🎉 Aoife — Level up! (12:30)        ║
║  ✅ Fionn — Task done (11:15)        ║
║                                      ║
║ Yesterday                            ║
║  ✅ Bran — Task done (18:45)         ║
║  ❌ Deirdre — Deploy failed (16:20)  ║
║  🏆 Scáthach — Achievement (15:00)   ║
║                                      ║
║ [Load More]                          ║
╚══════════════════════════════════════╝
```

## Why This Works

1. **Prioritized attention** — critical issues can't be missed
2. **In-character** — fits the RPG theme
3. **Actionable** — respond directly from notification
4. **Non-intrusive** — casual notifications don't interrupt flow
5. **History** — never lose track of what happened

## Future Enhancements

- **Custom notification styles** — user-designed speech bubbles
- **Agent personalities** — notifications match agent's tone
- **Notification grouping** — "3 agents need your attention"
- **Smart batching** — combine similar notifications
- **Learning** — system learns which notifications you care about
- **Integrations** — send notifications to Slack, email, SMS

---

Your Míls will get your attention. You just have to listen.
