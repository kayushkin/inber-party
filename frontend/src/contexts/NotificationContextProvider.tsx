import { useState, type ReactNode } from 'react';
import type { Notification, NotificationContextType } from '../types/notification';
import { NotificationContext } from './NotificationContext';

interface NotificationProviderProps {
  children: ReactNode;
}

// Helper function to generate IDs outside of render
const generateId = () => Math.random().toString(36).substr(2, 9);

export const NotificationProvider = ({ children }: NotificationProviderProps) => {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const addNotification = (notification: Omit<Notification, 'id'>) => {
    const newNotification = {
      ...notification,
      id: generateId(),
    };
    setNotifications(prev => [...prev, newNotification]);

    // Auto-remove after duration (if specified and not 0)
    if (notification.duration && notification.duration > 0) {
      setTimeout(() => {
        removeNotification(newNotification.id);
      }, notification.duration);
    }
  };

  const removeNotification = (id: string) => {
    setNotifications(prev => prev.filter(notification => notification.id !== id));
  };

  const showSuccess = (title: string, message?: string) => {
    addNotification({ type: 'success', title, message, duration: 5000 });
  };

  const showError = (title: string, message?: string) => {
    addNotification({ type: 'error', title, message, duration: 8000 });
  };

  const showWarning = (title: string, message?: string) => {
    addNotification({ type: 'warning', title, message, duration: 6000 });
  };

  const showInfo = (title: string, message?: string) => {
    addNotification({ type: 'info', title, message, duration: 4000 });
  };

  const value: NotificationContextType = {
    notifications,
    addNotification,
    removeNotification,
    showSuccess,
    showError,
    showWarning,
    showInfo,
  };

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
};