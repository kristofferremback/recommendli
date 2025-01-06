package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/kristofferostlund/recommendli/internal/recommendations"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/kristofferostlund/recommendli/pkg/spotifyutil"
	"github.com/zmb3/spotify"
)

var _ recommendations.TrackIndex = (*TrackIndex)(nil)

type TrackIndex struct {
	db          *DB
	trackIDFunc func(spotify.SimpleTrack) string
}

type Querier interface {
	sqlx.Queryer
	sqlx.QueryerContext
	sqlx.Execer
	sqlx.ExecerContext
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
}

func NewTrackIndex(db *DB, trackIDFunc func(spotify.SimpleTrack) string) *TrackIndex {
	return &TrackIndex{
		db:          db,
		trackIDFunc: trackIDFunc,
	}
}

func (t *TrackIndex) Has(ctx context.Context, userID string, track spotify.SimpleTrack) (bool, error) {
	playlists, err := t.Lookup(ctx, userID, track)
	if err != nil {
		return false, fmt.Errorf("looking up track in index: %w", err)
	}
	return len(playlists) > 0, nil
}

func (t *TrackIndex) Lookup(ctx context.Context, userID string, track spotify.SimpleTrack) ([]spotify.SimplePlaylist, error) {
	db, release := t.db.RGet(ctx)
	defer release()

	values := map[string]any{
		"track_key": t.trackIDFunc(track),
		"user_id":   userID,
	}

	rows, err := db.NamedQueryContext(ctx, `
		SELECT tp.simple_playlist
		FROM trackindex_playlists AS tp
		INNER JOIN
			trackindex_playlist_tracks AS tpt
			ON tp.id = tpt.playlist_id
				AND tp.user_id = tpt.user_id
		WHERE
			tpt.track_key = :track_key
			AND tp.user_id = :user_id
	`, values)
	if err != nil {
		return nil, fmt.Errorf("looking up track playlists for track %s (%s): %w", track.Name, track.ID, err)
	}
	defer rows.Close()

	var playlists []spotify.SimplePlaylist
	for rows.Next() {
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, fmt.Errorf("scanning track index: %w", err)
		}

		var playlist spotify.SimplePlaylist
		if err := json.Unmarshal(b, &playlist); err != nil {
			return nil, fmt.Errorf("unmarshalling playlist: %w", err)
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func (t *TrackIndex) Diff(ctx context.Context, userID string, playlists []spotify.SimplePlaylist) (added, changed, removed []spotify.SimplePlaylist, err error) {
	db, release := t.db.RGet(ctx)
	defer release()

	rows, err := db.NamedQueryContext(ctx, `
		SELECT simple_playlist
		FROM trackindex_playlists
		WHERE user_id = :user_id
	`, map[string]any{"user_id": userID})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("querying track index playlists to diff: %w", err)
	}
	defer rows.Close()

	prev := make(map[string]spotify.SimplePlaylist)
	for rows.Next() {
		var b []byte

		if err := rows.Scan(&b); err != nil {
			return nil, nil, nil, fmt.Errorf("scanning track index: %w", err)
		}

		var playlist spotify.SimplePlaylist
		if err := json.Unmarshal(b, &playlist); err != nil {
			return nil, nil, nil, fmt.Errorf("unmarshalling playlist: %w", err)
		}
		prev[playlist.ID.String()] = playlist
	}

	var addedPlaylists, changedPlaylists, removedPlaylists []spotify.SimplePlaylist

	next := make(map[string]spotify.SimplePlaylist)
	for _, p := range playlists {
		next[p.ID.String()] = p
	}

	for id, n := range next {
		if o, exists := prev[id]; !exists {
			addedPlaylists = append(addedPlaylists, n)
		} else if spotifyutil.SimplePlaylistHasChanged(n, o) {
			changedPlaylists = append(changedPlaylists, n)
		}
	}

	for id, p := range prev {
		if _, exists := next[id]; !exists {
			removedPlaylists = append(removedPlaylists, p)
		}
	}

	return addedPlaylists, changedPlaylists, removedPlaylists, nil
}

func (t *TrackIndex) Sync(ctx context.Context, userID string, added, changed, removed []spotify.FullPlaylist) error {
	ctx = slogutil.WithAttrs(ctx, slog.String("user", userID), slog.Int("added", len(added)), slog.Int("changed", len(changed)), slog.Int("removed", len(removed)))
	slog.DebugContext(ctx, "syncing track index")

	db, release := t.db.Get(ctx)
	defer release()

	slog.DebugContext(ctx, "beginning tx")

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("beginning tx: %w", err)
	}
	defer tx.Rollback()

	slog.DebugContext(ctx, "inserting added playlists")
	for _, playlist := range added {
		if err := t.addPlaylistToTrackIndex(ctx, tx, userID, playlist); err != nil {
			return fmt.Errorf("adding playlist %s (%s) to track index: %w", playlist.Name, playlist.ID, err)
		}
	}

	slog.DebugContext(ctx, "inserting changed playlists")
	for _, playlist := range changed {
		// I can improve this, but for now I don't think it's worth it.
		// A remove and re-add is a simple way to handle this.
		slog.DebugContext(ctx, "removing updated playlist so it can be re-added", slog.String("playlist_id", playlist.ID.String()), slog.String("playlist_name", playlist.Name))
		if err := removeTrackIndexPlaylist(ctx, tx, userID, playlist.ID); err != nil {
			return fmt.Errorf("removing playlist %s (%s) from track index: %w", playlist.Name, playlist.ID, err)
		}

		slog.DebugContext(ctx, "adding updated playlist again", slog.String("playlist_id", playlist.ID.String()), slog.String("playlist_name", playlist.Name))
		if err := t.addPlaylistToTrackIndex(ctx, tx, userID, playlist); err != nil {
			return fmt.Errorf("re-adding playlist %s (%s) to track index: %w", playlist.Name, playlist.ID, err)
		}
	}

	slog.DebugContext(ctx, "removing removed playlists")
	for _, playlist := range removed {
		slog.DebugContext(ctx, "removing playlist", slog.String("playlist_id", playlist.ID.String()), slog.String("playlist_name", playlist.Name))
		if err := removeTrackIndexPlaylist(ctx, tx, userID, playlist.ID); err != nil {
			return fmt.Errorf("removing playlist %s (%s) from track index: %w", playlist.Name, playlist.ID, err)
		}
	}

	slog.DebugContext(ctx, "committing tx")

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing tx: %w", err)
	}

	slog.DebugContext(ctx, "synced track index")

	return nil
}

func (t *TrackIndex) addPlaylistToTrackIndex(ctx context.Context, q Querier, userID string, playlist spotify.FullPlaylist) error {
	if err := insertOrReplaceTrackIndexPlaylists(ctx, q, userID, playlist.SimplePlaylist); err != nil {
		return fmt.Errorf("inserting playlist %s (%s): %w", playlist.Name, playlist.ID, err)
	}

	simpleTracks := make([]spotify.SimpleTrack, 0, len(playlist.Tracks.Tracks))
	for _, track := range playlist.Tracks.Tracks {
		simpleTracks = append(simpleTracks, track.Track.SimpleTrack)
	}

	if err := t.insertTrackIndexTrackOnPlaylist(ctx, q, userID, playlist.ID.String(), simpleTracks); err != nil {
		return fmt.Errorf("inserting tracks for playlist %s (%s): %w", playlist.Name, playlist.ID, err)
	}

	return nil
}

func removeTrackIndexPlaylist(ctx context.Context, q Querier, userID string, playlistID spotify.ID) error {
	// Remove the playlist
	if _, err := q.NamedExecContext(ctx, `
		DELETE FROM trackindex_playlists
		WHERE id = :id AND user_id = :user_id
	`, map[string]any{"id": playlistID, "user_id": userID}); err != nil {
		return fmt.Errorf("deleting playlist from track index: %w", err)
	}

	// Remove playlist tracks
	if _, err := q.NamedExecContext(ctx, `
		DELETE FROM trackindex_playlist_tracks
		WHERE playlist_id = :playlist_id AND user_id = :user_id
	`, map[string]any{"playlist_id": playlistID, "user_id": userID}); err != nil {
		return fmt.Errorf("deleting playlist tracks from track index: %w", err)
	}

	// Remove tracks that are no longer on any playlists
	if _, err := q.QueryContext(ctx, `
		DELETE FROM trackindex_tracks
		WHERE NOT EXISTS (
			SELECT 1
			FROM trackindex_playlist_tracks
			WHERE trackindex_playlist_tracks.track_key = trackindex_tracks.key
		)
	`); err != nil {
		return fmt.Errorf("deleting orphaned tracks from track index: %w", err)
	}

	return nil
}

func (t *TrackIndex) insertTrackIndexTrackOnPlaylist(ctx context.Context, q Querier, userID string, playlistID string, tracks []spotify.SimpleTrack) error {
	trackRows := make([]map[string]any, 0, len(tracks))
	playlistTrackRows := make([]map[string]any, 0, len(tracks))

	for _, track := range tracks {
		trackKey := t.trackIDFunc(track)

		trackJSON, err := json.Marshal(track)
		if err != nil {
			return fmt.Errorf("marshalling track: %w", err)
		}

		trackRows = append(trackRows, map[string]any{
			"key":          trackKey,
			"name":         track.Name,
			"simple_track": trackJSON,
			"user_id":      userID,
		})

		playlistTrackRows = append(playlistTrackRows, map[string]any{
			"playlist_id": playlistID,
			"track_key":   trackKey,
			"user_id":     userID,
		})
	}

	if _, err := q.NamedExecContext(ctx, `
		INSERT OR REPLACE INTO trackindex_tracks (
			key,
			name,
			simple_track,
			user_id,
			updated_at
		)
		VALUES (:key, :name, :simple_track, :user_id, datetime('now'))
	`, trackRows); err != nil {
		return fmt.Errorf("inserting tracks into track index: %w", err)
	}

	if _, err := q.NamedExecContext(ctx, `
		INSERT OR REPLACE INTO trackindex_playlist_tracks (
			playlist_id,
			track_key,
			user_id,
			updated_at
		)
		VALUES (:playlist_id, :track_key, :user_id, datetime('now'))
	`, playlistTrackRows); err != nil {
		return fmt.Errorf("inserting playlist-tracks into track index: %w", err)
	}

	return nil
}

func insertOrReplaceTrackIndexPlaylists(ctx context.Context, q Querier, userID string, playlist spotify.SimplePlaylist) error {
	playlistJSON, err := json.Marshal(playlist)
	if err != nil {
		return fmt.Errorf("marshalling playlist: %w", err)
	}

	values := map[string]any{
		"id":              playlist.ID.String(),
		"snapshot_id":     playlist.SnapshotID,
		"name":            playlist.Name,
		"simple_playlist": playlistJSON,
		"user_id":         userID,
	}
	if _, err := q.NamedExecContext(ctx, `
			INSERT OR REPLACE INTO trackindex_playlists (
				id,
				snapshot_id,
				name,
				simple_playlist,
				user_id,
				updated_at
			)
			VALUES (:id, :snapshot_id, :name, :simple_playlist, :user_id, datetime('now'))
		`, values); err != nil {
		return fmt.Errorf("inserting playlist into track index: %w", err)
	}
	return nil
}

func (t *TrackIndex) CountTracksByArtist(ctx context.Context, userID string, artistName string) (int, error) {
	db, release := t.db.RGet(ctx)
	defer release()

	var count int
	if err := db.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM trackindex_tracks
		CROSS JOIN json_each(simple_track, '$.artists') AS artist
		WHERE
			json_extract(artist.value, '$.name') = ?
			AND user_id = ?
	`, artistName, userID); err != nil {
		return 0, fmt.Errorf("counting tracks by artist (%s): %w", artistName, err)
	}

	return count, nil
}

func (t *TrackIndex) Summarize(ctx context.Context, userID string) (recommendations.IndexSummary, error) {
	// TODO: Get the track count to work 🤷
	db, release := t.db.RGet(ctx)
	defer release()

	var uniqueTrackCount int

	if err := db.GetContext(ctx, &uniqueTrackCount, `
		SELECT COUNT(*)
		FROM trackindex_tracks
		WHERE user_id = ?
	`, userID); err != nil {
		return recommendations.IndexSummary{}, fmt.Errorf("querying track index for track count: %w", err)
	}

	trackRows, err := db.QueryContext(ctx, `
		SELECT simple_playlist
		FROM trackindex_playlists
		WHERE user_id = ?
		`, userID)
	if err != nil {
		return recommendations.IndexSummary{}, fmt.Errorf("querying track index for playlists: %w", err)
	}
	defer trackRows.Close()

	var playlists []spotify.SimplePlaylist
	for trackRows.Next() {
		var b []byte
		if err := trackRows.Scan(&b); err != nil {
			return recommendations.IndexSummary{}, fmt.Errorf("scanning track index: %w", err)
		}

		var playlist spotify.SimplePlaylist
		if err := json.Unmarshal(b, &playlist); err != nil {
			return recommendations.IndexSummary{}, fmt.Errorf("unmarshalling playlist: %w", err)
		}
		playlists = append(playlists, playlist)
	}

	return recommendations.IndexSummary{
		UniqueTrackCount: uniqueTrackCount,
		PlaylistCount:    len(playlists),
		Playlists:        playlists,
	}, nil
}
