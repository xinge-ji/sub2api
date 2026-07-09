ALTER TABLE usage_logs
  ADD COLUMN IF NOT EXISTS upstream_user_agent VARCHAR(512);
