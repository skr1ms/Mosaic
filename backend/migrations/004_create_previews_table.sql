-- Create previews table for storing generated preview data
CREATE TABLE IF NOT EXISTS previews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url VARCHAR(500) NOT NULL,
    style VARCHAR(100),
    contrast VARCHAR(50),
    size VARCHAR(20),
    session_id VARCHAR(255),
    user_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Create indexes for better query performance
CREATE INDEX idx_previews_session_id ON previews(session_id);
CREATE INDEX idx_previews_user_id ON previews(user_id);
CREATE INDEX idx_previews_expires_at ON previews(expires_at);
CREATE INDEX idx_previews_created_at ON previews(created_at DESC);

-- Add foreign key constraint if users table exists
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        ALTER TABLE previews 
        ADD CONSTRAINT fk_previews_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE;
    END IF;
END $$;