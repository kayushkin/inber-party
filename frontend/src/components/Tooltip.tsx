import { type ReactNode, useState, useRef } from 'react';
import './Tooltip.css';

interface TooltipProps {
  content: string;
  children: ReactNode;
  position?: 'top' | 'bottom' | 'left' | 'right';
  maxWidth?: string;
}

export default function Tooltip({ content, children, position = 'top', maxWidth = '200px' }: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 });
  const timeoutRef = useRef<number | null>(null);

  const handleMouseEnter = (e: React.MouseEvent) => {
    setMousePosition({ x: e.clientX, y: e.clientY });
    if (timeoutRef.current) {
      window.clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = window.setTimeout(() => setIsVisible(true), 300); // Delay to avoid flickering
  };

  const handleMouseLeave = () => {
    if (timeoutRef.current) {
      window.clearTimeout(timeoutRef.current);
    }
    setIsVisible(false);
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    setMousePosition({ x: e.clientX, y: e.clientY });
  };

  return (
    <div
      className="tooltip-wrapper"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      onMouseMove={handleMouseMove}
      style={{ display: 'inline-block' }}
    >
      {children}
      {isVisible && (
        <div
          className={`tooltip tooltip-${position}`}
          style={{
            maxWidth,
            position: 'fixed',
            left: `${mousePosition.x}px`,
            top: `${mousePosition.y - 40}px`,
            zIndex: 1000,
          }}
        >
          {content}
        </div>
      )}
    </div>
  );
}