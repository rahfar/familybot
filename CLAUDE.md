# FamilyBot - Telegram Bot Project

## Overview
A Go-based Telegram bot designed for family chat interactions with multiple API integrations for weather, currency, AI assistance, and news aggregation.

## Architecture

### Tech Stack
- **Language**: Go 1.24.1
- **Deployment**: Docker with compose.yaml
- **Database**: Redis for caching
- **Dependencies**: Vendor-based (use `-mod vendor`)

### Core Components

#### Main Application (`src/main.go`)
- Entry point with configuration parsing using `go-flags`
- Environment variable configuration for all API keys
- Initializes all API clients and bot instance
- Supports debug mode via `DEBUG` env var

#### Bot Core (`src/bot/`)
- **bot.go**: Main bot logic, message routing, morning digest scheduler
- **handler.go**: Message handlers for all commands
- **command.go**: Command definitions and registry
- **audio.go**: Voice message transcription support

#### API Clients (`src/apiclient/`)
- **openai.go**: ChatGPT integration for AI responses
- **weather.go**: Weather API for location-based forecasts
- **currency.go**: Exchange rates and currency conversion
- **deepl.go**: Translation services
- **miniflux.go**: RSS news aggregation
- **anthropic.go**: Alternative AI provider

#### Infrastructure
- **metrics/**: Prometheus metrics collection
- **redisclient/**: Redis caching layer

## Key Features

### Bot Commands
- `/gpt` - ChatGPT integration with conversation history
- `/weather` - Multi-location weather forecasts
- `/new` - Reset ChatGPT context
- `/fix` - English grammar correction
- `/en2ru` / `/ru2en` - Translation between English and Russian
- `/list` - Show available commands
- **Admin-only commands**:
  - `/add <user_id>` - Add user to authorized list
  - `/remove <user_id>` - Remove user from authorized list
  - `/users` - List all authorized users
  - `/invite` - Generate invite link for new users
- Hidden admin commands: `/ping`, `/whoami`, `/revision`, `/mourning`

### Automated Features
- **Morning Digest** (7 AM daily): Combines weather, currency rates, and news
- **Voice Message Transcription**: Converts Telegram voice messages to text
- **Access Control**: Multi-tier authorization system with admin controls
- **Invite System**: Generate temporary invite links for easy user onboarding

### API Integrations
- **OpenAI**: GPT chat completions, audio transcription, text corrections
- **Weather API**: Multi-city weather forecasts with timezone support
- **Currency API**: Exchange rates with historical comparison
- **DeepL**: Professional translation services
- **Miniflux**: RSS feed aggregation for news

## Configuration

### Required Environment Variables
```
TG_TOKEN - Telegram bot token
TG_GROUP - Main group chat ID for morning digest
TG_ADMINUSERIDS - Comma-separated admin user IDs
WEATHERAPI_KEY - Weather service API key
CURRENCYAPI_KEY - Currency service API key
OPENAIAPI_KEY - OpenAI API key
DEEPLAPI_KEY - DeepL API key
MINIFLUXAPI_KEY - Miniflux API key
MINIFLUXAPI_BASEURL - Miniflux instance URL
REDIS_ADDR - Redis connection string
```

### Build & Deploy
```bash
# Local build
cd src && go build -mod vendor -o main .

# Docker build
docker-compose up -d
```

## Development Notes

### Code Organization
- All source code in `src/` directory
- Vendor dependencies included (exclude from reviews)
- Russian language used for user-facing messages
- Structured logging with slog
- Prometheus metrics integration

### Key Design Patterns
- Command pattern for bot handlers
- API client abstraction layer
- LRU caching for ChatGPT conversations
- Graceful error handling with user feedback
- Message splitting for Telegram length limits

### Authentication & Security
- **Multi-tier access control**:
  - Admins (TG_ADMINUSERIDS): Full access to user management commands
  - Authorized users (stored in Redis): Regular bot access
  - Group chat (TG_GROUP): Always allowed for morning digest
- **Invite system**: Temporary 24-hour tokens for secure user onboarding
- **Redis-based user management**: Persistent storage of authorized users
- No public access - family/private use only
- API key management through environment variables
- Private chat unauthorized access responses

### Current TODO Items
From `TODO.md`:
- [ ] Add support for image input in GPT calls
- [x] Add admin commands to manage access to bot (completed)

## Testing & Maintenance
- Monitor via `/metrics` endpoint (Prometheus)
- Health check via `/ping` endpoint
- Debug mode available via `DEBUG=true`
- Revision tracking via `REVISION` env var
- Log analysis through structured JSON logging

## Recent Changes
Based on git history:
- Refactored OpenAI integration with default model selection
- Improved response history formatting
- Added unauthorized access responses for private chats
- Disabled auto-cleanup of old ChatGPT messages
- Updated dependencies