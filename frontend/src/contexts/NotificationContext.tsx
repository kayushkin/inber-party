import { createContext } from 'react';
import type { NotificationContextType } from '../types/notification';

export const NotificationContext = createContext<NotificationContextType | undefined>(undefined);