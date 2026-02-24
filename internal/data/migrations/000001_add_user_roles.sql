
ALTER TABLE users ADD COLUMN IF NOT EXISTS user_role text NOT NULL DEFAULT 'user';

UPDATE users SET user_role = 'admin' WHERE email = 'techie@example.com';