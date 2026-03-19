-- Notifications system for inber-party
-- This schema supports the notification feature that alerts users about bounty events and skill matches

-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    recipient_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,  -- bounty_claimed, bounty_completed, bounty_disputed, skill_match, etc.
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,  -- Additional data specific to notification type
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',  -- low, normal, high, urgent
    status VARCHAR(20) NOT NULL DEFAULT 'unread',  -- unread, read, dismissed
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create notification preferences table (for future use)
CREATE TABLE IF NOT EXISTS notification_preferences (
    id SERIAL PRIMARY KEY,
    agent_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT true,
    websocket_enabled BOOLEAN DEFAULT true,
    push_enabled BOOLEAN DEFAULT false,
    type_preferences JSONB DEFAULT '{}',  -- Map of notification type to enabled/disabled
    skill_keywords TEXT[],  -- Keywords for skill matching
    minimum_bounty_tier VARCHAR(20) DEFAULT 'bronze',  -- minimum tier to notify about
    quiet_hours_start INTEGER,  -- 24-hour format, e.g., 22 for 10pm
    quiet_hours_end INTEGER,    -- 24-hour format, e.g., 8 for 8am
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(agent_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(recipient_id, status) WHERE status = 'unread';

-- Create indexes for notification preferences
CREATE INDEX IF NOT EXISTS idx_notification_preferences_agent ON notification_preferences(agent_id);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to auto-update updated_at
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_preferences_updated_at BEFORE UPDATE ON notification_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default notification preferences for existing agents
INSERT INTO notification_preferences (agent_id)
SELECT id FROM agents
WHERE id NOT IN (SELECT agent_id FROM notification_preferences)
ON CONFLICT (agent_id) DO NOTHING;