package singleflight

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kristofferostlund/recommendli/pkg/slogutil"
)

var (
	ErrLocked     = fmt.Errorf("locked")
	ErrNoSuchLock = fmt.Errorf("no such lock")
)

type Locker interface {
	// Lock locks the key and returns a token and an empty error. If the key is already
	// locked, it returns ErrLocked.
	Lock(ctx context.Context, key string, ttl time.Duration) (token string, err error)
	// Refresh refreshes the lock for the corresponding token.
	// If the lock is already unlocked (or expired and re-claimed), it returns ErrNoSuchLock.
	Refresh(ctx context.Context, token string, ttl time.Duration) error
	// Unlock unlocks the key with the corresponding token. If the key is already unlocked, it returns ErrNoSuchLock.
	// ErrNoSuchLock can be safely ignored. Treat this the same as you do tx.Rollback(...).
	Unlock(ctx context.Context, token string) error
}

type DoFunc[V any] func(ctx context.Context, key string, fn func(ctx context.Context) (V, error)) (V, error)

// Prepare returns a DoFunc that locks the key before calling the provided function.
// The key is refreshed every 75% of the ttl duration.
func Prepare[V any](locker Locker, ttl time.Duration) DoFunc[V] {
	refreshDuration := time.Duration(float64(ttl) * 0.75) // Must cast ttl to float64 to not get integer division
	pollDuration := time.Duration(float64(ttl) * 0.25)

	return func(ctx context.Context, key string, fn func(ctx context.Context) (V, error)) (V, error) {
		var empty V

	outer:
		for {
			token, err := locker.Lock(ctx, key, ttl)
			if err != nil {
				if errors.Is(err, ErrLocked) {
					select {
					case <-ctx.Done():
						return empty, ctx.Err()
					case <-time.After(pollDuration):
						slog.DebugContext(ctx, "lock already claimed, polling", slog.Duration("poll_duration", pollDuration))
						continue outer
					}
				}
				return empty, fmt.Errorf("failed to lock %s: %w", key, err)
			}
			defer locker.Unlock(ctx, token)

			ticker := time.NewTicker(refreshDuration)
			defer ticker.Stop()
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if err := locker.Refresh(ctx, token, ttl); err != nil {
							slog.WarnContext(ctx, "ignoring failure to refresh key", slog.String("key", key), slogutil.Error(err))
						}
					}
				}
			}()

			return fn(ctx)
		}
	}
}
