CREATE TABLE secrets (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     title VARCHAR(255) NOT NULL,
     type VARCHAR(20) NOT NULL,
     encrypted_data TEXT NOT NULL,
     metadata JSONB NOT NULL DEFAULT '{}',
     file_path VARCHAR(512),
     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_secrets_user_id ON secrets(user_id);