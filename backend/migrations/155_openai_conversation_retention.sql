CREATE TABLE IF NOT EXISTS openai_retained_sessions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  api_key_id BIGINT NOT NULL,
  account_id BIGINT NULL,
  group_id BIGINT NULL,
  session_key VARCHAR(128) NOT NULL,
  client_session_id TEXT NOT NULL DEFAULT '',
  requested_model VARCHAR(100) NOT NULL DEFAULT '',
  upstream_model VARCHAR(100) NOT NULL DEFAULT '',
  reasoning_effort VARCHAR(20) NOT NULL DEFAULT '',
  first_request_id VARCHAR(64) NOT NULL DEFAULT '',
  last_request_id VARCHAR(64) NOT NULL DEFAULT '',
  first_response_id VARCHAR(128) NOT NULL DEFAULT '',
  last_response_id VARCHAR(128) NOT NULL DEFAULT '',
  turn_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_openai_retained_sessions_session_key
  ON openai_retained_sessions (session_key);

CREATE INDEX IF NOT EXISTS idx_openai_retained_sessions_user_updated_at
  ON openai_retained_sessions (user_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_openai_retained_sessions_client_session_id
  ON openai_retained_sessions (client_session_id);

CREATE INDEX IF NOT EXISTS idx_openai_retained_sessions_model_effort_updated_at
  ON openai_retained_sessions (requested_model, reasoning_effort, updated_at DESC);

CREATE TABLE IF NOT EXISTS openai_retained_turns (
  id BIGSERIAL PRIMARY KEY,
  session_id BIGINT NOT NULL REFERENCES openai_retained_sessions(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL,
  api_key_id BIGINT NOT NULL,
  account_id BIGINT NULL,
  group_id BIGINT NULL,
  request_id VARCHAR(64) NOT NULL,
  response_id VARCHAR(128) NOT NULL DEFAULT '',
  turn_index INTEGER NOT NULL,
  requested_model VARCHAR(100) NOT NULL DEFAULT '',
  upstream_model VARCHAR(100) NOT NULL DEFAULT '',
  reasoning_effort VARCHAR(20) NOT NULL DEFAULT '',
  stream BOOLEAN NOT NULL DEFAULT FALSE,
  request_type SMALLINT NOT NULL DEFAULT 0,
  inbound_endpoint VARCHAR(128) NOT NULL DEFAULT '',
  upstream_endpoint VARCHAR(128) NOT NULL DEFAULT '',
  user_input_text TEXT NOT NULL DEFAULT '',
  assistant_output_text TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_openai_retained_turns_request_id
  ON openai_retained_turns (request_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_openai_retained_turns_session_turn_index
  ON openai_retained_turns (session_id, turn_index);

CREATE INDEX IF NOT EXISTS idx_openai_retained_turns_session_created_at
  ON openai_retained_turns (session_id, created_at ASC);

CREATE INDEX IF NOT EXISTS idx_openai_retained_turns_user_created_at
  ON openai_retained_turns (user_id, created_at DESC);
