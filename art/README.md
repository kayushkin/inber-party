# Pixel Art Pipeline for Míl Party

This directory will contain the pixel art assets for agent avatars, UI elements, and animations.

## Overview

The visual identity of Míl Party relies on **pixel art** with a **Celtic/Irish fantasy theme**. Think classic RPGs meets Irish mythology.

## Style Guide

### Core Aesthetic
- **64x64 pixel sprites** for agent avatars
- **Fantasy RPG style** (Final Fantasy Tactics, Fire Emblem, Chrono Trigger)
- **Celtic/Irish theme** (Celtic knots, torcs, tartans, mythological motifs)
- **Warm, earthy color palette** (greens, golds, browns)

### Technical Specs
- **Resolution**: 64x64 pixels per character sprite
- **Format**: PNG with transparency
- **Color depth**: Indexed color (16-256 colors per sprite)
- **Pixel-perfect**: Clean edges, no anti-aliasing on sprites
- **Frame rate**: 8-12 fps for animations

## Color Palette

### Primary Colors
```
Greens (nature, forest, Irish landscape):
#2d5016 — Dark forest green
#4a7c2e — Medium green
#76a34d — Light grass green
#a0c46d — Pale green highlight

Golds (armor, accents, magic):
#d4af37 — Old gold
#ffdf00 — Bright gold
#ffd700 — Golden yellow
#f4c542 — Warm gold

Browns (leather, wood, earth):
#5c4033 — Dark brown
#8b7355 — Medium brown
#a0826d — Light tan
#c9a679 — Pale sand

Blues (magic, water, night):
#1e3a5f — Deep blue
#2e5984 — Medium blue
#4a90c4 — Sky blue
#7fb3d5 — Pale blue

Reds (critical, errors, fire):
#8b0000 — Dark red
#b22222 — Fire brick
#dc143c — Crimson
#ff6347 — Tomato
```

### Accent Colors
- **Purple** (#6a0dad, #9370db) — magic, rare items
- **Orange** (#ff8c00, #ffa500) — fire, urgency
- **White/Gray** (#f0f0f0, #cccccc, #888888) — neutral, UI elements

## Generation Pipeline

### Step 1: DALL-E Generation

Use OpenAI's DALL-E API to generate initial sprites:

**Example Prompt:**
```
64x64 pixel art character, fantasy RPG style, Celtic Irish warrior,
male, standing idle pose, front-facing view, clean background,
pixel-perfect edges, retro game aesthetic, warm color palette
```

**Variations:**
- **Bran the Methodical** — wizard, staff, blue robes, thoughtful expression
- **Scáthach the Swift** — ranger, bow, green cloak, alert stance
- **Aoife the Bold** — warrior, sword, red armor, confident pose
- **Fionn the Wise** — cleric, book, white robes, serene expression
- **Deirdre the Clever** — rogue, daggers, dark leather, sly grin

**API Usage:**
```bash
curl https://api.openai.com/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "prompt": "64x64 pixel art character, fantasy RPG...",
    "n": 1,
    "size": "1024x1024",
    "model": "dall-e-3"
  }'
```

Then scale down to 64x64 using nearest-neighbor (no blur).

### Step 2: Manual Touch-Up (Optional)

Tools for editing:
- **Aseprite** (paid, best for pixel art) — https://www.aseprite.org/
- **Piskel** (free, web-based) — https://www.piskelapp.com/
- **GIMP** (free) — with pencil tool + pixel grid

Refine:
- Clean up edges
- Adjust colors to match palette
- Add details (Celtic knots, torcs, patterns)
- Ensure readability at 64x64

### Step 3: Create Animation Frames

Each character needs multiple sprites for different states:

#### Required Sprites (Priority 1)
- **Idle** (1-2 frames) — standing still, gentle breathing
- **Working** (2-4 frames) — typing, wielding tool, focused
- **Waving** (2-3 frames) — arm raised, attention-grabbing
- **Celebrating** (2-4 frames) — jumping, cheering, success pose

#### Nice-to-Have Sprites (Priority 2)
- **Walking** (4 frames) — left-right foot cycle
- **Resting** (1-2 frames) — sitting, sleeping
- **Stuck** (2-3 frames) — confused, error state, question mark

#### Advanced Sprites (Priority 3)
- **Attacking** (3-4 frames) — for battle view (casting spell, swinging sword)
- **Damaged** (1-2 frames) — recoil, hit reaction
- **Victory** (4-6 frames) — elaborate celebration

### Step 4: Create Sprite Sheets

Combine animation frames into sprite sheets for efficient loading.

**Layout Example (64x64 per frame):**
```
┌────┬────┬────┬────┬────┬────┐
│Idle│Idle│Work│Work│Work│Wave│  Bran (Row 1)
│ 1  │ 2  │ 1  │ 2  │ 3  │ 1  │
├────┼────┼────┼────┼────┼────┤
│Wave│Wave│Cele│Cele│Cele│Cele│  Bran (Row 2)
│ 2  │ 3  │ 1  │ 2  │ 3  │ 4  │
├────┼────┼────┼────┼────┼────┤
│Idle│Idle│Work│Work│Work│Wave│  Scáthach (Row 3)
│ 1  │ 2  │ 1  │ 2  │ 3  │ 1  │
└────┴────┴────┴────┴────┴────┘
```

**Tools:**
- **TexturePacker** — https://www.codeandweb.com/texturepacker
- **Shoebox** (free) — https://renderhjs.net/shoebox/
- Manual grid in Aseprite

### Step 5: Export for PixiJS

Export sprite sheets as:
- **PNG** file (sprites.png)
- **JSON** metadata (sprites.json) — frame coordinates, animation data

**PixiJS Loading:**
```javascript
const loader = PIXI.Loader.shared;
loader.add('sprites', 'sprites.json');
loader.load((loader, resources) => {
  const sheet = resources.sprites.spritesheet;
  const idle = new PIXI.AnimatedSprite(sheet.animations.bran_idle);
  idle.animationSpeed = 0.1;
  idle.play();
});
```

## Asset Checklist

### Characters (Phase 1 MVP)
- [ ] Bran the Methodical (wizard)
  - [ ] Idle (2 frames)
  - [ ] Working (3 frames)
  - [ ] Waving (3 frames)
- [ ] Scáthach the Swift (ranger)
  - [ ] Idle (2 frames)
  - [ ] Working (3 frames)
  - [ ] Waving (3 frames)
- [ ] Aoife the Bold (warrior)
  - [ ] Idle (2 frames)
  - [ ] Working (3 frames)
  - [ ] Waving (3 frames)

### UI Elements
- [ ] Fire animation (4 frames, loop)
- [ ] Trees (static sprites, 3 variations)
- [ ] Camp props (logs, tents, stones)
- [ ] Quest board (background image)
- [ ] Speech bubble templates (info, question, error)

### Icons
- [ ] Quest difficulty (swords: 1-4)
- [ ] Skills (TypeScript, React, Testing, etc.)
- [ ] Achievements (trophies, badges)
- [ ] Status icons (working, stuck, resting)

## Future Enhancements

### Character Progression Visual Upgrades
As agents level up, their sprites evolve:
- **Level 1-9**: Basic gear
- **Level 10-19**: Better equipment, subtle glow
- **Level 20-29**: Advanced gear, visible aura
- **Level 30+**: Legendary gear, particle effects

Example evolution for Bran:
```
Level 1:  🧙 — Simple robe, wooden staff
Level 10: 🧙✨ — Nicer robe, glow effect
Level 20: 🧙💫 — Elegant robe, magic aura
Level 30: 🧙⭐ — Epic robes, swirling magic
```

### Cosmetic Items (Unlockable)
- Hats (wizard hat, crown, bandana)
- Capes (red, blue, green, gold)
- Weapons (different staffs, swords, bows)
- Seasonal skins (Halloween pumpkin head, winter scarf)

### Environment Art
- Day/night variations of camp scene
- Weather effects (rain, snow, fog)
- Seasonal themes (spring flowers, autumn leaves, winter snow)

## Tools & Resources

### Creation Tools
- **Aseprite** — $19.99, best pixel art editor
- **Piskel** — free, browser-based
- **GIMP** — free, general-purpose
- **Inkscape** — free, vector art (for scaling)

### Inspiration
- **Chrono Trigger** sprites
- **Final Fantasy Tactics** character design
- **Stardew Valley** aesthetic
- **Celtic art & mythology** references

### Palette Generators
- **Lospec Palette List** — https://lospec.com/palette-list
- **Coolors** — https://coolors.co/

### Pixel Art Communities
- **Pixelation** — https://pixelation.org/
- **r/PixelArt** — https://reddit.com/r/PixelArt

## Directory Structure (Future)

```
art/
├── README.md (this file)
├── sprites/
│   ├── characters/
│   │   ├── bran/
│   │   │   ├── idle.png
│   │   │   ├── working.png
│   │   │   ├── waving.png
│   │   │   └── celebrating.png
│   │   ├── scathach/
│   │   └── aoife/
│   ├── ui/
│   │   ├── fire.png
│   │   ├── trees.png
│   │   └── camp-bg.png
│   └── icons/
│       ├── difficulty-swords.png
│       ├── skills.png
│       └── achievements.png
├── sprite-sheets/
│   ├── characters.png
│   ├── characters.json
│   ├── ui.png
│   └── ui.json
└── prompts/
    ├── bran.txt
    ├── scathach.txt
    └── aoife.txt
```

---

**Goal:** Make every agent feel like a character you'd meet in a classic RPG. Give them personality through their pixel art.
