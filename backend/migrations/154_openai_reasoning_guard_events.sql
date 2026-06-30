CREATE TABLE IF NOT EXISTS openai_reasoning_guard_events (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  api_key_id BIGINT NOT NULL,
  account_id BIGINT NOT NULL,
  group_id BIGINT NULL,
  subscription_id BIGINT NULL,
  request_id TEXT NOT NULL DEFAULT '',
  requested_model TEXT NOT NULL DEFAULT '',
  upstream_model TEXT NOT NULL DEFAULT '',
  inbound_endpoint TEXT NOT NULL DEFAULT '',
  upstream_endpoint TEXT NOT NULL DEFAULT '',
  service_tier TEXT NOT NULL DEFAULT '',
  reasoning_effort TEXT NOT NULL DEFAULT '',
  reasoning_tokens INTEGER NOT NULL DEFAULT 0,
  has_reasoning_tokens BOOLEAN NOT NULL DEFAULT FALSE,
  matched_reasoning_code INTEGER NULL,
  intercepted BOOLEAN NOT NULL DEFAULT FALSE,
  response_status_code INTEGER NOT NULL DEFAULT 0,
  stream BOOLEAN NOT NULL DEFAULT FALSE,
  openai_ws_mode BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_openai_reasoning_guard_events_user_created_at
  ON openai_reasoning_guard_events (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_openai_reasoning_guard_events_user_intercepted_created_at
  ON openai_reasoning_guard_events (user_id, intercepted, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_openai_reasoning_guard_events_user_requested_model_created_at
  ON openai_reasoning_guard_events (user_id, requested_model, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_openai_reasoning_guard_events_user_reasoning_effort_created_at
  ON openai_reasoning_guard_events (user_id, reasoning_effort, created_at DESC);
