# Recommendli

Recommendli combines and ranks tracks from Spotify's Release Radar and Discover
Weekly playlists, filtering out music that is already in the user's library.

## Local development

1. Copy `.env.example` to `.env` and add Spotify credentials.
2. Register this Spotify callback URI:
   `http://127.0.0.1:9999/recommendations/v1/spotify/auth/callback`.
3. Run the backend and frontend:

```bash
make frontend-install
make dev-with-frontend
```

Open <http://127.0.0.1:5173>.

To build and run the production container locally:

```bash
docker compose up --build
```

## Configuration

| Variable | Purpose | Default |
| --- | --- | --- |
| `SPOTIFY_ID` | Spotify client ID | required |
| `SPOTIFY_SECRET` | Spotify client secret | required |
| `SPOTIFY_REDIRECT_HOST` | Public origin used for OAuth callbacks | `http://127.0.0.1:9999` |
| `PORT` | HTTP port (automatically supplied by Railway) | `9999` |
| `ADDR` | Optional full listen-address override | `0.0.0.0:$PORT` |
| `SQLITE_DB_PATH` | SQLite database file | `/tmp/recommendli.sqlite` |
| `LOG_LEVEL` | Application log level | `info` |

## Railway deployment

Railway deployment configuration, persistent-volume requirements, Spotify OAuth
setup, cutover, and rollback instructions are in
[`deploy/railway.md`](deploy/railway.md).
