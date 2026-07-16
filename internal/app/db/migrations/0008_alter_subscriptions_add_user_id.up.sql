ALTER TABLE subscriptions ADD COLUMN user_id BIGINT REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);
DROP INDEX IF EXISTS idx_subscriptions_email_repo;
CREATE UNIQUE INDEX idx_subscriptions_user_repo ON subscriptions (user_id, repo) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_subscriptions_email_repo_api ON subscriptions (email, repo) WHERE user_id IS NULL;
