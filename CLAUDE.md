# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Recommendli is a Spotify recommendation service that analyzes playlists and listening habits to generate personalized music recommendations. It filters and ranks tracks from Discovery Weekly and Release Radar playlists, excluding songs already in the user's library.

**Tech Stack**: Go backend with vanilla JavaScript frontend, SQLite for persistence, Spotify API integration.

## Development Commands

### Running the Application

```bash
# Run locally (requires .env file with Spotify credentials)
make dev

# Build executable
make build

# Run with Docker Compose
docker-compose up --build
```

### Environment Variables

Required environment variables (configure in `.env` for local development):

- `SPOTIFY_ID` - Spotify client ID
- `SPOTIFY_SECRET` - Spotify client secret
- `SPOTIFY_REDIRECT_HOST` - OAuth redirect host (default: `http://0.0.0.0:9999`)
- `ADDR` - Server address (default: `0.0.0.0:9999`)
- `LOG_LEVEL` - Logging level (default: `info`)
- `FILE_CACHE_BASE_DIR` - Cache directory (default: `/tmp/recommendli`)
- `SQLITE_DB_PATH` - SQLite database path (default: `/tmp/recommendli.sqlite`)

### Build Details

- CGO is **enabled** for SQLite support (`CGO_ENABLED=1`)
- Production Dockerfile disables CGO (`CGO_ENABLED=0`) and uses distroless base image
- Built executable location: `./build/main`

## Architecture

### Core Concepts

**Track Index**: SQLite-backed index that maps tracks to playlists. The system syncs playlist snapshots, detects changes (added/changed/removed playlists), and maintains a normalized view of user's library. This enables fast lookups to check if a track already exists in any user playlist.

**Singleflight Locking**: The `pkg/singleflight` package provides distributed locking for expensive operations like syncing the track index. When multiple requests try to sync simultaneously, only one proceeds while others wait and poll. Locks auto-refresh using a ticker to prevent expiration during long operations.

**Spotify Adaptor**: Wraps the Spotify API client with caching (via KeyValueStore) and pagination helpers. Split across multiple files:
- `spotifyadaptor.go` - Core adaptor, tracks, current user
- `spotifyadaptor_auth.go` - OAuth flow handlers
- `spotifyadaptor_albums.go` - Album operations
- `spotifyadaptor_playlists.go` - Playlist operations

**Service Layer**: Business logic lives in `internal/recommendations/recommendations.go`. The service uses dependency injection via factories to compose SpotifyProvider, KeyValueStore, UserPreferences, and TrackIndex.

### Key Components

#### `internal/recommendations/`
- `recommendations.go` - Core service implementing recommendation logic
- `recommendations_lib.go` - Helper functions for track scoring and filtering
- `http_handler.go` - HTTP handlers using chi router
- `trackindex.go` - Interface for track-to-playlist index
- `keyvaluestore.go` - Generic key-value persistence interface
- `spotifyadaptor*.go` - Spotify API client wrapper with caching

#### `internal/sqlite/`
- `db.go` - Wrapper around sqlx.DB with RWMutex for safe concurrent access
- `trackindex.go` - SQLite implementation of TrackIndex interface
- `keyvaluestore.go` - SQLite-backed key-value store
- `locker.go` - SQLite-backed distributed lock for singleflight

#### `pkg/`
Reusable utilities:
- `singleflight/` - Distributed locking with auto-refresh
- `slogutil/` - Structured logging helpers (slog wrappers)
- `migrations/` - Database migration runner (uses golang-migrate)
- `paginator/` - Pagination helpers
- `srv/` - HTTP utilities (JSON responses, 404 redirector)
- `ctxhelper/` - Context utilities
- `sortby/` - Sorting helpers
- `spotifyutil/` - Spotify-specific utilities

### Database

Migrations are in `./migrations/` and run automatically on startup via `migrations.Up()` in `main.go`.

Key tables:
- `keyvaluestore` - Generic key-value storage
- `trackindex_playlists` - Stores playlist snapshots per user
- `trackindex_playlist_tracks` - Normalized track-to-playlist mapping
- `singleflight_locker` - Distributed lock state

### Request Flow

1. User authenticates via `/recommendations/v1/spotify/auth/callback` (OAuth)
2. Auth middleware injects Spotify client into request context
3. HTTP handler creates service with SpotifyAdaptor factory
4. Service operations:
   - Sync track index (using singleflight to prevent duplicate work)
   - Query Spotify for Discovery/Release Radar playlists
   - Score tracks based on user preferences (artist relevance, album size, weighted words)
   - Filter out tracks already in user's library (via track index lookup)
   - Generate playlist with ranked recommendations

### Caching Strategy

Two levels of caching:
1. **Spotify API responses** - Cached in KeyValueStore (SQLite) to reduce API calls
2. **Track index** - Pre-computed mapping of tracks to playlists for fast "already in library" checks

### Frontend

Static files in `./static/` served by Go's `http.FileServer` with custom 404 redirector that routes to `index.html` for SPA routing.

JavaScript uses vanilla ES6 modules (no build step) with:
- `components/` - UI components
- `store/` - Client-side state management
- `hooks/` - Custom hooks
- `recommendli/` - Domain logic

## Deployment

The application is deployed to **recommendli.remback.se** using systemd and nginx.

### Deployment Setup

**Application location**: `/usr/share/recommendli`

**Components**:
- `deploy/recommendli.service` - systemd unit file
- `deploy/nginx.conf` - nginx reverse proxy configuration
- `deploy/start.sh` - helper script that builds if needed and starts the application

### Deployment Steps

1. **Build the application**:
   ```bash
   make build
   ```

2. **Copy files to server**:
   ```bash
   # Copy entire repo to /usr/share/recommendli
   # Ensure .env file exists at /usr/share/recommendli/.env
   ```

3. **Install systemd service**:
   ```bash
   sudo cp deploy/recommendli.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable recommendli
   sudo systemctl start recommendli
   ```

4. **Configure nginx**:
   ```bash
   # Copy nginx config to /etc/nginx/sites-available/
   # Enable site and setup SSL with certbot
   sudo ln -s /etc/nginx/sites-available/recommendli.conf /etc/nginx/sites-enabled/
   sudo certbot --nginx -d recommendli.remback.se
   sudo systemctl reload nginx
   ```

### Quick Deploy (Manual)

To deploy changes manually on the server:

```bash
cd /usr/share/recommendli
git pull
make build
sudo systemctl restart recommendli
```

### Service Management

```bash
# Check status
sudo systemctl status recommendli

# View logs
sudo journalctl -u recommendli -f

# Restart service
sudo systemctl restart recommendli
```

### Notes

- The service runs with a memory limit of 1500M
- Auto-restarts on failure with 5s delay
- Application serves on port 9999 (proxied by nginx)
- SSL certificates managed by certbot

## Testing

Currently no test files exist in the repository. When adding tests:
- Place test files adjacent to implementation (`*_test.go`)
- Use standard Go testing conventions
- Consider table-driven tests for different user preferences/scoring scenarios
