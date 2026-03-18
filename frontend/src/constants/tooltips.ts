// Tooltip explanations for stats and skills

export const STAT_TOOLTIPS = {
  level: "The adventurer's experience level. Higher levels unlock new capabilities and indicate more completed quests.",
  
  xp: "Experience Points gained from completing quests and tasks. Earn enough XP to level up!",
  
  energy: "Current energy level (0-100%). Low energy may affect performance and availability for new quests.",
  
  mood: "Agent's emotional state based on recent task success rate, current workload, and rest time. Ranges from Exhausted (0-20) to Happy (81-100). Better mood often correlates with better performance.",
  
  tokens: "Total language model tokens consumed across all sessions. Higher values indicate more complex conversations and tasks.",
  
  cost: "Total monetary cost in USD for all token usage. Based on the language model's pricing tier.",
  
  sessions: "Number of conversation sessions completed. Each quest typically involves one or more sessions.",
  
  // Guild stats
  adventurers: "Total number of active adventurers in the guild roster.",
  
  total_sessions: "Combined session count across all adventurers.",
  
  total_tokens: "Sum of all tokens consumed by the entire guild.",
  
  total_cost: "Total monetary cost for all guild operations in USD.",
  
  quests_done: "Successfully completed quests across all adventurers.",
  
  quests_failed: "Failed or abandoned quests. Learning opportunities!",
  
  active: "Currently running quests and active sessions.",
  
  uptime: "How long the guild systems have been operational."
};

export const SKILL_TOOLTIPS = {
  // Common skill explanations - these can be customized per skill type
  default: "Skill level represents proficiency in this area. Higher levels indicate more experience and better performance.",
  
  // Specific skill types (can be expanded)
  coding: "Programming and software development proficiency. Higher levels enable complex refactoring and architecture tasks.",
  
  communication: "Ability to interact effectively with humans and other agents. Affects collaboration and user interaction quality.",
  
  analysis: "Data analysis and problem-solving skills. Higher levels enable deeper insights and complex reasoning.",
  
  creativity: "Creative thinking and generation of novel solutions. Important for design and innovative problem-solving.",
  
  research: "Information gathering and research capabilities. Higher levels enable more thorough investigations.",
  
  documentation: "Writing and maintaining clear documentation. Essential for knowledge sharing and project maintenance.",
  
  testing: "Quality assurance and testing skills. Higher levels result in more thorough and effective testing strategies.",
  
  deployment: "Infrastructure and deployment management. Critical for production systems and DevOps tasks.",
  
  security: "Cybersecurity awareness and implementation. Important for protecting systems and data.",
  
  leadership: "Project management and team coordination skills. Enables effective orchestration of multi-agent tasks."
};

export const ACHIEVEMENT_TOOLTIPS = {
  // Achievement categories
  first_quest: "Completed their very first quest. Everyone starts somewhere!",
  
  speedster: "Completed a quest in record time. Efficiency matters!",
  
  marathon: "Handled an exceptionally long or complex quest. Persistence pays off!",
  
  collaborator: "Successfully worked with other agents on team quests.",
  
  problem_solver: "Overcome a particularly challenging technical obstacle.",
  
  mentor: "Helped train or guide other agents in their development.",
  
  innovator: "Pioneered a new approach or solution to a common problem.",
  
  reliable: "Maintained consistent performance over many quests.",
  
  specialist: "Achieved mastery in a particular domain or skill area.",
  
  generalist: "Demonstrated competence across multiple skill areas."
};

// Helper function to get skill tooltip based on skill name
export function getSkillTooltip(skillName: string): string {
  const lowerSkill = skillName.toLowerCase();
  
  // Check for specific skill matches
  for (const [key, tooltip] of Object.entries(SKILL_TOOLTIPS)) {
    if (lowerSkill.includes(key)) {
      return tooltip;
    }
  }
  
  // Return default if no specific match
  return SKILL_TOOLTIPS.default;
}