# Deploying Recommendli to Railway

The repository is ready to deploy as one Railway service. Railway builds the
multi-stage `Dockerfile`, injects `PORT`, and checks `/status` according to
`railway.toml`.

## 1. Create the service

1. In Railway, create a project and choose **Deploy from GitHub repo**.
2. Select this repository and the branch to deploy.
3. Leave the root directory at the repository root. Railway will use
   `railway.toml` and the `Dockerfile` automatically.
4. Generate a Railway public domain under **Settings > Networking** (or add the
   final custom domain now).

Do not increase the service above one replica. Recommendli uses one SQLite file,
and a Railway volume can only be mounted by one service replica.

## 2. Add persistent storage

Create a Railway volume on the service and mount it at:

```text
/data
```

Set `SQLITE_DB_PATH=/data/recommendli.sqlite` as shown below. Without the volume,
the database is lost whenever Railway replaces the container.

## 3. Configure variables

Add these service variables under **Variables**:

| Variable | Value |
| --- | --- |
| `SPOTIFY_ID` | Spotify application client ID |
| `SPOTIFY_SECRET` | Spotify application client secret |
| `SPOTIFY_REDIRECT_HOST` | Public origin, e.g. `https://recommendli.remback.se` or the generated Railway origin; no trailing slash |
| `SQLITE_DB_PATH` | `/data/recommendli.sqlite` |
| `LOG_LEVEL` | `info` |

Do not set `PORT`; Railway provides it. The application listens on Railway's
`PORT` automatically. `ADDR` is only an optional explicit override.

## 4. Update Spotify OAuth

In the Spotify Developer Dashboard, add this exact redirect URI for the chosen
public origin:

```text
https://<public-host>/recommendations/v1/spotify/auth/callback
```

`SPOTIFY_REDIRECT_HOST` must use the same scheme and host. If both the Railway
domain and a custom domain should work, Spotify must list each callback URI and
the deployed variable must correspond to the domain users visit.

## 5. Deploy and verify

Trigger a deployment and verify:

```bash
curl -fsS https://<public-host>/status
```

The response should be `OK`. Then open the public URL and complete a Spotify
login and recommendation flow. On first start, the application creates the
SQLite database and applies all migrations automatically.

Useful checks in Railway:

- Build logs show both the frontend (`npm run build`) and Go build succeeding.
- Deploy logs include `Starting server` with Railway's assigned port.
- The volume is mounted at `/data`.
- A redeploy preserves `/data/recommendli.sqlite`.

## Existing deployment and cutover

The SQLite data is primarily cached Spotify data and the track index, so a new
empty database is safe; it will be rebuilt as users use the application. If the
existing database must be retained, stop writes, copy the SQLite file into the
Railway volume as `/data/recommendli.sqlite`, and only then start the Railway
service. Do not copy a live SQLite file without using SQLite's backup mechanism.

For a custom domain such as `recommendli.remback.se`:

1. Validate the deployment on its generated Railway domain.
2. Add the custom domain in Railway and apply the DNS record Railway provides.
3. Change `SPOTIFY_REDIRECT_HOST` to the custom HTTPS origin and confirm its
   callback URI is registered with Spotify.
4. Cut DNS over only after `/status` and Spotify login succeed.
5. Keep the old service available until the new deployment has been verified;
   then stop it to avoid two independent SQLite databases receiving traffic.

## Rollback

Point DNS back to the previous deployment (if a custom domain was cut over) and
restart that service. Railway's volume should be retained rather than deleted,
so its database remains available for investigation or a later retry.
