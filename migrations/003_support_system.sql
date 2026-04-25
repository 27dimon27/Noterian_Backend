CREATE TABLE IF NOT EXISTS support_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO support_categories (name, description) VALUES
    ('bug', 'Report a bug or technical issue'),
    ('suggestion', 'Feature suggestion or improvement idea'),
    ('complaint', 'Product complaint or quality issue'),
    ('other', 'Other support request')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS support_statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO support_statuses (name, description) VALUES
    ('open', 'Ticket is open and awaiting response'),
    ('in_progress', 'Support team is working on this ticket'),
    ('waiting_user', 'Waiting for user response'),
    ('closed', 'Ticket has been resolved and closed'),
    ('reopened', 'Closed ticket has been reopened')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO user_roles (name, description) VALUES
    ('user', 'Regular user'),
    ('support_tier1', 'First tier support'),
    ('support_tier2', 'Second tier support'),
    ('admin', 'Administrator')
ON CONFLICT (name) DO NOTHING;

ALTER TABLE profiles ADD COLUMN IF NOT EXISTS role_id INTEGER DEFAULT 1 REFERENCES user_roles(id);

CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES support_categories(id),
    status_id INTEGER NOT NULL REFERENCES support_statuses(id),
    assigned_to UUID REFERENCES profiles(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    priority INTEGER DEFAULT 2, -- 1 - самый высокий, 5 - самый низкий
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS support_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS support_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL UNIQUE REFERENCES support_tickets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_user_id ON support_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assigned_to ON support_tickets(assigned_to);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status_id ON support_tickets(status_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_category_id ON support_tickets(category_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_created_at ON support_tickets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_messages_ticket_id ON support_messages(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_messages_user_id ON support_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_support_messages_created_at ON support_messages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_ratings_ticket_id ON support_ratings(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_ratings_user_id ON support_ratings(user_id);

CREATE TRIGGER update_support_tickets_updated_at 
    BEFORE UPDATE ON support_tickets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_support_messages_updated_at 
    BEFORE UPDATE ON support_messages 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_support_ratings_updated_at 
    BEFORE UPDATE ON support_ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
