// seasonalEvents.ts - Seasonal event detection and theming

export type Season = 'spring' | 'summer' | 'autumn' | 'winter';
export type Holiday = 
  | 'new-years' | 'valentines' | 'st-patricks' | 'easter' | 'halloween' 
  | 'thanksgiving' | 'christmas' | 'independence-day' | 'april-fools';

export interface SeasonalEvent {
  id: string;
  name: string;
  season: Season;
  holiday?: Holiday;
  startDate: Date;
  endDate: Date;
  description: string;
  theme: SeasonalTheme;
  questModifiers: QuestModifiers;
}

export interface SeasonalTheme {
  colors: {
    primary: string;
    secondary: string;
    accent: string;
    background: string;
  };
  decorations: string[];
  animations: string[];
  particles?: string;
}

export interface QuestModifiers {
  namePrefix?: string;
  nameSuffix?: string;
  difficultyMultiplier?: number;
  xpBonus?: number;
  specialRewards?: string[];
  questTypes?: string[];
}

// Get current season based on date (Northern hemisphere)
export function getCurrentSeason(): Season {
  const now = new Date();
  const month = now.getMonth() + 1; // 1-12
  
  if (month >= 3 && month <= 5) return 'spring';
  if (month >= 6 && month <= 8) return 'summer';
  if (month >= 9 && month <= 11) return 'autumn';
  return 'winter';
}

// Define seasonal events for the year
export function getSeasonalEvents(year: number): SeasonalEvent[] {
  return [
    // New Year's Event
    {
      id: 'new-years-2026',
      name: 'New Beginnings Festival',
      season: 'winter',
      holiday: 'new-years',
      startDate: new Date(year, 0, 1), // Jan 1
      endDate: new Date(year, 0, 3),   // Jan 3
      description: 'A time for fresh starts and new adventures!',
      theme: {
        colors: {
          primary: '#ffd700',   // Gold
          secondary: '#c0c0c0', // Silver
          accent: '#00ffff',    // Cyan
          background: '#0a0a2e' // Dark blue
        },
        decorations: ['✨', '🎆', '🎊', '🥳'],
        animations: ['sparkle', 'confetti']
      },
      questModifiers: {
        namePrefix: 'New Year\'s',
        xpBonus: 0.5,
        specialRewards: ['Resolution Keeper', 'Fresh Start']
      }
    },

    // Valentine's Day
    {
      id: 'valentines-2026',
      name: 'Festival of Hearts',
      season: 'winter',
      holiday: 'valentines',
      startDate: new Date(year, 1, 13), // Feb 14-16
      endDate: new Date(year, 1, 16),
      description: 'Spread love and cooperation among adventurers!',
      theme: {
        colors: {
          primary: '#ff69b4',   // Hot pink
          secondary: '#ff1493', // Deep pink
          accent: '#ffffff',    // White
          background: '#2d0a1f' // Dark pink
        },
        decorations: ['💖', '💝', '🌹', '💕'],
        animations: ['hearts', 'love-pulse']
      },
      questModifiers: {
        namePrefix: 'Romantic',
        nameSuffix: 'of Love',
        xpBonus: 0.25,
        questTypes: ['collaboration', 'pair-programming']
      }
    },

    // St. Patrick's Day
    {
      id: 'st-patricks-2026',
      name: 'Emerald Adventure Day',
      season: 'spring',
      holiday: 'st-patricks',
      startDate: new Date(year, 2, 17), // Mar 17
      endDate: new Date(year, 2, 17),
      description: 'May luck be with your coding adventures!',
      theme: {
        colors: {
          primary: '#00ff00',   // Bright green
          secondary: '#228b22', // Forest green
          accent: '#ffd700',    // Gold
          background: '#0a2e0a' // Dark green
        },
        decorations: ['🍀', '🌈', '🏆', '🪙'],
        animations: ['rainbow', 'sparkle']
      },
      questModifiers: {
        namePrefix: 'Lucky',
        xpBonus: 0.17, // 17% for luck!
        specialRewards: ['Pot of Gold', 'Four Leaf Clover']
      }
    },

    // Halloween
    {
      id: 'halloween-2026',
      name: 'Haunted Code Celebration',
      season: 'autumn',
      holiday: 'halloween',
      startDate: new Date(year, 9, 25), // Oct 25-31
      endDate: new Date(year, 9, 31),
      description: 'Face your spookiest bugs and darkest deadlines!',
      theme: {
        colors: {
          primary: '#ff4500',   // Orange red
          secondary: '#9932cc', // Dark orchid
          accent: '#32cd32',    // Lime green
          background: '#1a0d0d' // Very dark red
        },
        decorations: ['🎃', '👻', '🦇', '🕷️', '🕸️', '💀'],
        animations: ['spooky-glow', 'bat-fly', 'fog'],
        particles: 'falling-leaves'
      },
      questModifiers: {
        namePrefix: 'Haunted',
        nameSuffix: 'of Shadows',
        difficultyMultiplier: 1.31, // Halloween difficulty spike
        specialRewards: ['Phantom Debugger', 'Spectral Refactor']
      }
    },

    // Christmas
    {
      id: 'christmas-2026',
      name: 'Winter Solstice Festival',
      season: 'winter',
      holiday: 'christmas',
      startDate: new Date(year, 11, 20), // Dec 20-26
      endDate: new Date(year, 11, 26),
      description: 'A time of giving, sharing, and collaborative coding!',
      theme: {
        colors: {
          primary: '#dc143c',   // Crimson
          secondary: '#228b22', // Forest green
          accent: '#ffd700',    // Gold
          background: '#0f2a0f' // Dark forest green
        },
        decorations: ['🎄', '❄️', '🎁', '⭐', '🔔', '🕯️'],
        animations: ['snow-fall', 'twinkle', 'bell-chime'],
        particles: 'snowflakes'
      },
      questModifiers: {
        namePrefix: 'Festive',
        nameSuffix: 'of Wonder',
        xpBonus: 0.25,
        specialRewards: ['Gift of Knowledge', 'Stellar Performance'],
        questTypes: ['gift-giving', 'helping-others', 'documentation']
      }
    }
  ];
}

// Check if a seasonal event is currently active
export function getActiveSeasonalEvent(): SeasonalEvent | null {
  const now = new Date();
  const currentYear = now.getFullYear();
  const events = getSeasonalEvents(currentYear);
  
  return events.find(event => {
    return now >= event.startDate && now <= event.endDate;
  }) || null;
}

// Get seasonal quest name variations
export function getSeasonalQuestName(baseQuestName: string, season: Season): string {
  const activeEvent = getActiveSeasonalEvent();
  
  if (activeEvent) {
    const modifiers = activeEvent.questModifiers;
    let name = baseQuestName;
    
    if (modifiers.namePrefix) {
      name = `${modifiers.namePrefix} ${name}`;
    }
    if (modifiers.nameSuffix) {
      name = `${name} ${modifiers.nameSuffix}`;
    }
    
    return name;
  }
  
  // Fallback seasonal naming based on season
  const seasonalPrefixes = {
    spring: ['Blooming', 'Fresh', 'Renewed', 'Growing'],
    summer: ['Blazing', 'Radiant', 'Vibrant', 'Energetic'],
    autumn: ['Harvest', 'Golden', 'Rustic', 'Crisp'],
    winter: ['Frosty', 'Crystal', 'Stark', 'Cozy']
  };
  
  const prefixes = seasonalPrefixes[season];
  const randomPrefix = prefixes[Math.floor(Math.random() * prefixes.length)];
  
  return `${randomPrefix} ${baseQuestName}`;
}

// Apply seasonal theme to CSS variables
export function applySeasonalTheme(event: SeasonalEvent | null): void {
  if (!event) return;
  
  const root = document.documentElement;
  const theme = event.theme;
  
  // Apply seasonal color overrides
  root.style.setProperty('--seasonal-primary', theme.colors.primary);
  root.style.setProperty('--seasonal-secondary', theme.colors.secondary);
  root.style.setProperty('--seasonal-accent', theme.colors.accent);
  root.style.setProperty('--seasonal-background', theme.colors.background);
  
  // Add seasonal class to body for CSS animations/effects
  document.body.classList.add(`seasonal-${event.holiday || event.season}`);
}

// Remove seasonal theme
export function removeSeasonalTheme(): void {
  const root = document.documentElement;
  
  // Remove seasonal CSS variables
  root.style.removeProperty('--seasonal-primary');
  root.style.removeProperty('--seasonal-secondary');
  root.style.removeProperty('--seasonal-accent');
  root.style.removeProperty('--seasonal-background');
  
  // Remove seasonal classes
  document.body.className = document.body.className.replace(/seasonal-\w+/g, '').trim();
}

// Get seasonal decorations for UI
export function getSeasonalDecorations(): string[] {
  const activeEvent = getActiveSeasonalEvent();
  return activeEvent?.theme.decorations || [];
}

// Check if we should show seasonal animations
export function shouldShowSeasonalAnimations(): boolean {
  return getActiveSeasonalEvent() !== null;
}

// Get seasonal XP bonus multiplier
export function getSeasonalXpMultiplier(): number {
  const activeEvent = getActiveSeasonalEvent();
  return activeEvent?.questModifiers.xpBonus ? (1 + activeEvent.questModifiers.xpBonus) : 1;
}