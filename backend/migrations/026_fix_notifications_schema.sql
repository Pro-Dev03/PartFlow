-- Fix notifications schema to match the code expectations
-- This migration updates the notifications table to support the full notification system

BEGIN;

-- Add new columns
ALTER TABLE notifications 
ADD COLUMN IF NOT EXISTS priority VARCHAR(50) DEFAULT 'medium',
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'unread',
ADD COLUMN IF NOT EXISTS action_url VARCHAR(500),
ADD COLUMN IF NOT EXISTS action_text VARCHAR(100),
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS read_at TIMESTAMP WITH TIME ZONE;

-- Migrate existing data from is_read to status
UPDATE notifications 
SET status = CASE WHEN is_read = true THEN 'read' ELSE 'unread' END
WHERE status IS NULL OR status = 'unread';

-- Drop the old is_read column after migration
ALTER TABLE notifications DROP COLUMN IF EXISTS is_read;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_priority ON notifications(priority);
CREATE INDEX IF NOT EXISTS idx_notifications_expires_at ON notifications(expires_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status);

-- Create notification_preferences table if it doesn't exist
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT true,
    push_enabled BOOLEAN DEFAULT true,
    low_stock BOOLEAN DEFAULT true,
    debt_overdue BOOLEAN DEFAULT true,
    return_requests BOOLEAN DEFAULT true,
    expense_approval BOOLEAN DEFAULT true,
    sales_updates BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_preferences_user ON notification_preferences(user_id);

COMMIT;