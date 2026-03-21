import React, { forwardRef } from 'react';
import type { ReactNode, ButtonHTMLAttributes } from 'react';
import './EnhancedButton.css';

export interface EnhancedButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'size'> {
  variant?: 'primary' | 'secondary' | 'danger' | 'success' | 'ghost';
  size?: 'small' | 'medium' | 'large' | 'icon';
  loading?: boolean;
  icon?: ReactNode;
  iconPosition?: 'left' | 'right';
  pulse?: boolean;
  glow?: boolean;
  fab?: boolean;
  children?: ReactNode;
  href?: string; // For link-style buttons
}

const EnhancedButton = forwardRef<HTMLButtonElement, EnhancedButtonProps>(({
  variant = 'primary',
  size = 'medium',
  loading = false,
  icon,
  iconPosition = 'left',
  pulse = false,
  glow = false,
  fab = false,
  className = '',
  children,
  href,
  disabled,
  onClick,
  ...props
}, ref) => {
  const classes = [
    'enhanced-button',
    variant !== 'primary' ? `variant-${variant}` : '',
    size !== 'medium' ? `size-${size}` : '',
    loading ? 'loading' : '',
    pulse ? 'pulse' : '',
    glow ? 'glow' : '',
    fab ? 'fab' : '',
    className
  ].filter(Boolean).join(' ');

  const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    if (loading || disabled) {
      e.preventDefault();
      return;
    }
    
    // Add ripple effect coordinates for better UX
    const button = e.currentTarget;
    const rect = button.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    
    // Store ripple position for CSS animation
    button.style.setProperty('--ripple-x', `${x}px`);
    button.style.setProperty('--ripple-y', `${y}px`);
    
    if (onClick) {
      onClick(e);
    }
  };

  const content = (
    <>
      {loading && <div className="enhanced-button-spinner" aria-hidden="true" />}
      
      {!fab && icon && iconPosition === 'left' && (
        <span className="enhanced-button-icon" aria-hidden="true">
          {icon}
        </span>
      )}
      
      {fab && icon ? (
        <span className="enhanced-button-icon icon-only" aria-hidden="true">
          {icon}
        </span>
      ) : (
        children && (
          <span className="button-text">
            {children}
          </span>
        )
      )}
      
      {!fab && icon && iconPosition === 'right' && (
        <span className="enhanced-button-icon" aria-hidden="true">
          {icon}
        </span>
      )}
    </>
  );

  // If href is provided, render as anchor tag with button styling
  if (href && !disabled && !loading) {
    return (
      <a
        href={href}
        className={classes}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            window.location.href = href;
          }
        }}
        {...(props as any)} // Type assertion needed for anchor props
      >
        {content}
      </a>
    );
  }

  return (
    <button
      ref={ref}
      className={classes}
      disabled={disabled || loading}
      onClick={handleClick}
      aria-busy={loading}
      {...props}
    >
      {content}
    </button>
  );
});

EnhancedButton.displayName = 'EnhancedButton';

export default EnhancedButton;

// Utility components for common button patterns
export const PrimaryButton = (props: Omit<EnhancedButtonProps, 'variant'>) => (
  <EnhancedButton variant="primary" {...props} />
);

export const SecondaryButton = (props: Omit<EnhancedButtonProps, 'variant'>) => (
  <EnhancedButton variant="secondary" {...props} />
);

export const DangerButton = (props: Omit<EnhancedButtonProps, 'variant'>) => (
  <EnhancedButton variant="danger" {...props} />
);

export const SuccessButton = (props: Omit<EnhancedButtonProps, 'variant'>) => (
  <EnhancedButton variant="success" {...props} />
);

export const GhostButton = (props: Omit<EnhancedButtonProps, 'variant'>) => (
  <EnhancedButton variant="ghost" {...props} />
);

export const FloatingActionButton = (props: Omit<EnhancedButtonProps, 'fab'>) => (
  <EnhancedButton fab {...props} />
);

// Hook for programmatic ripple effects
export const useRippleEffect = () => {
  const createRipple = (element: HTMLElement, x: number, y: number) => {
    const ripple = document.createElement('span');
    ripple.style.cssText = `
      position: absolute;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.6);
      transform: translate(-50%, -50%) scale(0);
      animation: ripple-effect 0.6s ease-out;
      pointer-events: none;
      left: ${x}px;
      top: ${y}px;
      width: 100px;
      height: 100px;
    `;
    
    element.appendChild(ripple);
    
    setTimeout(() => {
      ripple.remove();
    }, 600);
  };
  
  return { createRipple };
};