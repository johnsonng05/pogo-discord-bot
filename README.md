# Pokémon GO Event & Raid Discord Bot

A lightweight Discord bot written in Go using `[discordgo](https://github.com/bwmarrin/discordgo)` for tracking Pokémon GO events, raids, and Pokémon information.

## Features

- `/pogo-current-events` — View current Pokémon GO events.
- `/pogo-upcoming-events` — View upcoming Pokémon GO events.
- `/pogo-raids` — View currently active raid bosses.
- `/pokemon-lookup` — Look up Pokémon stats, moves, and type effectiveness.
- Daily automated event announcements at **8:00 AM**.
- Slash command support through Discord's Application Commands API.
- Community-maintained Pokémon GO data sources.
- Concurrent background processing using Go goroutines.

## Data Sources

The bot currently uses community-maintained JSON datasets:

| Feature            | Data Source                                                                                          |
| ------------------ | ---------------------------------------------------------------------------------------------------- |
| Event schedules    | [Leek Duck via ScrapedDuck](https://raw.githubusercontent.com/bigfoott/ScrapedDuck/data/events.json) |
| Raid bosses        | [Leek Duck via ScrapedDuck](https://raw.githubusercontent.com/bigfoott/ScrapedDuck/data/raids.json)  |
| Pokémon stats      | [pogoapi.net](https://pogoapi.net/api/v1/pokemon_stats.json)                                         |
| Pokémon moves      | [pogoapi.net](https://pogoapi.net/api/v1/current_pokemon_moves.json)                                 |
| Pokémon types      | [pogoapi.net](https://pogoapi.net/api/v1/pokemon_types.json)                                         |
| Type effectiveness | [pogoapi.net](https://pogoapi.net/api/v1/type_effectiveness.json)                                    |
| Sprites            | [pokemon-go-api](https://pokemon-go-api.github.io/pokemon-go-api/api/pokedex.json)                   |

These sources are community-maintained and may change or become unavailable over time.

## Requirements

- Go 1.21+
- A Discord application and bot
- A Discord server where the bot can be installed

## Configuration

Set the following environment variables before running the bot:

| Variable                  | Description                             |
| ------------------------- | --------------------------------------- |
| `DISCORD_TOKEN`           | Discord bot authentication token        |
| `ANNOUNCEMENT_CHANNEL_ID` | Channel ID used for daily announcements |

Example:

```bash
export DISCORD_TOKEN="your-bot-token"
export ANNOUNCEMENT_CHANNEL_ID="your-channel-id"

go run main.go
```

## Discord Setup

Create a bot application through the [Discord Developer Portal](https://discord.com/developers/applications).

The bot should be invited with the following OAuth2 scopes:

- `bot`
- `applications.commands`

Required permissions:

- View Channels
- Send Messages
- Embed Links

The bot uses Discord slash commands, so `applications.commands` is required when generating the authorization URL.

## Project Structure

```text
pogo-bot/
├── internal/
│   ├── api/
│   │   ├── client.go
│   │   └── client_test.go
│   ├── bot/
│   │   └── bot.go
│   ├── commands/
│   │   ├── commands.go
│   │   ├── events.go
│   │   ├── lookup.go
│   │   └── raids.go
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   └── models.go
│   └── scheduler/
│       └── scheduler.go
├── go.mod
├── go.sum
└── main.go
```

## How It Works

The bot has two main processes:

### Slash Commands

When a user runs a Pokémon GO command, the bot acknowledges the interaction and retrieves the required community data before responding with the results.

### Daily Announcements

A background goroutine checks the server's local time every minute. At **8:00 AM**, it retrieves the current event schedule and posts a daily summary to the configured announcement channel.

## Roadmap

- [ ] Add redis cache to reduce repeated API requests.
- [ ] Add HTTP timeouts and improved error handling.
- [ ] Add structured logging.
- [ ] Complete `/pogo-current-events`.
- [ ] Complete `/pogo-upcoming-events`.
- [ ] Complete `/pogo-raids`.
- [ ] Complete `/pokemon-lookup`.
- [ ] Add Pokémon type and moveset recommendations.
- [ ] Improve Discord embeds and command responses.
- [ ] Add tests for API parsing and command handlers.
- [ ] Add Dockerfile and containerize the application.
- [ ] Add Docker Compose configuration for local development.
- [ ] Set up AWS EC2 instance for production deployment.
- [ ] Deploy the Dockerized application to AWS EC2.
- [ ] Document local development and production deployment.
- [ ] Configure environment variables and secrets for deployment using AWS Systems Manager Parameter Store.
- [ ] Configure Redis for production.
- [ ] Set up application logging and monitoring on EC2.

## License

This project is currently under development.
