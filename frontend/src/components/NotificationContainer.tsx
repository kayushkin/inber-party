import { useNotifications, type Notification } from '../contexts/NotificationContext';
import './NotificationContainer.css';

interface NotificationItemProps {
  notification: Notification;
  onClose: () => void;
}

const NotificationItem = ({ notification, onClose }: NotificationItemProps) => {
  const getIcon = () => {
    switch (notification.type) {
      case 'success': return '✅';
      case 'error': return '❌';
      case 'warning': return '⚠️';
      case 'info': return 'ℹ️';
      default: return '📢';
    }
  };

  return (
    <div className={`notification notification--${notification.type}`}>
      <div className="notification__icon">{getIcon()}</div>
      <div className="notification__content">
        <div className="notification__title">{notification.title}</div>
        {notification.message && (
          <div className="notification__message">{notification.message}</div>
        )}
      </div>
      <button 
        className="notification__close"
        onClick={onClose}
        aria-label="Close notification"
      >
        ×
      </button>
    </div>
  );
};

export const NotificationContainer = () => {
  const { notifications, removeNotification } = useNotifications();

  if (notifications.length === 0) {
    return null;
  }

  return (
    <div className="notification-container">
      {notifications.map(notification => (
        <NotificationItem
          key={notification.id}
          notification={notification}
          onClose={() => removeNotification(notification.id)}
        />
      ))}
    </div>
  );
};

export default NotificationContainer;