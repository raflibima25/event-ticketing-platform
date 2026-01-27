-- Drop indexes first
DROP INDEX IF EXISTS idx_reset_tokens_expires;
DROP INDEX IF EXISTS idx_reset_tokens_user;
DROP INDEX IF EXISTS idx_reset_tokens_token;

-- Drop table
DROP TABLE IF EXISTS password_reset_tokens;
