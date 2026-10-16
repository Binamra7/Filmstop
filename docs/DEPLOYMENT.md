# Deployment

## How it ships

- **Live:** https://filmstop.onrender.com
- **App:** Render free web service `filmstop`, Singapore region, built from the `Dockerfile`. `render.yaml` in the repo is the Blueprint.
- **Database:** Turso database `filmstop-binamra`, AWS Mumbai, the closest Turso region to Singapore.
- **Deploys:** Render deploys every push to `main` automatically. Work on a branch and open a pull request; CI must pass before merging.
- **Health check:** Render polls `/healthz`.

The free service sleeps after 15 minutes without traffic, and the next request waits about a minute while it starts.

## Settings

Set these in Render under **Environment**. Never commit them, and keep the Turso ones out of your local `.env`: there they would point local runs at production.

| Variable | Value |
| --- | --- |
| `TMDB_API_KEY` | TMDB v3 key or v4 read token |
| `TURSO_DATABASE_URL` | `libsql://filmstop-binamra.aws-ap-south-1.turso.io` |
| `TURSO_AUTH_TOKEN` | A Turso database token with read and write access |

Render sets `PORT` itself, and the app reads it.

## Setting it up from scratch

1. **Turso:** create a database, then **Create Token** (no expiry, read and write). Copy the `libsql://` URL and the token. Turso shows the token only once.
2. **Render:** **New → Blueprint**, connect GitHub (access to this repo is enough), pick the repo and branch `main`. Fill in the three variables above and click **Deploy Blueprint**.
3. **Check:** wait for "Your service is live", then open `/healthz` and `/`.

If the first deploy fails with `401 Unauthorized ... empty JWT token`, the auth token wasn't saved. Add it under Environment and choose **Save, rebuild, and deploy**.

## Restarting

The server builds the list in memory at startup and rebuilds it only after votes cast through the site. After changing the database directly, restart the service: **Manual Deploy → Restart service** in Render.

## Seeding or resetting production

Production holds about 5,000 made-up ballots, added on purpose for the demo. Seeded voters have ids starting with `seed`, and the five personas have fixed ids.

To reseed, from a machine with Go:

```bash
set -a; . ~/.config/filmstop/turso.env; set +a      # TURSO_DATABASE_URL and TURSO_AUTH_TOKEN
TMDB_API_KEY=... go run ./cmd/seed -remote -voters 400
```

Then restart the service. Running the seed again replaces the earlier seeded ballots and resets the personas, and it never touches real visitors' ballots. Over the network it writes one ballot per transaction, so keep `-voters` in the hundreds. For thousands, seed a local file and copy the rows across in batches instead.

To remove the seeded data, run this in Turso's SQL console (**Edit Data** on the database), then restart:

```sql
DELETE FROM ballots      WHERE substr(voter, 1, 4) = 'seed';
DELETE FROM ballot_times WHERE substr(voter, 1, 4) = 'seed';
```

The personas' ballots stay. Remove those by their voter ids, which the seed prints with each cookie.

## Troubleshooting

| Symptom | Cause | Fix |
| --- | --- | --- |
| Deploy fails, logs say `empty JWT token` | `TURSO_AUTH_TOKEN` missing | Add it in Environment and redeploy |
| A 500 after idle time, logs say `.../v2/pipeline": EOF` | A pooled connection to Turso died | Retried automatically since `store/remote.go`; if it reappears, look at that transport |
| New ballots in the database don't show | The list is built in memory | Restart the service |
| Search returns nothing | `TMDB_API_KEY` missing or wrong | Logs show `TMDB_API_KEY is not set` or the TMDB error |
| First visit takes about a minute | Free service was asleep | Expected on the free plan |
