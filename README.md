# Pokémon GO Event & Raid Discord Bot

A lightweight Discord bot written in Go using [discordgo](https://github.com/bwmarrin/discordgo) for tracking Pokémon GO events, raids, and Pokémon information.

## Why

Recently, I was encouraged by a friend to play Pokemon Go over the summer. I'd never played before, but I've been hooked ever since I started. However, I noticed keeping up with current and upcoming events, raids, etc. was bothersome. So, why not a discord bot using the new language I'm learning called Go? Pokemon Go Bot using Go HAHA.

## Features

### Slash commands

| Command                      | Description                                                               |
| ---------------------------- | ------------------------------------------------------------------------- |
| `/pogo-current-events`       | Live Pokémon GO events                                                    |
| `/pogo-upcoming-events`      | Upcoming events (capped at 10)                                            |
| `/pogo-raids`                | Current raid bosses by tier                                               |
| `/pokemon-lookup`            | Pokémon stats, types, moves, and GO icon                                  |
| `/pogo-set-announce-channel` | Server admins set the daily announcement channel (requires Manage Server) |

### Background jobs

- **Daily announcements** — a goroutine checks every minute and posts a live-events summary at **8:00 AM PT** to channels configured in Redis (or `ANNOUNCEMENT_CHANNEL_ID` as a fallback).
- **Redis caching** — community API JSON is cached with a 24-hour TTL to reduce upstream requests.

## Data Sources

Community-maintained JSON datasets (Niantic has no public API):

| Feature            | Source                                                                                               |
| ------------------ | ---------------------------------------------------------------------------------------------------- |
| Event schedules    | [ScrapedDuck](https://raw.githubusercontent.com/bigfoott/ScrapedDuck/data/events.json) `events.json` |
| Raid bosses        | [ScrapedDuck](https://raw.githubusercontent.com/bigfoott/ScrapedDuck/data/raids.json) `raids.json`   |
| Pokémon stats      | [pogoapi.net](https://pogoapi.net/api/v1/pokemon_stats.json)                                         |
| Pokémon moves      | [pogoapi.net](https://pogoapi.net/api/v1/current_pokemon_moves.json)                                 |
| Pokémon types      | [pogoapi.net](https://pogoapi.net/api/v1/pokemon_types.json)                                         |
| Type effectiveness | [pogoapi.net](https://pogoapi.net/api/v1/type_effectiveness.json)                                    |
| GO icons           | [pokemon-go-api](https://pokemon-go-api.github.io/pokemon-go-api/api/pokedex.json)                   |

These sources may change or become unavailable over time.

## Requirements

- Go 1.24+
- Docker & Docker Compose (recommended for local dev with Redis)
- A [Discord application](https://discord.com/developers/applications) and bot token
- Redis (required — announcement channels + API cache)

### Redis by environment

| Environment    | Redis                                         | `REDIS_URL` source                                                              |
| -------------- | --------------------------------------------- | ------------------------------------------------------------------------------- |
| **Local**      | Self-hosted (`docker-compose` includes Redis) | Compose sets `redis://redis:6379/0`, or `redis://127.0.0.1:6379/0` for `go run` |
| **Production** | [Upstash](https://upstash.com/) Redis Cloud   | SSM `/prod/discord/redis_url` (`rediss://...` from Upstash console)             |

No code changes between environments — only the connection URL differs.

## Configuration

Copy `.env.example` to `.env` and fill in your values:

| Variable                  | Required    | Description                                                                       |
| ------------------------- | ----------- | --------------------------------------------------------------------------------- |
| `APP_ENV`                 | Yes (local) | Set to `development` for local env vars; leave **unset** in production (SSM path) |
| `DISCORD_TOKEN`           | Yes (local) | Bot token from the Discord Developer Portal                                       |
| `REDIS_URL`               | Yes (local) | e.g. `redis://127.0.0.1:6379/0`                                                   |
| `ANNOUNCEMENT_CHANNEL_ID` | No          | Fallback channel for daily posts if no guild has run `/pogo-set-announce-channel` |

Local: `APP_ENV=development` loads secrets from `.env`. Production: leave `APP_ENV` unset so `config.Load()` reads SSM Parameter Store.

## Local development

### Option A — Docker Compose (recommended)

Runs the bot and Redis together:

```bash
cp .env.example .env
# Edit .env — set DISCORD_TOKEN at minimum

docker compose up --build -d
docker compose logs -f bot
docker compose down
```

Compose sets `REDIS_URL=redis://redis:6379/0` and `TZ=America/Los_Angeles` automatically.

Verify caching after a slash command:

```bash
docker exec -it pogo-discord-bot-redis redis-cli KEYS 'pogo:*'
docker exec -it pogo-discord-bot-redis redis-cli TTL pogo:events
```

### Option B — Run on the host

Start Redis, then:

```bash
export DISCORD_TOKEN="your-bot-token"
export REDIS_URL="redis://127.0.0.1:6379/0"
# optional: export ANNOUNCEMENT_CHANNEL_ID="your-channel-id"

go run .
```

### Tests

Live API tests (no Redis):

```bash
go test ./internal/api/ -count=1
```

Redis tests (requires `REDIS_URL`):

```bash
export REDIS_URL=redis://127.0.0.1:6379/0
go test ./internal/cache/ ./internal/api/ -run 'TTL|Expire|CachesWith' -v
```

## Discord setup

1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications).
2. Under **OAuth2 → URL Generator**, select scopes: `bot`, `applications.commands`.
3. Bot permissions: **View Channels**, **Send Messages**, **Embed Links**.
4. Invite the bot to your server with the generated URL.

## How it works

### Slash command pipeline

User runs a command → bot defers the interaction → fetches data (Redis cache first, then HTTP) → responds with embeds.

### Daily announcement scheduler

A background goroutine wakes every minute. At 8:00 AM PT it loads announcement channel IDs from Redis, fetches live events once, and posts the daily brief to each channel.

### Redis keys

| Key                       | Type         | Purpose                 |
| ------------------------- | ------------ | ----------------------- |
| `announcement:channels`   | Hash         | `guildID` → `channelID` |
| `pogo:events`             | String + TTL | Cached events JSON      |
| `pogo:raids`              | String + TTL | Cached raids JSON       |
| `pogo:pokemon_stats`      | String + TTL | Cached stats JSON       |
| `pogo:pokemon_moves`      | String + TTL | Cached moves JSON       |
| `pogo:pokemon_types`      | String + TTL | Cached types JSON       |
| `pogo:type_effectiveness` | String + TTL | Cached matchup JSON     |
| `pogo:go_images`          | String + TTL | Slim name→form→GO icon URL map (not full pokedex) |

Default TTL: **24 hours** (events, raids, stats, etc.).  
`pogo:go_images` TTL: **30 days** (icon URLs change rarely).

## Project structure

```text
pogo-bot/
├── internal/
│   ├── api/          # HTTP client + Redis cache-aside
│   ├── bot/          # Discord session wiring
│   ├── cache/        # Redis (announcements + API cache)
│   ├── commands/     # Slash command handlers
│   ├── config/       # Env + AWS SSM loading
│   ├── models/       # JSON struct types
│   └── scheduler/    # Daily 8 AM announcements
├── docker-compose.yml
├── Dockerfile
├── main.go
├── go.sum
└── go.mod
```

## Production deployment

EC2 + ECR + Parameter Store + **Upstash Redis**. Local `docker-compose.yml` is self-hosted Redis for development only.

Replace placeholders with your values (example region: `us-west-1`):

```text
YOUR_ACCOUNT_ID
YOUR_REGION          # e.g. us-west-1
ECR_URI=YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/pogo-bot
```

### Redeploy (after code changes)

#### 1. Laptop — build and push

```bash
cd /path/to/Pokemon

aws ecr get-login-password --region YOUR_REGION \
  | docker login --username AWS --password-stdin \
  YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com

docker build --platform linux/arm64 -t pogo-discord-bot .

docker tag pogo-discord-bot:latest \
  YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/pogo-discord-bot:manual-test

docker push \
  YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/pogo-discord-bot:manual-test
```

#### 2. EC2 — pull and restart

```bash
aws ecr get-login-password --region YOUR_REGION \
  | docker login --username AWS --password-stdin \
  YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com

docker pull \
  YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/pogo-discord-bot:manual-test

docker stop pogo-bot
docker rm pogo-bot

docker run -d \
  --name pogo-discord-bot \
  --restart unless-stopped \
  -e AWS_REGION=YOUR_REGION \
  YOUR_ACCOUNT_ID.dkr.ecr.YOUR_REGION.amazonaws.com/pogo-discord-bot:manual-test

docker logs -f pogo-discord-bot
```

Do **not** pass `APP_ENV`, `DISCORD_TOKEN`, or `REDIS_URL` on EC2. Secrets come from SSM; `AWS_REGION` is required so the SDK can call Parameter Store.

#### 3. Verify

- Logs show the bot connected to Discord
- Slash commands work in your server
- Upstash shows `pogo:*` keys after a command

## Roadmap

- [x] Slash commands: events, raids, lookup
- [x] Daily announcement scheduler (8 AM PT)
- [x] Per-guild announcement channels via Redis
- [x] Redis API caching (24h TTL)
- [x] Dockerfile and Docker Compose
- [x] Deploy to AWS EC2
- [ ] Type effectiveness in `/pokemon-lookup`
- [ ] Optional `form` parameter on lookup (Origin, Altered, etc.)
- [ ] Raid bosses in daily announcement
- [ ] GitHub Actions deploy pipeline

## License

This project is currently under development.
