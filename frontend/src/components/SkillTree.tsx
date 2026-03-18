import { useState } from 'react';
import { getSkillTree, isSkillUnlocked, getUnlockedSkills, type SkillNode } from '../constants/skillTrees';
import { classColor } from '../store';
import Tooltip from './Tooltip';
import './SkillTree.css';

interface SkillTreeProps {
  agentClass: string;
  agentLevel: number;
  agentSkills: { skill_name: string; level: number }[];
}

export default function SkillTree({ agentClass, agentLevel }: SkillTreeProps) {
  const [selectedSkill, setSelectedSkill] = useState<SkillNode | null>(null);
  
  const skillTree = getSkillTree(agentClass);
  
  // Map current skill names to skill tree IDs (simplified mapping)
  const unlockedSkillIds = getUnlockedSkills(agentClass, agentLevel, []);
  
  // Calculate grid dimensions
  const maxX = Math.max(...skillTree.skills.map(s => s.x)) + 1;
  const maxY = Math.max(...skillTree.skills.map(s => s.y)) + 1;
  
  const color = classColor(agentClass);
  
  const getSkillStatus = (skill: SkillNode): 'unlocked' | 'available' | 'locked' => {
    if (unlockedSkillIds.includes(skill.id)) return 'unlocked';
    if (isSkillUnlocked(skill, agentLevel, unlockedSkillIds)) return 'available';
    return 'locked';
  };
  
  const renderConnection = (from: SkillNode, to: SkillNode) => {
    const fromX = from.x * 120 + 60; // Center of skill node
    const fromY = from.y * 120 + 60;
    const toX = to.x * 120 + 60;
    const toY = to.y * 120 + 60;
    
    const isActive = getSkillStatus(to) !== 'locked';
    
    return (
      <line
        key={`${from.id}-${to.id}`}
        x1={fromX}
        y1={fromY}
        x2={toX}
        y2={toY}
        stroke={isActive ? color : '#4a5568'}
        strokeWidth={isActive ? 3 : 2}
        strokeOpacity={isActive ? 0.8 : 0.3}
        strokeDasharray={isActive ? 'none' : '5,5'}
      />
    );
  };
  
  return (
    <div className="skill-tree-container">
      <div className="skill-tree-header">
        <h3 style={{ color }}>
          {skillTree.className} Skill Tree
        </h3>
        <p className="skill-tree-description">{skillTree.description}</p>
        <div className="skill-tree-legend">
          <div className="legend-item">
            <div className="legend-icon unlocked" style={{ borderColor: color }}></div>
            <span>Unlocked</span>
          </div>
          <div className="legend-item">
            <div className="legend-icon available" style={{ borderColor: color }}></div>
            <span>Available</span>
          </div>
          <div className="legend-item">
            <div className="legend-icon locked"></div>
            <span>Locked</span>
          </div>
        </div>
      </div>
      
      <div className="skill-tree-wrapper">
        <div 
          className="skill-tree"
          style={{
            width: maxX * 120 + 'px',
            height: maxY * 120 + 'px',
            position: 'relative'
          }}
        >
          {/* Render connection lines */}
          <svg 
            className="skill-connections"
            width={maxX * 120}
            height={maxY * 120}
            style={{ position: 'absolute', top: 0, left: 0, pointerEvents: 'none' }}
          >
            {skillTree.skills.map(skill => 
              skill.prerequisites.map(prereqId => {
                const prereq = skillTree.skills.find(s => s.id === prereqId);
                return prereq ? renderConnection(prereq, skill) : null;
              })
            ).flat()}
          </svg>
          
          {/* Render skill nodes */}
          {skillTree.skills.map(skill => {
            const status = getSkillStatus(skill);
            const position = {
              left: skill.x * 120 + 'px',
              top: skill.y * 120 + 'px'
            };
            
            return (
              <Tooltip 
                key={skill.id} 
                content={`${skill.icon} ${skill.name}\n${skill.description}\nLevel Required: ${skill.levelRequired}\nCategory: ${skill.category}${skill.prerequisites.length > 0 ? '\nPrerequisites: ' + skill.prerequisites.map(id => {
                  const prereq = skillTree.skills.find(s => s.id === id);
                  return prereq?.name || id;
                }).join(', ') : ''}`}
              >
                <div
                  className={`skill-node ${status} ${skill.category} ${selectedSkill?.id === skill.id ? 'selected' : ''}`}
                  style={{
                    ...position,
                    borderColor: status === 'locked' ? '#4a5568' : color,
                    backgroundColor: status === 'unlocked' ? `${color}20` : undefined
                  }}
                  onClick={() => setSelectedSkill(selectedSkill?.id === skill.id ? null : skill)}
                >
                  <div className="skill-icon">{skill.icon}</div>
                  <div className="skill-name">{skill.name}</div>
                  {status === 'unlocked' && (
                    <div className="skill-checkmark" style={{ color }}>✓</div>
                  )}
                  {status === 'locked' && agentLevel < skill.levelRequired && (
                    <div className="skill-level-req">Lv {skill.levelRequired}</div>
                  )}
                </div>
              </Tooltip>
            );
          })}
        </div>
      </div>
      
      {/* Skill detail panel */}
      {selectedSkill && (
        <div className="skill-detail-panel" style={{ borderColor: color }}>
          <div className="skill-detail-header" style={{ backgroundColor: `${color}20` }}>
            <span className="skill-detail-icon">{selectedSkill.icon}</span>
            <h4>{selectedSkill.name}</h4>
            <button 
              className="skill-detail-close"
              onClick={() => setSelectedSkill(null)}
            >
              ×
            </button>
          </div>
          <div className="skill-detail-content">
            <p>{selectedSkill.description}</p>
            <div className="skill-detail-meta">
              <div className="meta-item">
                <strong>Level Required:</strong> {selectedSkill.levelRequired}
              </div>
              <div className="meta-item">
                <strong>Category:</strong> {selectedSkill.category}
              </div>
              {selectedSkill.prerequisites.length > 0 && (
                <div className="meta-item">
                  <strong>Prerequisites:</strong>
                  <ul>
                    {selectedSkill.prerequisites.map(id => {
                      const prereq = skillTree.skills.find(s => s.id === id);
                      const prereqStatus = prereq ? getSkillStatus(prereq) : 'locked';
                      return (
                        <li key={id} className={prereqStatus}>
                          {prereq ? (
                            <>
                              {prereq.icon} {prereq.name}
                              {prereqStatus === 'unlocked' && <span className="prereq-check">✓</span>}
                            </>
                          ) : id}
                        </li>
                      );
                    })}
                  </ul>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}