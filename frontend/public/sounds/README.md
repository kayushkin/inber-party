# Sound Assets

This directory contains sound effects for the Inber Party app.

## Current Files

The app will fallback to generated tones if these files are not found:

- `level-up.wav` - Played when an agent levels up
- `quest-complete.wav` - Played when a quest is completed  
- `message.wav` - Played when receiving new messages
- `error.wav` - Played when errors occur

## Adding Custom Sounds

To add custom sound effects:

1. Replace any of the .wav files with your own audio files
2. Keep files short (1-3 seconds) for best UX
3. Use common web formats: .wav, .mp3, .ogg
4. Test volume levels - sounds should be pleasant, not jarring

## Fallback Behavior

If sound files are missing, the app generates simple tones:
- Level up: Ascending tone (C5, 523Hz)
- Quest complete: Triumphant tone (A4, 440Hz) 
- Message: Short notification (800Hz)
- Error: Warning tone (200Hz sawtooth)