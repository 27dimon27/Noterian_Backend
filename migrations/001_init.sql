CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password BYTEA NOT NULL,
    token_version INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS avatars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL UNIQUE REFERENCES profiles(id) ON DELETE CASCADE,
    minio_key VARCHAR(255) NOT NULL,
    avatar_url TEXT NOT NULL UNIQUE,
    url_expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS block_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO block_types (name) VALUES 
    ('text'),
    ('image'),
    ('code'),
    ('quote'),
    ('subnote'),
    ('music'),
    ('video')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES notes(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
    icon VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    block_type_id INTEGER NOT NULL REFERENCES block_types(id),
    position INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS block_formatting (
    block_id UUID NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    start_pos INTEGER NOT NULL,
    end_pos INTEGER NOT NULL,
    bold BOOLEAN NOT NULL DEFAULT FALSE,
    italic BOOLEAN NOT NULL DEFAULT FALSE,
    underline BOOLEAN NOT NULL DEFAULT FALSE,
    text_align INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT check_positions CHECK (start_pos >= 0 AND end_pos > start_pos),
    CONSTRAINT check_text_align CHECK (text_align IS NULL OR (text_align >= 0 AND text_align <= 2))
);

CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY,
    block_id UUID NOT NULL UNIQUE REFERENCES blocks(id) ON DELETE CASCADE,
    minio_key VARCHAR(255) NOT NULL,
    attach_url TEXT NOT NULL UNIQUE,
    url_expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS headers (
    id UUID PRIMARY KEY,
    note_id UUID NOT NULL UNIQUE REFERENCES notes(id) ON DELETE CASCADE,
    minio_key VARCHAR(255) NOT NULL,
    header_url TEXT NOT NULL UNIQUE,
    url_expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_parent_id ON notes(parent_id);
CREATE INDEX idx_blocks_note_id ON blocks(note_id);
CREATE INDEX idx_blocks_note_position ON blocks(note_id, position);
CREATE INDEX idx_block_formatting_block_id ON block_formatting(block_id);
CREATE INDEX idx_block_formatting_positions ON block_formatting(block_id, start_pos, end_pos);
CREATE INDEX idx_attachments_block_id ON attachments(block_id);
CREATE INDEX idx_attachments_created_at ON attachments(created_at DESC);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_profiles_updated_at 
    BEFORE UPDATE ON profiles 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notes_updated_at 
    BEFORE UPDATE ON notes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_blocks_updated_at 
    BEFORE UPDATE ON blocks 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_block_formatting_updated_at 
    BEFORE UPDATE ON block_formatting 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();