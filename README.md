# FamilyBot

A private Go-based Telegram bot for family chat with AI assistance, weather updates, translations, and automated news digests.

## Features

- **AI Chat**: ChatGPT integration with conversation history and grammar correction
- **Weather**: Multi-location forecasts with timezone support
- **Translation**: English ↔ Russian via DeepL
- **Morning Digest**: Automated 7 AM updates with weather, currency rates, and RSS news
- **Voice Transcription**: Convert Telegram voice messages to text
- **Access Control**: Admin-managed authorization with invite links

## Quick Start

```bash
# Local build
cd src && go build -mod vendor -o main .

# Docker deployment
docker-compose up -d
```

## Configuration

Required environment variables:
```
TG_TOKEN              # Telegram bot token
TG_GROUP              # Main group chat ID
TG_ADMINUSERIDS       # Comma-separated admin user IDs
WEATHERAPI_KEY        # Weather API key
CURRENCYAPI_KEY       # Currency API key
OPENAIAPI_KEY         # OpenAI API key
DEEPLAPI_KEY          # DeepL API key
MINIFLUXAPI_KEY       # Miniflux API key
MINIFLUXAPI_BASEURL   # Miniflux instance URL
REDIS_ADDR            # Redis connection string
```

## Commands

**User Commands:**
- `/gpt <message>` - Chat with AI
- `/weather` - Get weather forecast
- `/fix <text>` - Fix English grammar
- `/en2ru`, `/ru2en` - Translate text
- `/restart` - Reset ChatGPT context
- `/list` - Show all commands

**Admin Commands:**
- `/add <user_id>`, `/remove <user_id>` - Manage authorized users
- `/users` - List authorized users
- `/invite` - Generate invite link

## Tech Stack

- Go 1.24.1
- Redis (caching)
- Docker
- OpenAI, DeepL, Weather & Currency APIs, Miniflux
- Prometheus metrics
