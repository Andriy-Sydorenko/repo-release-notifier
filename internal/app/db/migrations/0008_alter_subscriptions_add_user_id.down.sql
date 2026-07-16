DROP INDEX IF EXISTS idx_subscriptions_email_repo_api;
DROP INDEX IF EXISTS idx_subscriptions_user_repo;
CREATE UNIQUE INDEX idx_subscriptions_email_repo ON subscriptions (email, repo);
DROP INDEX IF EXISTS idx_subscriptions_user_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS user_id;
