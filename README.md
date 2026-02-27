# Míl Party — Agent Party UI Concept

> **Míl** (pl. Míls) — from Irish mythology: warriors, adventurers, the Milesians who conquered Ireland. Here: your AI coding agents, visualized as pixel-art RPG heroes.

## Vision

What if managing AI coding agents felt less like watching logs scroll by and more like running a party of adventurers in a classic RPG?

**Míl Party** is a concept exploration for visualizing the [inber](https://github.com/kayushkin/inber) multi-agent system as a pixel-art party management UI. Think:

- **Chrono Trigger** camp scenes where your party hangs out between quests
- **Final Fantasy Tactics** job system where agents gain skills and level up
- **Fire Emblem** character progression with relationships and synergies
- **Stardew Valley** charm and personality in a technical tool

## The Core Idea

Each AI agent is a **Míl** — a character with:

- **Pixel-art avatar** (64x64, Celtic/fantasy theme)
- **Name & personality** (Bran the Methodical, Scáthach the Swift)
- **Level & XP** (gained from completing tasks)
- **Stats** (tasks completed, success rate, lines written, tests passed)
- **Quest log** (active, completed, failed tasks)
- **Skills & equipment** (tools they have access to, specializations they've unlocked)
- **Mood & energy** (overworked agents need rest)

## Key UI Views

All detailed in the [concepts/](./concepts/) directory:

1. **[Camp View](./concepts/camp-view.md)** — The main hub. Pixel art camp scene where agents hang out around a fire. Idle agents rest, active ones are "out on quest," stuck ones wave for help.

2. **[Quest Board](./concepts/quest-board.md)** — Post tasks like quests on a bulletin board. Agents claim them or you assign them. Quest cards show difficulty, XP rewards, estimated time.

3. **[Character Sheets](./concepts/character-sheets.md)** — Detailed agent view with stats, quest history, skill trees (e.g., "TypeScript Specialist" unlocked after 50 TS tasks).

4. **[Battle View](./concepts/battle-view.md)** — Active tasks shown as battles. "Bran casts Write File!" "Scáthach uses Test Suite — Critical Hit!" Errors are damage taken. Fun + informative.

5. **[Notifications](./concepts/notifications.md)** — Agents get your attention via waving animations, speech bubbles, notification badges. Priority levels: casual wave → urgent flag → critical SOS.

6. **[Party Management](./concepts/party-management.md)** — Compose teams for complex tasks. See synergies ("Bran + Scáthach: +15% test coverage"). Track which combos work best.

7. **[Progression System](./concepts/progression.md)** — Agents level up, unlock new avatar elements, earn achievements, compete on leaderboards.

## Mockups

See [mockups/](./mockups/) for static HTML/CSS prototypes:

- **camp-view.html** — A simple camp scene mockup
- **character-sheet.html** — Agent stats and history

These use CSS pixel art / emoji placeholders. They're visual sketches, not production code.

## Art Pipeline

See [art/README.md](./art/README.md) for the pixel art generation approach:

- DALL-E for initial sprites
- 64x64 pixel art style, Celtic/Irish fantasy theme
- Each character needs: idle, walking, working, waving, celebrating
- Color palette and style guide

## Why This Matters

Developer tools don't have to be boring. When you're managing a fleet of AI agents working on your codebase:

- **Awareness** — know at a glance who's doing what, who's stuck, who's idle
- **Trust** — see agent history and success rates before assigning critical tasks
- **Engagement** — progression systems make you care about your agents
- **Fun** — debugging is more enjoyable when it's a battle scene

## Tech Stack

Proposed implementation (see [concepts/tech-stack.md](./concepts/tech-stack.md)):

- **React** for UI components
- **PixiJS or Canvas** for pixel art rendering
- **WebSocket** for real-time updates from inber sessions
- **Data model** linking inber's session DB to UI state

## Status

🎨 **This is a concept repo** — a game design document for a UI that doesn't exist yet.

The goal: capture the vision, explore the ideas, inspire the build.

## Next Steps

1. Refine the concepts based on feedback
2. Create actual pixel art for a few example Míls
3. Build a working prototype of the camp view
4. Integrate with a real inber instance
5. Playtest and iterate

---

**Míls, assemble!** 🗡️✨
