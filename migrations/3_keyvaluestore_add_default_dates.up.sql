CREATE TABLE IF NOT EXISTS keyvaluestore_tmp (
  key TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  value JSONB NOT NULL,
  inserted_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO keyvaluestore_tmp (key, kind, value, inserted_at, updated_at)
SELECT key,
  kind,
  value,
  COALESCE(inserted_at, datetime('now')),
  COALESCE(updated_at, datetime('now'))
FROM keyvaluestore;

DROP TABLE keyvaluestore;

ALTER TABLE keyvaluestore_tmp RENAME TO keyvaluestore;
