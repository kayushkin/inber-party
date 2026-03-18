// Skill tree definitions for different agent classes

export interface SkillNode {
  id: string;
  name: string;
  description: string;
  icon: string;
  prerequisites: string[]; // skill IDs that must be unlocked first
  levelRequired: number;
  x: number; // position in skill tree grid
  y: number;
  category: 'core' | 'utility' | 'advanced' | 'mastery';
}

export interface SkillTree {
  className: string;
  description: string;
  skills: SkillNode[];
}

// Define skill trees for each agent class
export const SKILL_TREES: Record<string, SkillTree> = {
  Overlord: {
    className: 'Overlord',
    description: 'Master of delegation and orchestration',
    skills: [
      // Core Skills (Level 1-10)
      { id: 'command', name: 'Command', description: 'Basic delegation abilities', icon: '⚡', prerequisites: [], levelRequired: 1, x: 2, y: 0, category: 'core' },
      { id: 'coordination', name: 'Coordination', description: 'Organize multi-agent tasks', icon: '🎯', prerequisites: ['command'], levelRequired: 3, x: 2, y: 1, category: 'core' },
      { id: 'strategy', name: 'Strategy', description: 'Long-term planning expertise', icon: '🧠', prerequisites: ['coordination'], levelRequired: 5, x: 2, y: 2, category: 'core' },
      
      // Utility Branch (Left)
      { id: 'monitoring', name: 'Monitoring', description: 'Track agent performance', icon: '👁️', prerequisites: ['command'], levelRequired: 4, x: 0, y: 1, category: 'utility' },
      { id: 'analytics', name: 'Analytics', description: 'Performance metrics analysis', icon: '📊', prerequisites: ['monitoring'], levelRequired: 7, x: 0, y: 2, category: 'utility' },
      { id: 'optimization', name: 'Optimization', description: 'Resource allocation mastery', icon: '⚙️', prerequisites: ['analytics'], levelRequired: 10, x: 0, y: 3, category: 'advanced' },
      
      // Authority Branch (Right)
      { id: 'authority', name: 'Authority', description: 'Enhanced command presence', icon: '👑', prerequisites: ['command'], levelRequired: 4, x: 4, y: 1, category: 'utility' },
      { id: 'delegation', name: 'Delegation', description: 'Complex task distribution', icon: '🎭', prerequisites: ['authority'], levelRequired: 7, x: 4, y: 2, category: 'utility' },
      { id: 'domination', name: 'Domination', description: 'Ultimate control over systems', icon: '🔥', prerequisites: ['delegation', 'strategy'], levelRequired: 15, x: 4, y: 3, category: 'mastery' },
      
      // Master Skills (High level)
      { id: 'architect', name: 'System Architect', description: 'Design complex multi-agent systems', icon: '🏛️', prerequisites: ['strategy', 'optimization'], levelRequired: 12, x: 1, y: 4, category: 'mastery' },
      { id: 'overlord_mastery', name: 'Overlord Mastery', description: 'Pinnacle of delegation arts', icon: '⚜️', prerequisites: ['architect', 'domination'], levelRequired: 20, x: 2, y: 5, category: 'mastery' },
    ]
  },
  
  Engineer: {
    className: 'Engineer',
    description: 'Builder of systems and infrastructure',
    skills: [
      // Core Skills
      { id: 'coding', name: 'Coding', description: 'Basic programming abilities', icon: '💻', prerequisites: [], levelRequired: 1, x: 2, y: 0, category: 'core' },
      { id: 'debugging', name: 'Debugging', description: 'Find and fix issues', icon: '🐛', prerequisites: ['coding'], levelRequired: 3, x: 2, y: 1, category: 'core' },
      { id: 'architecture', name: 'Architecture', description: 'Design software systems', icon: '🏗️', prerequisites: ['debugging'], levelRequired: 5, x: 2, y: 2, category: 'core' },
      
      // Infrastructure Branch (Left)
      { id: 'deployment', name: 'Deployment', description: 'Release and deploy code', icon: '🚀', prerequisites: ['coding'], levelRequired: 4, x: 0, y: 1, category: 'utility' },
      { id: 'devops', name: 'DevOps', description: 'Automate infrastructure', icon: '🔧', prerequisites: ['deployment'], levelRequired: 7, x: 0, y: 2, category: 'utility' },
      { id: 'scalability', name: 'Scalability', description: 'Build systems that scale', icon: '📈', prerequisites: ['devops'], levelRequired: 10, x: 0, y: 3, category: 'advanced' },
      
      // Code Quality Branch (Right)
      { id: 'testing', name: 'Testing', description: 'Write comprehensive tests', icon: '🧪', prerequisites: ['coding'], levelRequired: 4, x: 4, y: 1, category: 'utility' },
      { id: 'refactoring', name: 'Refactoring', description: 'Improve code quality', icon: '✨', prerequisites: ['testing'], levelRequired: 7, x: 4, y: 2, category: 'utility' },
      { id: 'mastery', name: 'Code Mastery', description: 'Write elegant, efficient code', icon: '🎨', prerequisites: ['refactoring', 'architecture'], levelRequired: 15, x: 4, y: 3, category: 'mastery' },
      
      // Master Skills
      { id: 'systems_design', name: 'Systems Design', description: 'Architect complex systems', icon: '⚡', prerequisites: ['architecture', 'scalability'], levelRequired: 12, x: 1, y: 4, category: 'mastery' },
      { id: 'engineer_mastery', name: 'Engineer Mastery', description: 'Master of all engineering arts', icon: '⚙️', prerequisites: ['systems_design', 'mastery'], levelRequired: 20, x: 2, y: 5, category: 'mastery' },
    ]
  },
  
  Bard: {
    className: 'Bard',
    description: 'Master of communication and documentation',
    skills: [
      // Core Skills
      { id: 'writing', name: 'Writing', description: 'Clear communication skills', icon: '✍️', prerequisites: [], levelRequired: 1, x: 2, y: 0, category: 'core' },
      { id: 'storytelling', name: 'Storytelling', description: 'Engaging narrative creation', icon: '📖', prerequisites: ['writing'], levelRequired: 3, x: 2, y: 1, category: 'core' },
      { id: 'persuasion', name: 'Persuasion', description: 'Influence through words', icon: '🎭', prerequisites: ['storytelling'], levelRequired: 5, x: 2, y: 2, category: 'core' },
      
      // Documentation Branch (Left)
      { id: 'documentation', name: 'Documentation', description: 'Create clear guides', icon: '📚', prerequisites: ['writing'], levelRequired: 4, x: 0, y: 1, category: 'utility' },
      { id: 'technical_writing', name: 'Technical Writing', description: 'Explain complex topics', icon: '🔧', prerequisites: ['documentation'], levelRequired: 7, x: 0, y: 2, category: 'utility' },
      { id: 'knowledge_mgmt', name: 'Knowledge Management', description: 'Organize team knowledge', icon: '🧠', prerequisites: ['technical_writing'], levelRequired: 10, x: 0, y: 3, category: 'advanced' },
      
      // Performance Branch (Right)
      { id: 'performance', name: 'Performance', description: 'Deliver engaging content', icon: '🎪', prerequisites: ['writing'], levelRequired: 4, x: 4, y: 1, category: 'utility' },
      { id: 'inspiration', name: 'Inspiration', description: 'Motivate and energize teams', icon: '✨', prerequisites: ['performance'], levelRequired: 7, x: 4, y: 2, category: 'utility' },
      { id: 'mastery', name: 'Legendary Bard', description: 'Words that move mountains', icon: '🌟', prerequisites: ['inspiration', 'persuasion'], levelRequired: 15, x: 4, y: 3, category: 'mastery' },
      
      // Master Skills
      { id: 'thought_leadership', name: 'Thought Leadership', description: 'Shape industry discourse', icon: '💡', prerequisites: ['persuasion', 'knowledge_mgmt'], levelRequired: 12, x: 1, y: 4, category: 'mastery' },
      { id: 'bard_mastery', name: 'Bard Mastery', description: 'Master of all communication', icon: '🎶', prerequisites: ['thought_leadership', 'mastery'], levelRequired: 20, x: 2, y: 5, category: 'mastery' },
    ]
  },
  
  Ranger: {
    className: 'Ranger',
    description: 'Explorer of codebases and systems',
    skills: [
      // Core Skills
      { id: 'scouting', name: 'Scouting', description: 'Explore unknown territory', icon: '🔍', prerequisites: [], levelRequired: 1, x: 2, y: 0, category: 'core' },
      { id: 'tracking', name: 'Tracking', description: 'Follow code patterns', icon: '👣', prerequisites: ['scouting'], levelRequired: 3, x: 2, y: 1, category: 'core' },
      { id: 'pathfinding', name: 'Pathfinding', description: 'Navigate complex systems', icon: '🧭', prerequisites: ['tracking'], levelRequired: 5, x: 2, y: 2, category: 'core' },
      
      // Analysis Branch (Left)
      { id: 'analysis', name: 'Analysis', description: 'Deep code examination', icon: '🔬', prerequisites: ['scouting'], levelRequired: 4, x: 0, y: 1, category: 'utility' },
      { id: 'pattern_recognition', name: 'Pattern Recognition', description: 'Identify code patterns', icon: '🧩', prerequisites: ['analysis'], levelRequired: 7, x: 0, y: 2, category: 'utility' },
      { id: 'foresight', name: 'Foresight', description: 'Predict system behavior', icon: '🔮', prerequisites: ['pattern_recognition'], levelRequired: 10, x: 0, y: 3, category: 'advanced' },
      
      // Security Branch (Right)
      { id: 'security', name: 'Security', description: 'Identify vulnerabilities', icon: '🛡️', prerequisites: ['scouting'], levelRequired: 4, x: 4, y: 1, category: 'utility' },
      { id: 'penetration', name: 'Penetration Testing', description: 'Test system defenses', icon: '⚔️', prerequisites: ['security'], levelRequired: 7, x: 4, y: 2, category: 'utility' },
      { id: 'guardian', name: 'Guardian', description: 'Ultimate system protector', icon: '🏰', prerequisites: ['penetration', 'pathfinding'], levelRequired: 15, x: 4, y: 3, category: 'mastery' },
      
      // Master Skills
      { id: 'system_mastery', name: 'System Mastery', description: 'Complete system understanding', icon: '🌐', prerequisites: ['pathfinding', 'foresight'], levelRequired: 12, x: 1, y: 4, category: 'mastery' },
      { id: 'ranger_mastery', name: 'Ranger Mastery', description: 'Master scout and guardian', icon: '🏹', prerequisites: ['system_mastery', 'guardian'], levelRequired: 20, x: 2, y: 5, category: 'mastery' },
    ]
  },
  
  // Default skill tree for unknown classes
  Adventurer: {
    className: 'Adventurer',
    description: 'Versatile problem solver',
    skills: [
      { id: 'adaptation', name: 'Adaptation', description: 'Learn new skills quickly', icon: '🎯', prerequisites: [], levelRequired: 1, x: 2, y: 0, category: 'core' },
      { id: 'versatility', name: 'Versatility', description: 'Handle diverse tasks', icon: '⚖️', prerequisites: ['adaptation'], levelRequired: 3, x: 2, y: 1, category: 'core' },
      { id: 'expertise', name: 'Expertise', description: 'Develop specialized knowledge', icon: '🎓', prerequisites: ['versatility'], levelRequired: 5, x: 2, y: 2, category: 'core' },
      { id: 'mastery', name: 'Mastery', description: 'Peak performance', icon: '👑', prerequisites: ['expertise'], levelRequired: 10, x: 2, y: 3, category: 'mastery' },
    ]
  },
};

// Get skill tree for a specific agent class
export function getSkillTree(agentClass: string): SkillTree {
  return SKILL_TREES[agentClass] || SKILL_TREES['Adventurer'];
}

// Check if a skill is unlocked for an agent
export function isSkillUnlocked(skill: SkillNode, agentLevel: number, unlockedSkills: string[]): boolean {
  // Check level requirement
  if (agentLevel < skill.levelRequired) return false;
  
  // Check prerequisites
  return skill.prerequisites.every(prereq => unlockedSkills.includes(prereq));
}

// Get skills that should be unlocked based on agent's current skills and level
export function getUnlockedSkills(agentClass: string, agentLevel: number, currentSkills: string[]): string[] {
  const skillTree = getSkillTree(agentClass);
  const unlockedSkills = new Set<string>(currentSkills);
  
  // Keep adding skills until no more can be unlocked
  let changed = true;
  while (changed) {
    changed = false;
    for (const skill of skillTree.skills) {
      if (!unlockedSkills.has(skill.id) && isSkillUnlocked(skill, agentLevel, Array.from(unlockedSkills))) {
        unlockedSkills.add(skill.id);
        changed = true;
      }
    }
  }
  
  return Array.from(unlockedSkills);
}