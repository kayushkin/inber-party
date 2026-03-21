import { useState } from 'react';
import { 
  FadeIn, 
  SlideInLeft, 
  SlideInRight, 
  ScaleIn, 
  BounceIn,
  InteractiveHover,
  ProgressBar,
  Skeleton,
  StateIndicator,
  Tooltip,
  StaggerList,
  Typewriter,
  FloatingElement
} from '../components/MicroInteractions';
import EnhancedButton, { 
  PrimaryButton, 
  SecondaryButton, 
  DangerButton, 
  SuccessButton, 
  GhostButton, 
  FloatingActionButton 
} from '../components/EnhancedButton';
import EnhancedAgentCard from '../components/EnhancedAgentCard';
import type { RPGAgent } from '../store';
import './UIShowcase.css';

// Mock agent data for demonstration
const mockAgents: RPGAgent[] = [
  {
    id: 'demo-1',
    name: 'Demo Wizard',
    title: 'Arcane Scholar',
    class: 'Wizard',
    level: 12,
    xp: 8500,
    xp_to_next: 1500,
    status: 'working',
    mood: 'happy',
    mood_score: 95,
    total_tokens: 125000,
    total_cost: 12.50,
    gold: 450,
    energy: 85,
    max_energy: 100,
    orchestrator: 'inber',
    session_count: 45,
    quest_count: 23,
    error_count: 2,
    skills: [
      { skill_name: 'Magic', level: 10, task_count: 15 },
      { skill_name: 'Research', level: 8, task_count: 20 }
    ],
    last_active: new Date().toISOString(),
    avatar_emoji: '🧙‍♂️'
  },
  {
    id: 'demo-2',
    name: 'Demo Warrior',
    title: 'Battle Hardened',
    class: 'Warrior',
    level: 8,
    xp: 3200,
    xp_to_next: 800,
    status: 'stuck',
    mood: 'stressed',
    mood_score: 25,
    total_tokens: 89000,
    total_cost: 8.90,
    gold: 280,
    energy: 30,
    max_energy: 100,
    orchestrator: 'inber',
    session_count: 32,
    quest_count: 18,
    error_count: 8,
    skills: [
      { skill_name: 'Combat', level: 9, task_count: 25 },
      { skill_name: 'Defense', level: 7, task_count: 12 }
    ],
    last_active: new Date(Date.now() - 30000).toISOString(),
    avatar_emoji: '⚔️'
  },
  {
    id: 'demo-3',
    name: 'Demo Rogue',
    title: 'Shadow Walker',
    class: 'Rogue',
    level: 15,
    xp: 12000,
    xp_to_next: 3000,
    status: 'on_quest',
    mood: 'content',
    mood_score: 80,
    total_tokens: 198000,
    total_cost: 19.80,
    gold: 680,
    energy: 75,
    max_energy: 100,
    orchestrator: 'inber',
    session_count: 67,
    quest_count: 34,
    error_count: 1,
    skills: [
      { skill_name: 'Stealth', level: 12, task_count: 30 },
      { skill_name: 'Lockpicking', level: 10, task_count: 18 },
      { skill_name: 'Agility', level: 11, task_count: 22 }
    ],
    last_active: new Date(Date.now() - 120000).toISOString(),
    avatar_emoji: '🗡️'
  }
];

export default function UIShowcase() {
  const [progress, setProgress] = useState(65);
  const [loading, setLoading] = useState(false);
  const [selectedAgent, setSelectedAgent] = useState<RPGAgent | null>(null);
  const [currentState, setCurrentState] = useState<'success' | 'error' | 'warning' | 'info' | 'default'>('default');
  const [showTypewriter, setShowTypewriter] = useState(false);

  const handleProgressChange = (value: number) => {
    setProgress(value);
  };

  const handleLoadingToggle = () => {
    setLoading(!loading);
    if (!loading) {
      setTimeout(() => setLoading(false), 3000);
    }
  };

  const handleStateChange = (state: typeof currentState) => {
    setCurrentState(state);
  };

  const handleAgentClick = (agent: RPGAgent) => {
    setSelectedAgent(agent);
  };

  return (
    <div className="ui-showcase">
      <div className="showcase-container">
        {/* Header */}
        <FadeIn>
          <header className="showcase-header">
            <h1 className="text-gradient">Enhanced UI/UX Showcase</h1>
            <p className="showcase-subtitle">
              Demonstrating improved visual design, animations, and user feedback
            </p>
          </header>
        </FadeIn>

        {/* Button Showcase */}
        <SlideInLeft>
          <section className="showcase-section">
            <h2>Enhanced Buttons</h2>
            <div className="button-grid">
              <PrimaryButton>Primary Action</PrimaryButton>
              <SecondaryButton>Secondary Action</SecondaryButton>
              <SuccessButton icon="✓">Success Action</SuccessButton>
              <DangerButton icon="⚠️">Danger Action</DangerButton>
              <GhostButton>Ghost Action</GhostButton>
              
              <EnhancedButton 
                loading={loading} 
                onClick={handleLoadingToggle}
                pulse={!loading}
              >
                {loading ? 'Loading...' : 'Toggle Loading'}
              </EnhancedButton>
              
              <EnhancedButton 
                glow 
                variant="success"
                icon="⭐"
              >
                Glowing Button
              </EnhancedButton>
              
              <EnhancedButton 
                size="large"
                variant="primary"
                icon="🚀"
                iconPosition="right"
              >
                Launch Mission
              </EnhancedButton>
            </div>
          </section>
        </SlideInLeft>

        {/* Progress & Loading States */}
        <SlideInRight>
          <section className="showcase-section">
            <h2>Progress Indicators & Loading States</h2>
            
            <div className="progress-demo">
              <h4>Interactive Progress Bar</h4>
              <ProgressBar progress={progress} animated />
              <div className="progress-controls">
                <input 
                  type="range" 
                  min="0" 
                  max="100" 
                  value={progress}
                  onChange={(e) => handleProgressChange(Number(e.target.value))}
                  className="progress-slider"
                />
                <span>{progress}%</span>
              </div>
            </div>

            <div className="skeleton-demo">
              <h4>Skeleton Loading States</h4>
              <div className="skeleton-grid">
                <Skeleton width="100%" height="20px" />
                <Skeleton width="80%" height="20px" />
                <Skeleton width="60%" height="20px" />
                <Skeleton width="40px" height="40px" variant="circular" />
              </div>
            </div>
          </section>
        </SlideInRight>

        {/* State Indicators */}
        <ScaleIn>
          <section className="showcase-section">
            <h2>State Indicators</h2>
            <div className="state-demo">
              <div className="state-buttons">
                <button onClick={() => handleStateChange('success')}>Success</button>
                <button onClick={() => handleStateChange('error')}>Error</button>
                <button onClick={() => handleStateChange('warning')}>Warning</button>
                <button onClick={() => handleStateChange('info')}>Info</button>
                <button onClick={() => handleStateChange('default')}>Default</button>
              </div>
              
              <StateIndicator state={currentState} animate>
                <div className="state-card">
                  <h4>Current State: {currentState}</h4>
                  <p>This card changes appearance based on the selected state.</p>
                </div>
              </StateIndicator>
            </div>
          </section>
        </ScaleIn>

        {/* Enhanced Agent Cards */}
        <section className="showcase-section">
          <h2>Enhanced Agent Cards</h2>
          <StaggerList className="agent-showcase-grid">
            {mockAgents.map(agent => (
              <EnhancedAgentCard
                key={agent.id}
                agent={agent}
                isSelected={selectedAgent?.id === agent.id}
                onClick={handleAgentClick}
                compact={false}
                animateIn={true}
              />
            ))}
          </StaggerList>
        </section>

        {/* Interactive Elements */}
        <BounceIn>
          <section className="showcase-section">
            <h2>Interactive Elements</h2>
            
            <div className="interactive-demo">
              <Tooltip content="This is a helpful tooltip!">
                <InteractiveHover scale>
                  <div className="demo-card">
                    <h4>Hover for Scale Effect</h4>
                    <p>This card scales on hover and has a tooltip.</p>
                  </div>
                </InteractiveHover>
              </Tooltip>

              <Tooltip content="Click to experience glow effect">
                <InteractiveHover glow>
                  <div className="demo-card">
                    <h4>Glow on Hover</h4>
                    <p>This card glows when you hover over it.</p>
                  </div>
                </InteractiveHover>
              </Tooltip>

              <FloatingElement>
                <div className="demo-card">
                  <h4>Floating Animation</h4>
                  <p>This card gently floats up and down.</p>
                </div>
              </FloatingElement>
            </div>
          </section>
        </BounceIn>

        {/* Typography Effects */}
        <FadeIn delay>
          <section className="showcase-section">
            <h2>Typography Effects</h2>
            
            <div className="typography-demo">
              <button 
                className="btn"
                onClick={() => setShowTypewriter(!showTypewriter)}
              >
                Toggle Typewriter Effect
              </button>
              
              {showTypewriter && (
                <div className="typewriter-demo">
                  <Typewriter 
                    text="Welcome to the enhanced Inber Party experience! These improvements provide better visual feedback and smoother interactions."
                    speed={30}
                    onComplete={() => console.log('Typewriter complete!')}
                  />
                </div>
              )}
            </div>
          </section>
        </FadeIn>

        {/* Theme Examples */}
        <section className="showcase-section">
          <h2>Enhanced Styling</h2>
          <div className="theme-demo">
            <div className="card">
              <h4>Enhanced Card</h4>
              <p>Cards now have better depth and hover effects with backdrop blur.</p>
            </div>
            
            <div className="form-demo">
              <div className="form-group">
                <label className="form-label">Enhanced Form Input</label>
                <input 
                  type="text" 
                  className="form-input focus-ring" 
                  placeholder="Type something..."
                />
              </div>
            </div>
          </div>
        </section>
      </div>

      {/* Floating Action Button */}
      <FloatingActionButton 
        icon="🎨" 
        onClick={() => alert('Enhanced UI/UX activated!')}
        title="UI Enhancements Active"
      />
    </div>
  );
}