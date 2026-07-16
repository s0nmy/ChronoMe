ALTER TABLE users ADD COLUMN IF NOT EXISTS supabase_user_id uuid UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_migrated boolean DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_users_supabase_id ON users(supabase_user_id);
