import React, { useEffect, useState, useRef } from 'react';
import type { ReactNode } from 'react';
import './MicroInteractions.css';

// Animation wrapper components
export const FadeIn: React.FC<{ children: ReactNode; delay?: boolean; className?: string }> = ({ 
  children, 
  delay = false, 
  className = '' 
}) => (
  <div className={`${delay ? 'fade-in-delayed' : 'fade-in'} ${className}`}>
    {children}
  </div>
);

export const SlideInLeft: React.FC<{ children: ReactNode; className?: string }> = ({ 
  children, 
  className = '' 
}) => (
  <div className={`slide-in-left ${className}`}>
    {children}
  </div>
);

export const SlideInRight: React.FC<{ children: ReactNode; className?: string }> = ({ 
  children, 
  className = '' 
}) => (
  <div className={`slide-in-right ${className}`}>
    {children}
  </div>
);

export const ScaleIn: React.FC<{ children: ReactNode; className?: string }> = ({ 
  children, 
  className = '' 
}) => (
  <div className={`scale-in ${className}`}>
    {children}
  </div>
);

export const BounceIn: React.FC<{ children: ReactNode; className?: string }> = ({ 
  children, 
  className = '' 
}) => (
  <div className={`bounce-in ${className}`}>
    {children}
  </div>
);

// Interactive hover wrapper
export const InteractiveHover: React.FC<{ 
  children: ReactNode; 
  scale?: boolean;
  glow?: boolean;
  className?: string;
}> = ({ 
  children, 
  scale = false, 
  glow = false, 
  className = '' 
}) => {
  const classes = [
    scale ? 'interactive-scale' : 'interactive-hover',
    glow ? 'glow-on-hover' : '',
    className
  ].filter(Boolean).join(' ');

  return (
    <div className={classes}>
      {children}
    </div>
  );
};

// Progress bar component
export interface ProgressBarProps {
  progress: number; // 0-100
  animated?: boolean;
  className?: string;
  size?: 'small' | 'medium' | 'large';
}

export const ProgressBar: React.FC<ProgressBarProps> = ({ 
  progress, 
  className = '',
  size = 'medium'
}) => {
  const sizeStyles = {
    small: { height: '2px' },
    medium: { height: '4px' },
    large: { height: '8px' }
  };

  return (
    <div 
      className={`progress-bar ${className}`}
      style={sizeStyles[size]}
    >
      <div 
        className="progress-fill"
        style={{ width: `${Math.max(0, Math.min(100, progress))}%` }}
      />
    </div>
  );
};

// Skeleton loader component
export interface SkeletonProps {
  width?: string | number;
  height?: string | number;
  className?: string;
  variant?: 'text' | 'circular' | 'rectangular';
}

export const Skeleton: React.FC<SkeletonProps> = ({ 
  width = '100%', 
  height = '20px', 
  className = '',
  variant = 'rectangular'
}) => {
  const style: React.CSSProperties = {
    width,
    height,
    borderRadius: variant === 'circular' ? '50%' : variant === 'text' ? '4px' : '8px'
  };

  return <div className={`skeleton ${className}`} style={style} />;
};

// Shimmer loading wrapper
export const ShimmerWrapper: React.FC<{ 
  children: ReactNode; 
  loading: boolean; 
  className?: string;
}> = ({ children, loading, className = '' }) => {
  if (loading) {
    return <div className={`shimmer ${className}`}>{children}</div>;
  }
  return <>{children}</>;
};

// State indicator component
export interface StateIndicatorProps {
  state: 'success' | 'error' | 'warning' | 'info' | 'default';
  children: ReactNode;
  animate?: boolean;
  className?: string;
}

export const StateIndicator: React.FC<StateIndicatorProps> = ({ 
  state, 
  children, 
  animate = true,
  className = '' 
}) => {
  const classes = [
    'state-transition',
    state !== 'default' ? `state-${state}` : '',
    animate && state === 'success' ? 'bounce-in' : '',
    animate && state === 'error' ? 'shake' : '',
    className
  ].filter(Boolean).join(' ');

  return <div className={classes}>{children}</div>;
};

// Enhanced tooltip component
export interface TooltipProps {
  content: string;
  children: ReactNode;
  placement?: 'top' | 'bottom' | 'left' | 'right';
  delay?: number;
  className?: string;
}

export const Tooltip: React.FC<TooltipProps> = ({ 
  content, 
  children, 
  placement = 'top',
  delay = 300,
  className = '' 
}) => {
  const [visible, setVisible] = useState(false);
  const timeoutRef = useRef<number | undefined>(undefined);

  const showTooltip = () => {
    timeoutRef.current = window.setTimeout(() => setVisible(true), delay);
  };

  const hideTooltip = () => {
    if (timeoutRef.current) {
      window.clearTimeout(timeoutRef.current);
    }
    setVisible(false);
  };

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        window.clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const placementClasses = {
    top: 'tooltip-top',
    bottom: 'tooltip-bottom', 
    left: 'tooltip-left',
    right: 'tooltip-right'
  };

  return (
    <div 
      className={`tooltip-container ${className}`}
      onMouseEnter={showTooltip}
      onMouseLeave={hideTooltip}
      onFocus={showTooltip}
      onBlur={hideTooltip}
    >
      {children}
      {visible && (
        <div className={`tooltip ${placementClasses[placement]}`}>
          {content}
        </div>
      )}
    </div>
  );
};

// Stagger animation for lists
export const StaggerList: React.FC<{ 
  children: ReactNode[]; 
  className?: string;
  itemClassName?: string;
}> = ({ children, className = '', itemClassName = '' }) => (
  <div className={className}>
    {React.Children.map(children, (child, index) => (
      <div key={index} className={`stagger-item ${itemClassName}`}>
        {child}
      </div>
    ))}
  </div>
);

// Typewriter effect component
export const Typewriter: React.FC<{ 
  text: string; 
  speed?: number;
  className?: string;
  onComplete?: () => void;
}> = ({ text, speed = 50, className = '', onComplete }) => {
  const [displayText, setDisplayText] = useState('');
  const [currentIndex, setCurrentIndex] = useState(0);

  useEffect(() => {
    if (currentIndex < text.length) {
      const timeout = setTimeout(() => {
        setDisplayText(prev => prev + text[currentIndex]);
        setCurrentIndex(prev => prev + 1);
      }, speed);
      return () => clearTimeout(timeout);
    } else if (onComplete) {
      onComplete();
    }
  }, [currentIndex, text, speed, onComplete]);

  useEffect(() => {
    setDisplayText('');
    setCurrentIndex(0);
  }, [text]);

  return (
    <span className={`typewriter ${className}`}>
      {displayText}
    </span>
  );
};

// Focus trap for modals and dialogs
export const FocusTrap: React.FC<{ children: ReactNode }> = ({ children }) => {
  const trapRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const trap = trapRef.current;
    if (!trap) return;

    const focusableElements = trap.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    
    const firstElement = focusableElements[0] as HTMLElement;
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;

      if (e.shiftKey) {
        if (document.activeElement === firstElement) {
          lastElement?.focus();
          e.preventDefault();
        }
      } else {
        if (document.activeElement === lastElement) {
          firstElement?.focus();
          e.preventDefault();
        }
      }
    };

    trap.addEventListener('keydown', handleTabKey);
    firstElement?.focus();

    return () => {
      trap.removeEventListener('keydown', handleTabKey);
    };
  }, []);

  return (
    <div ref={trapRef} className="focus-trap">
      {children}
    </div>
  );
};

// Floating action patterns
export const FloatingElement: React.FC<{ 
  children: ReactNode; 
  className?: string;
}> = ({ children, className = '' }) => (
  <div className={`floating ${className}`}>
    {children}
  </div>
);

// Enhanced button with ripple effect
export const RippleButton: React.FC<{
  children: ReactNode;
  onClick?: (e: React.MouseEvent) => void;
  className?: string;
  disabled?: boolean;
  [key: string]: any;
}> = ({ children, onClick, className = '', disabled, ...props }) => {
  const createRipple = (event: React.MouseEvent<HTMLButtonElement>) => {
    const button = event.currentTarget;
    const rect = button.getBoundingClientRect();
    
    const ripple = document.createElement('span');
    const size = Math.max(rect.width, rect.height);
    const x = event.clientX - rect.left - size / 2;
    const y = event.clientY - rect.top - size / 2;
    
    ripple.style.cssText = `
      position: absolute;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.6);
      transform: scale(0);
      animation: ripple-animation 0.6s linear;
      left: ${x}px;
      top: ${y}px;
      width: ${size}px;
      height: ${size}px;
      pointer-events: none;
      z-index: 1;
    `;
    
    button.appendChild(ripple);
    
    setTimeout(() => {
      ripple.remove();
    }, 600);
    
    if (onClick) {
      onClick(event);
    }
  };

  return (
    <button 
      className={`ripple-button ${className}`}
      onClick={createRipple}
      disabled={disabled}
      style={{ position: 'relative', overflow: 'hidden' }}
      {...props}
    >
      {children}
    </button>
  );
};

// Hook for animation visibility
export const useInViewAnimation = (threshold = 0.1) => {
  const [inView, setInView] = useState(false);
  const ref = useRef<HTMLElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setInView(true);
          observer.disconnect();
        }
      },
      { threshold }
    );

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => observer.disconnect();
  }, [threshold]);

  return { ref, inView };
};

// Global CSS injection for ripple animation
if (typeof document !== 'undefined') {
  const style = document.createElement('style');
  style.textContent = `
    @keyframes ripple-animation {
      to {
        transform: scale(4);
        opacity: 0;
      }
    }
  `;
  document.head.appendChild(style);
}