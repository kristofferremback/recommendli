package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kristofferostlund/recommendli/pkg/singleflight"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
)

type Locker struct {
	db *DB
}

func NewLocker(db *DB) *Locker {
	return &Locker{db: db}
}

func (l *Locker) Lock(ctx context.Context, key string, ttl time.Duration) (string, error) {
	db, release := l.db.Get(ctx)
	defer release()

	token := fmt.Sprintf("%s__%s", key, uuid.New().String())

	ctx = slogutil.WithAttrs(ctx, slog.String("key", key), slog.String("token", token), slog.Duration("ttl", ttl))
	slog.DebugContext(ctx, "locking key")

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	ttlSeconds := fmt.Sprintf("%d seconds", int(ttl.Seconds()))

	row := tx.QueryRowContext(ctx, `
		SELECT token, expires_at < datetime('now') AS is_expired
		FROM singleflight_locks
		WHERE key = ? AND expires_at IS NOT NULL;
	`, key)
	var existingToken string
	var isExpired bool
	if err := row.Scan(&existingToken, &isExpired); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("querying existing lock: %w", err)
	}
	if existingToken != "" && !isExpired {
		slog.DebugContext(ctx, "key is already locked", slog.String("existing_token", existingToken))
		return "", singleflight.ErrLocked
	}

	if existingToken != "" {
		slog.DebugContext(ctx, "existing lock is expired, unlocking it", slog.String("existing_token", existingToken))
		if err := l.releaseLock(ctx, tx, existingToken, token); err != nil {
			return "", fmt.Errorf("releasing existing lock: %w", err)
		}
	}

	if _, err := tx.NamedExecContext(ctx, `
		INSERT INTO singleflight_locks (key, token, expires_at)
		VALUES (:key, :token, datetime('now', :ttl))
	`, map[string]any{"key": key, "token": token, "ttl": ttlSeconds}); err != nil {
		if IsUniqueConstraintViolation(err) {
			return "", singleflight.ErrLocked
		}
		return "", fmt.Errorf("locking key %s: %w", key, err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("committing tx: %w", err)
	}

	return token, nil
}

func (l *Locker) Refresh(ctx context.Context, token string, ttl time.Duration) error {
	db, release := l.db.Get(ctx)
	defer release()

	ctx = slogutil.WithAttrs(ctx, slog.String("token", token), slog.Duration("ttl", ttl))
	slog.DebugContext(ctx, "refreshing lock")

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	ttlSeconds := fmt.Sprintf("%d seconds", int(ttl.Seconds()))

	result, err := tx.NamedExecContext(ctx, `
		UPDATE singleflight_locks
		SET expires_at = datetime('now', :ttl)
		WHERE token = :token;
	`, map[string]any{"token": token, "ttl": ttlSeconds})
	if err != nil {
		return fmt.Errorf("refreshing lock: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		slog.DebugContext(ctx, "lock not found, is it already unlocked?")
		return singleflight.ErrNoSuchLock
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	slog.DebugContext(ctx, "lock refreshed")

	return nil
}

func (l *Locker) Unlock(ctx context.Context, token string) error {
	db, release := l.db.Get(ctx)
	defer release()

	ctx = slogutil.WithAttrs(ctx, slog.String("token", token))
	slog.DebugContext(ctx, "unlocking lock")

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	if err := l.releaseLock(ctx, tx, token, token); err != nil {
		return fmt.Errorf("releasing lock: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	slog.DebugContext(ctx, "lock unlocked")

	return nil
}

func (*Locker) releaseLock(ctx context.Context, q Querier, token, releasedBy string) error {
	result, err := q.NamedExecContext(ctx, `
		UPDATE singleflight_locks
		SET expires_at = NULL,
			released_at = datetime('now'),
			released_by = :release_by
		WHERE token = :token
	`, map[string]any{"token": token, "release_by": releasedBy})
	if err != nil {
		return fmt.Errorf("unlocking key: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting affected rows: %w", err)
	}
	if affectedRows == 0 {
		slog.DebugContext(ctx, "lock not found, is it already unlocked?", "token_to_unlock", token)
		return singleflight.ErrNoSuchLock
	}

	return nil
}

func IsUniqueConstraintViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
