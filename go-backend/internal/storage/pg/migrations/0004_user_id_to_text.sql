-- +goose Up

-- Convert users.id from BIGSERIAL to TEXT (UUID)
-- Step 1: Add new UUID column
ALTER TABLE users ADD COLUMN new_id TEXT;

-- Step 2: Generate UUIDs for existing rows
UPDATE users SET new_id = gen_random_uuid()::TEXT;

-- Step 3: Make new_id NOT NULL
ALTER TABLE users ALTER COLUMN new_id SET NOT NULL;

-- Step 4: Add new user_id column to api_keys
ALTER TABLE api_keys ADD COLUMN new_user_id TEXT;

-- Step 5: Update api_keys with new user IDs
UPDATE api_keys SET new_user_id = users.new_id FROM users WHERE api_keys.user_id = users.id;

-- Step 6: Drop old foreign key and constraints
ALTER TABLE api_keys DROP CONSTRAINT api_keys_user_id_fkey;
DROP INDEX idx_api_keys_user_id;

-- Step 7: Drop old columns
ALTER TABLE api_keys DROP COLUMN user_id;
ALTER TABLE api_keys RENAME COLUMN new_user_id TO user_id;
ALTER TABLE api_keys ALTER COLUMN user_id SET NOT NULL;

-- Step 8: Handle users table
ALTER TABLE users DROP CONSTRAINT users_pkey;
ALTER TABLE users DROP COLUMN id;
ALTER TABLE users RENAME COLUMN new_id TO id;
ALTER TABLE users ADD PRIMARY KEY (id);

-- Step 9: Recreate foreign key and index
ALTER TABLE api_keys ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

-- Step 10: Also convert api_keys.id to TEXT
ALTER TABLE api_keys ADD COLUMN new_key_id TEXT;
UPDATE api_keys SET new_key_id = gen_random_uuid()::TEXT;
ALTER TABLE api_keys ALTER COLUMN new_key_id SET NOT NULL;
ALTER TABLE api_keys DROP CONSTRAINT api_keys_pkey;
ALTER TABLE api_keys DROP COLUMN id;
ALTER TABLE api_keys RENAME COLUMN new_key_id TO id;
ALTER TABLE api_keys ADD PRIMARY KEY (id);

-- +goose Down
-- This migration is not reversible in a clean way
-- You would need to restore from backup
