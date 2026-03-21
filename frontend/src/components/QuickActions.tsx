import { useStore, classColor } from '../store';
import type { RPGAgent } from '../store';

interface QuickActionsProps {
  agent: RPGAgent;
  onActionClick?: () => void;
}

export default function QuickActions({ agent, onActionClick }: QuickActionsProps) {
  const setSelectedAgent = useStore((s) => s.setSelectedAgent);
  const cc = classColor(agent.class);

  const handleActionClick = (message: string) => {
    setSelectedAgent(agent.id);
    const sendMessage = useStore.getState().sendMessage;
    sendMessage(agent.id, message);
    onActionClick?.();
  };

  const actions = [
    {
      icon: '🔍',
      label: 'Scout Repo',
      message: 'Scout this repo - analyze the codebase structure, recent changes, and identify any immediate issues or opportunities for improvement.'
    },
    {
      icon: '🐛',
      label: 'Fix Bugs',
      message: 'Fix bugs - scan the codebase for any obvious bugs, linting errors, or issues that need immediate attention and fix them.'
    },
    {
      icon: '🧪',
      label: 'Write Tests',
      message: 'Write tests - identify areas of the codebase that need better test coverage and write comprehensive unit tests.'
    },
    {
      icon: '📚',
      label: 'Review Docs',
      message: 'Review documentation - check if README, docs, and code comments are up to date and comprehensive. Improve where needed.'
    },
    {
      icon: '🔧',
      label: 'Refactor',
      message: 'Refactor code - identify areas where code can be simplified, optimized, or made more maintainable without changing functionality.'
    },
    {
      icon: '🛡️',
      label: 'Security Check',
      message: 'Security audit - scan for potential security vulnerabilities, unsafe practices, or areas that need security hardening.'
    }
  ];

  return (
    <div className="section">
      <h3>Quick Actions</h3>
      <div className="quick-actions-grid">
        {actions.map((action) => (
          <button
            key={action.label}
            className="quick-action-btn"
            onClick={() => handleActionClick(action.message)}
            style={{ borderColor: cc }}
          >
            {action.icon} {action.label}
          </button>
        ))}
      </div>
    </div>
  );
}