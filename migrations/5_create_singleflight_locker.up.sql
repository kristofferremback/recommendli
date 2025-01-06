CREATE TABLE IF NOT EXISTS singleflight_locks (
  key TEXT NOT NULL,
  token TEXT NOT NULL,
  expires_at TEXT NULL,
  released_at TEXT NULL,
  released_by TEXT NULL,
  inserted_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (key, token),
  CONSTRAINT singleflight_locks_token_unique UNIQUE (token)
);
