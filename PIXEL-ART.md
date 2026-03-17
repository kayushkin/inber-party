# Pixel Art Generation — Research & Options

## Goal
Generate unique 64x64 pixel art portraits for each agent in Míl Party, replacing emoji avatars with proper RPG character sprites.

## Options

### 1. PixelLab API (pixellab.ai)
- **Cost**: ~$0.007 per sprite
- **Features**: Text-to-pixel-art, transparent backgrounds, animation support
- **Model**: "Pixflux" — up to 400x400 resolution
- **Pros**: Purpose-built for pixel art, consistent style, API available
- **Cons**: External dependency, requires account/credits

### 2. Perchance (perchance.org/ai-pixel-art-generator)
- **Cost**: Free
- **Features**: Browser-based pixel art generator
- **Pros**: Free, decent quality
- **Cons**: No API — manual/browser-only, can't automate

### 3. OpenAI DALL-E / gpt-image-1
- **Cost**: ~$0.04-0.08 per image (depending on size/model)
- **Features**: General image generation with pixel art prompting
- **Pros**: We already have the API key, flexible prompting, can generate any style
- **Cons**: Not pixel-art-specific (needs careful prompting), more expensive per image

### 4. LPC Spritesheet Generator (liberatedpixelcup.github.io)
- **Cost**: Free (CC licensed sprites)
- **Features**: Composable character parts (hair, armor, weapons, etc.)
- **Pros**: Free, consistent style, great for RPG characters, deterministic
- **Cons**: Limited to LPC art style, less unique per-agent

## Recommendation

**Use OpenAI's image API** (gpt-image-1) to generate 64x64 pixel art portraits:

1. We already have the API key configured
2. One-time generation cost: ~$1 for all agents
3. Save as static PNG assets in `frontend/public/avatars/`
4. Prompt template: `"64x64 pixel art portrait of a [class] character named [name], [description], dark fantasy RPG style, transparent background, retro game aesthetic"`
5. Fall back to emoji avatars if image not found

### Implementation Plan
1. Create a script (`scripts/gen-avatars.sh` or Go tool) that calls OpenAI image API
2. Generate one portrait per agent class/name combo
3. Save to `frontend/public/avatars/{agent-id}.png`
4. Update frontend `AgentCard` to use `<img>` with emoji fallback
