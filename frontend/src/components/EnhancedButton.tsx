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
    // Extract valid anchor props - cast to any to access HTML attributes
    const allProps = props as Record<string, unknown>;
    
    const anchorProps = {
      id: allProps.id as string | undefined,
      'aria-label': allProps['aria-label'] as string | undefined,
      'aria-describedby': allProps['aria-describedby'] as string | undefined,
      'data-testid': allProps['data-testid'] as string | undefined,
      target: allProps.target as string | undefined,
      rel: allProps.rel as string | undefined,
      download: allProps.download as string | undefined
    };
    
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
        {...anchorProps}
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