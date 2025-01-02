CREATE TABLE IF NOT EXISTS trackindex_tracks (
  -- key is a composite key of track name and artist names as the Spotify ID may differ
  -- accross different releases of a track, whereas I've found track name + artist names
  -- has works stupidly well for identifying tracks.
  key TEXT NOT NULL,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  simple_track JSONB NOT NULL,
  inserted_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (key, user_id)
);

CREATE TABLE IF NOT EXISTS trackindex_playlists (
  id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL,
  simple_playlist JSONB NOT NULL,
  snapshot_id TEXT NOT NULL,
  inserted_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (id, user_id)
);

CREATE TABLE IF NOT EXISTS trackindex_playlist_tracks (
  playlist_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  track_key TEXT NOT NULL,
  inserted_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (playlist_id, track_key, user_id)
);
