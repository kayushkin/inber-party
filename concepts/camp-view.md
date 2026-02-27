# Camp View — The Agent Hub

## Concept

The **Camp View** is the main screen you see when you open Míl Party. It's where your agents "live" when they're not out on quests.

Think of it like:
- The **camp scenes in Chrono Trigger** where your party rests between adventures
- The **base camp in Fire Emblem** where you manage your units
- The **guild hall in a classic RPG** where adventurers gather

## Visual Layout

```
┌────────────────────────────────────────────────────────┐
│  🌲  MÍLS' CAMP  🌲                    [Menu] [Quest] │
├────────────────────────────────────────────────────────┤
│                                                         │
│     🌲          🔥                        🌲           │
│                ╱│╲                                     │
│               ╱ │ ╲                                    │
│    [Bran]    ┴─┴─┴    [Scáthach]        [Aoife]      │
│      💤       FIRE       👋              ⚔️ AWAY      │
│     idle    crackling   stuck           on quest      │
│                                                         │
│   [Fionn]             [Deirdre]          🪨            │
│    💻 BUSY            💤 idle           stone          │
│   working                                              │
│                                                         │
│  🌲              🌲             🌲                     │
└────────────────────────────────────────────────────────┘
```

## Agent States

Each agent (Míl) in the camp is rendered as a pixel-art sprite (64x64) in one of several states:

### 1. **Idle** 💤
- Agent is **available** for tasks
- Animation: gentle breathing, occasional blink, idle sway
- Position: sitting by the fire or relaxing around camp
- Color: normal, warm lighting

### 2. **Working** 💻
- Agent is **actively executing a task**
- Animation: typing gesture, tool-wielding, focused
- Position: at a workbench or desk area in camp
- Color: slight glow/aura effect
- Shows a progress indicator or activity icon

### 3. **On Quest** ⚔️
- Agent is **away on a task** (not visible in camp)
- Represented by: an empty spot with a "trail" or departure marker
- Shows: agent name, quest name, time elapsed
- Click to see quest details or switch to battle view

### 4. **Stuck** 👋 / 🚩
- Agent **needs help** (error, waiting for input, blocked)
- Animation: **waving frantically**, flashing exclamation mark
- Color: pulsing red/orange glow
- Speech bubble: "I'm stuck on this test!" or "Error in deploy.sh"
- Click to see error details and intervene

### 5. **Celebrating** 🎉
- Agent just **completed a task successfully**
- Animation: jumping, cheering, confetti
- Brief state (5-10 seconds) before returning to idle
- Shows XP gained, level up notification if applicable

### 6. **Resting** 😴
- Agent is **exhausted** from too many tasks
- Animation: sleeping with Z's
- Can be woken up for urgent tasks but with a penalty
- Energy regenerates slowly over time

## Interactions

### Click an Agent
Opens a **quick-action menu**:
- **View Character Sheet** — full stats and history
- **Assign Quest** — pick a task from your backlog
- **Chat** — ask the agent about their current state
- **Rest** — put the agent on break (if working too long)

### Hover an Agent
Shows a **tooltip**:
```
╔════════════════════╗
║ Bran the Methodical║
║ Level 12 · Wizard  ║
║ Status: Idle       ║
║ Energy: ████░ 80%  ║
║ Last Quest: +45 XP ║
║ "Ready for action!"║
╚════════════════════╝
```

### Drag & Drop
- **Drag a quest card** from the Quest Board onto an agent to assign it
- **Drag an agent** to a team area to form a party

## Ambient Details

To make the camp feel alive:

- **Fireplace** at the center with animated flames
- **Day/night cycle** (optional) — camp lighting changes with time
- **Weather effects** (optional) — rain, snow, fog for ambiance
- **Props** scattered around:
  - Swords in the ground (completed quests)
  - Logs for sitting
  - Tent flaps moving in the breeze
  - A quest board visible in the background (click to zoom)
  
## Responsive Layout

- **Desktop**: Full camp scene, all agents visible
- **Tablet**: Scaled camp, scrollable if needed
- **Mobile**: List view with agent cards + mini camp scene at top

## Sound Design (Optional)

- Crackling fire (ambient loop)
- Footsteps when agents move
- Chime when an agent completes a task
- Alert sound when an agent gets stuck
- Soft medieval fantasy background music

## Example Scenarios

### Scenario 1: All Quiet
- 4 agents sitting around the fire, idle
- Fire crackling, peaceful music
- User opens the quest board to assign new tasks

### Scenario 2: Busy Day
- 2 agents out on quests (shown as "away")
- 1 agent actively working (typing at a desk)
- 1 agent stuck (waving with speech bubble)
- User clicks stuck agent, sees error log, fixes issue

### Scenario 3: Level Up!
- Agent Scáthach returns from a quest
- Celebration animation plays
- Notification: "Scáthach reached Level 15!"
- New skill unlocked: "TypeScript Specialist"
- User clicks to see updated character sheet

## Why This Works

1. **At-a-glance status** — you know immediately who's doing what
2. **Emotional connection** — agents feel like characters, not processes
3. **Low cognitive load** — visual state > reading logs
4. **Delightful** — watching your team work is actually fun
5. **Scalable** — works with 3 agents or 20 (with paging/grouping)

## Future Enhancements

- **Camp customization** — unlock decorations as agents level up
- **Agent conversations** — idle agents chat with each other (flavor text)
- **Seasons** — camp appearance changes over time
- **Special events** — festivals, boss battles, tournaments
- **Photo mode** — take screenshots of your camp to share

---

The camp is **home base**. Make it cozy. Make it yours.
