// Equipment system - maps OpenClaw tools to RPG gear
// Each piece of equipment represents a tool or capability the agent has access to

export interface Equipment {
  id: string;
  name: string;
  description: string;
  icon: string;
  type: 'weapon' | 'armor' | 'accessory' | 'tool' | 'hat';
  rarity: 'common' | 'uncommon' | 'rare' | 'epic' | 'legendary';
  toolRequirement?: string; // OpenClaw tool name required
  skillRequirement?: string; // Skill name required
  levelRequirement?: number; // Minimum level
  roleRequirement?: 'orchestrator' | 'agent'; // Role-based equipment
  statBonus?: Record<string, number>; // Stat bonuses
}

// Equipment definitions - maps to OpenClaw tools and capabilities
export const EQUIPMENT_CATALOG: Equipment[] = [
  // === HATS (Role-based headwear) ===
  {
    id: 'captain_hat',
    name: 'Captain\'s Hat',
    description: 'Distinguished naval cap worn by orchestrator agents who command fleets of sub-agents.',
    icon: '🎩',
    type: 'hat',
    rarity: 'rare',
    roleRequirement: 'orchestrator',
    levelRequirement: 1,
    statBonus: { leadership: 4, command: 3, presence: 2 }
  },
  {
    id: 'hard_hat',
    name: 'Engineer\'s Hard Hat',
    description: 'Safety-first headgear for agents who work with heavy code and system infrastructure.',
    icon: '⛑️',
    type: 'hat',
    rarity: 'uncommon',
    roleRequirement: 'agent',
    skillRequirement: 'Tool Mastery',
    levelRequirement: 3,
    statBonus: { protection: 3, efficiency: 2 }
  },
  {
    id: 'wizard_hat',
    name: 'Wizard\'s Pointy Hat',
    description: 'Mystical hat worn by agents who weave complex spells of code and logic.',
    icon: '🧙‍♂️',
    type: 'hat',
    rarity: 'epic',
    roleRequirement: 'agent',
    levelRequirement: 8,
    statBonus: { wisdom: 5, magical_power: 4, insight: 3 }
  },
  {
    id: 'cowboy_hat',
    name: 'Ranger\'s Cowboy Hat',
    description: 'Wide-brimmed hat for agents who venture into the wild frontiers of the web.',
    icon: '🤠',
    type: 'hat',
    rarity: 'uncommon',
    roleRequirement: 'agent',
    skillRequirement: 'Swift Execution',
    levelRequirement: 4,
    statBonus: { range: 4, adventure: 3, speed: 2 }
  },
  {
    id: 'top_hat',
    name: 'Gentleman\'s Top Hat',
    description: 'Elegant formal hat for distinguished agents of culture and refinement.',
    icon: '🎭',
    type: 'hat',
    rarity: 'rare',
    roleRequirement: 'agent',
    levelRequirement: 6,
    statBonus: { charisma: 4, eloquence: 3, sophistication: 2 }
  },
  {
    id: 'beret',
    name: 'Artist\'s Beret',
    description: 'Creative beret for agents specializing in design, documentation, and artistic expression.',
    icon: '👨‍🎨',
    type: 'hat',
    rarity: 'uncommon',
    roleRequirement: 'agent',
    levelRequirement: 2,
    statBonus: { creativity: 4, inspiration: 2, aesthetic: 1 }
  },
  {
    id: 'crown',
    name: 'Royal Crown',
    description: 'Majestic golden crown reserved for the highest-level orchestrator agents.',
    icon: '👑',
    type: 'hat',
    rarity: 'legendary',
    roleRequirement: 'orchestrator',
    levelRequirement: 15,
    statBonus: { command: 8, authority: 6, presence: 5, majesty: 4 }
  },
  
  // === WEAPONS (Primary Tools) ===
  {
    id: 'iron_sword',
    name: 'Iron Sword',
    description: 'Basic shell access for system commands. Sharp enough to cut through most problems.',
    icon: '⚔️',
    type: 'weapon',
    rarity: 'common',
    toolRequirement: 'exec',
    levelRequirement: 1,
    statBonus: { power: 2 }
  },
  {
    id: 'steel_blade',
    name: 'Steel Blade',
    description: 'Enhanced shell access with background process management. Forged for experienced warriors.',
    icon: '🗡️',
    type: 'weapon',
    rarity: 'uncommon',
    toolRequirement: 'exec',
    skillRequirement: 'Tool Mastery',
    levelRequirement: 5,
    statBonus: { power: 4, efficiency: 1 }
  },
  {
    id: 'spyglass',
    name: 'Web Spyglass',
    description: 'Magical lens that peers across the vast network. See what others cannot.',
    icon: '🔍',
    type: 'weapon',
    rarity: 'uncommon',
    toolRequirement: 'web_search',
    levelRequirement: 3,
    statBonus: { wisdom: 3, range: 2 }
  },
  {
    id: 'tome_of_knowledge',
    name: 'Tome of Web Knowledge',
    description: 'Ancient book that fetches wisdom from distant archives. Knowledge is power.',
    icon: '📚',
    type: 'weapon',
    rarity: 'rare',
    toolRequirement: 'web_fetch',
    levelRequirement: 4,
    statBonus: { wisdom: 4, insight: 2 }
  },
  {
    id: 'browser_staff',
    name: 'Staff of Browser Mastery',
    description: 'Mystical staff that commands the web browsers of the realm. Click, type, and navigate with power.',
    icon: '🪄',
    type: 'weapon',
    rarity: 'epic',
    toolRequirement: 'browser',
    levelRequirement: 8,
    statBonus: { power: 6, precision: 4 }
  },
  {
    id: 'quill_of_creation',
    name: 'Quill of Creation',
    description: 'Magical writing implement that brings files into existence with a mere stroke.',
    icon: '🪶',
    type: 'weapon',
    rarity: 'uncommon',
    toolRequirement: 'write',
    levelRequirement: 2,
    statBonus: { creativity: 3, productivity: 2 }
  },
  {
    id: 'lens_of_reading',
    name: 'Lens of True Reading',
    description: 'Crystal lens that reveals the contents of any scroll or tome. No secret stays hidden.',
    icon: '🔎',
    type: 'weapon',
    rarity: 'common',
    toolRequirement: 'read',
    levelRequirement: 1,
    statBonus: { wisdom: 2, insight: 1 }
  },
  
  // === ARMOR (Defensive Capabilities) ===
  {
    id: 'apprentice_robes',
    name: 'Apprentice Robes',
    description: 'Simple robes worn by those just beginning their journey. Better than nothing.',
    icon: '👘',
    type: 'armor',
    rarity: 'common',
    levelRequirement: 1,
    statBonus: { defense: 1 }
  },
  {
    id: 'debugger_armor',
    name: 'Debugger\'s Chain Mail',
    description: 'Protective gear that shields against crashes and errors. Each link forged from resolved bugs.',
    icon: '🛡️',
    type: 'armor',
    rarity: 'uncommon',
    skillRequirement: 'debugging',
    levelRequirement: 4,
    statBonus: { defense: 3, resilience: 2 }
  },
  {
    id: 'masters_robes',
    name: 'Master\'s Enchanted Robes',
    description: 'Robes woven from pure expertise. Radiates competence and commands respect.',
    icon: '🧙',
    type: 'armor',
    rarity: 'rare',
    levelRequirement: 10,
    statBonus: { defense: 4, wisdom: 3, presence: 2 }
  },
  {
    id: 'overlord_crown',
    name: 'Crown of the Overlord',
    description: 'Majestic crown that marks the bearer as a true commander. Others naturally defer to your authority.',
    icon: '👑',
    type: 'armor',
    rarity: 'legendary',
    skillRequirement: 'overlord_mastery',
    levelRequirement: 15,
    statBonus: { defense: 6, command: 8, presence: 5 }
  },
  
  // === ACCESSORIES (Special Tools) ===
  {
    id: 'memory_crystal',
    name: 'Crystal of Memory',
    description: 'Glowing crystal that stores and recalls precious memories. Never forget important details.',
    icon: '💎',
    type: 'accessory',
    rarity: 'rare',
    toolRequirement: 'memory_search',
    levelRequirement: 6,
    statBonus: { memory: 4, recall: 3 }
  },
  {
    id: 'vision_amulet',
    name: 'Amulet of True Sight',
    description: 'Mystical amulet that reveals the truth within images. See beyond the surface.',
    icon: '👁️',
    type: 'accessory',
    rarity: 'epic',
    toolRequirement: 'image',
    levelRequirement: 7,
    statBonus: { perception: 5, insight: 3 }
  },
  {
    id: 'voice_pendant',
    name: 'Pendant of Melodic Voice',
    description: 'Musical pendant that grants the power of speech. Your words become sound.',
    icon: '🎵',
    type: 'accessory',
    rarity: 'uncommon',
    toolRequirement: 'tts',
    levelRequirement: 3,
    statBonus: { charisma: 3, expression: 2 }
  },
  {
    id: 'node_compass',
    name: 'Compass of Remote Nodes',
    description: 'Enchanted compass that points toward distant allies. Control devices from afar.',
    icon: '🧭',
    type: 'accessory',
    rarity: 'epic',
    toolRequirement: 'nodes',
    levelRequirement: 9,
    statBonus: { range: 6, coordination: 3 }
  },
  {
    id: 'messenger_horn',
    name: 'Horn of Swift Messaging',
    description: 'Silver horn that carries messages across vast distances instantly.',
    icon: '📯',
    type: 'accessory',
    rarity: 'rare',
    toolRequirement: 'message',
    levelRequirement: 5,
    statBonus: { communication: 4, speed: 2 }
  },
  {
    id: 'time_hourglass',
    name: 'Hourglass of Scheduling',
    description: 'Mystical timepiece that bends time to your will. Schedule tasks for future execution.',
    icon: '⏳',
    type: 'accessory',
    rarity: 'legendary',
    toolRequirement: 'cron',
    levelRequirement: 12,
    statBonus: { temporal_mastery: 6, foresight: 4 }
  },
  
  // === TOOLS (Utility Items) ===
  {
    id: 'canvas_brush',
    name: 'Enchanted Canvas Brush',
    description: 'Magical brush that paints interactive displays into existence.',
    icon: '🖌️',
    type: 'tool',
    rarity: 'uncommon',
    toolRequirement: 'canvas',
    levelRequirement: 4,
    statBonus: { creativity: 3, presentation: 3 }
  },
  {
    id: 'gateway_key',
    name: 'Master Key of Gateways',
    description: 'Golden key that unlocks the deepest system controls. Handle with extreme care.',
    icon: '🗝️',
    type: 'tool',
    rarity: 'legendary',
    toolRequirement: 'gateway',
    levelRequirement: 20,
    statBonus: { system_mastery: 8, authority: 6 }
  },
  {
    id: 'spawning_scroll',
    name: 'Scroll of Agent Spawning',
    description: 'Ancient scroll that summons fellow adventurers to aid in quests.',
    icon: '📜',
    type: 'tool',
    rarity: 'epic',
    toolRequirement: 'sessions_spawn',
    levelRequirement: 8,
    statBonus: { leadership: 5, coordination: 4 }
  },
  {
    id: 'editing_stylus',
    name: 'Stylus of Precise Editing',
    description: 'Delicate tool for making surgical changes to ancient texts.',
    icon: '✒️',
    type: 'tool',
    rarity: 'common',
    toolRequirement: 'edit',
    levelRequirement: 2,
    statBonus: { precision: 3, care: 1 }
  }
];

// Equipment rarity colors for UI
export const RARITY_COLORS = {
  common: '#9d9d9d',
  uncommon: '#1eff00',
  rare: '#0070dd',
  epic: '#a335ee',
  legendary: '#ff8000'
};

// Get equipment that an agent should have based on their capabilities
export function getAgentEquipment(agent: any, availableTools: string[]): Equipment[] {
  const equipped: Equipment[] = [];
  
  // Determine if this agent is an orchestrator
  const isOrchestrator = determineIfOrchestrator(agent);
  
  for (const item of EQUIPMENT_CATALOG) {
    let canEquip = true;
    
    // Check level requirement
    if (item.levelRequirement && agent.level < item.levelRequirement) {
      canEquip = false;
    }
    
    // Check tool requirement
    if (item.toolRequirement && !availableTools.includes(item.toolRequirement)) {
      canEquip = false;
    }
    
    // Check skill requirement
    if (item.skillRequirement) {
      const hasSkill = agent.skills?.some((skill: any) => skill.skill_name === item.skillRequirement);
      if (!hasSkill) {
        canEquip = false;
      }
    }
    
    // Check role requirement
    if (item.roleRequirement) {
      if (item.roleRequirement === 'orchestrator' && !isOrchestrator) {
        canEquip = false;
      }
      if (item.roleRequirement === 'agent' && isOrchestrator) {
        canEquip = false;
      }
    }
    
    if (canEquip) {
      equipped.push(item);
    }
  }
  
  // Sort by rarity and type for better display
  const rarityOrder = { common: 0, uncommon: 1, rare: 2, epic: 3, legendary: 4 };
  const typeOrder = { weapon: 0, armor: 1, hat: 2, accessory: 3, tool: 4 };
  
  return equipped.sort((a, b) => {
    const rarityDiff = rarityOrder[b.rarity] - rarityOrder[a.rarity];
    if (rarityDiff !== 0) return rarityDiff;
    return typeOrder[a.type] - typeOrder[b.type];
  });
}

// Helper function to determine if an agent is an orchestrator
function determineIfOrchestrator(agent: any): boolean {
  // Key orchestrator agents based on common naming patterns
  const knownOrchestrators = ['main', 'claxon', 'inber', 'openclaw'];
  
  // Check if the agent name is a known orchestrator
  if (knownOrchestrators.includes(agent.id?.toLowerCase() || agent.name?.toLowerCase())) {
    return true;
  }
  
  // Check if the agent's orchestrator field is empty or matches itself
  // This typically indicates it's a top-level orchestrator
  if (!agent.orchestrator || agent.orchestrator === agent.id || agent.orchestrator === agent.name) {
    return true;
  }
  
  // Additional logic: agents with very high levels and command-related classes
  if (agent.level >= 10 && ['Overlord', 'Sovereign'].includes(agent.class)) {
    return true;
  }
  
  return false;
}

// Get available tools for an agent (this would normally come from API)
// For now, we'll infer from agent class and skills
export function inferAvailableTools(agent: any): string[] {
  const tools = ['read', 'write', 'edit']; // Basic tools everyone has
  
  // Add tools based on agent level and class
  if (agent.level >= 1) tools.push('exec');
  if (agent.level >= 3) tools.push('web_search');
  if (agent.level >= 4) tools.push('web_fetch');
  if (agent.level >= 5) tools.push('message');
  if (agent.level >= 6) tools.push('memory_search');
  if (agent.level >= 7) tools.push('image');
  if (agent.level >= 8) tools.push('browser', 'sessions_spawn');
  if (agent.level >= 9) tools.push('nodes');
  
  // Class-specific tools
  switch (agent.class) {
    case 'Overlord':
      tools.push('sessions_spawn', 'cron', 'gateway');
      break;
    case 'Engineer':
      tools.push('exec', 'gateway');
      break;
    case 'Bard':
      tools.push('tts', 'message');
      break;
    case 'Ranger':
      tools.push('web_search', 'web_fetch', 'browser', 'nodes');
      break;
  }
  
  // Add based on skills
  if (agent.skills?.some((s: any) => s.skill_name === 'Tool Mastery' && s.level >= 5)) {
    tools.push('canvas');
  }
  
  return [...new Set(tools)]; // Remove duplicates
}