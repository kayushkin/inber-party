import { useStore } from '../store';
import './ThemeToggle.css';

export default function ThemeToggle() {
  const theme = useStore((s) => s.theme);
  const toggleTheme = useStore((s) => s.toggleTheme);

  return (
    <button
      className="theme-toggle"
      onClick={toggleTheme}
      title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
      aria-label={`Toggle to ${theme === 'dark' ? 'light' : 'dark'} theme`}
    >
      <span className="theme-icon">
        {theme === 'dark' ? '☀️' : '🌙'}
      </span>
    </button>
  );
}