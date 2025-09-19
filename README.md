# Librus Telegram Bot

A Telegram bot that fetches messages and news from Librus (Polish school system) and forwards them to users via Telegram.

## Architecture

The project consists of two main services:

1. **Go Bot Service** (`bot`) - Handles Telegram interactions, user management, and message processing
2. **Node.js Scraper Service** (`librus-scraper`) - Handles web scraping via Playwright and exposes gRPC API

## Features

- 📩 Fetch personal messages from Librus
- 🔔 Fetch news/announcements 
- 🌍 Multi-language support with automatic translation
- 📎 Attachment handling (files, images, videos)
- 💬 Reply to messages directly from Telegram
- 🔄 Periodic automatic updates
- 🗄️ MongoDB storage for users and message history

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for development)
- Node.js 18+ (for development)

### Environment Setup

1. Copy environment file:
```bash
cp .env.example .env
```

2. Fill in required environment variables:
```bash
TELEGRAM_TOKEN=your_telegram_bot_token
MONGO_EXPRESS_PASSWORD=your_mongo_password
OPEN_AI_KEY=your_openai_key  # Optional, for OpenAI translations
```

### Running with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f bot
docker-compose logs -f librus-scraper

# Stop services
docker-compose down
```

### Development Setup

1. **Start dependencies:**
```bash
docker-compose up -d mongodb librus-scraper
```

2. **Run Go bot locally:**
```bash
go run .
```

3. **For local development with port access:**
```bash
cp docker-compose.override.yml.example docker-compose.override.yml
# Edit ports as needed
docker-compose up -d
```

## Usage

### Bot Commands

1. **Login:** Send `login:password` to authenticate
2. **Manual Update:** Send `/update` to force check for new messages
3. **URL Processing:** Send `url:https://synergia.librus.pl/...` to fetch specific message
4. **Reply:** Reply to any message to send response back to Librus
5. **Language:** Send language code (e.g., `en`, `uk`, `de`) to set translation language

### Supported Languages

- Polish (pl) - original
- English (en)
- Ukrainian (uk) 
- German (de)
- And many others via Google Translate or OpenAI

## Development

### Project Structure

```
├── main.go                 # Entry point
├── telegram/              # Telegram bot logic
│   ├── bot.go
│   ├── periodic_processing.go
│   └── handler/           # Message handlers
├── pkg/
│   ├── grpc_client/       # gRPC client for scraper service
│   └── config/            # Configuration management
├── model/                 # Data models
├── mongo/                 # MongoDB operations
├── translator/            # Translation services
├── proto/                 # Protobuf definitions and generated code
├── playwright/            # Node.js scraper service
└── docs/                  # Documentation
```

### gRPC API

The scraper service exposes these gRPC methods:

- `ValidateLogin(login, password)` - Validate user credentials
- `GetMessages(login, password)` - Fetch personal messages
- `GetNews(login, password)` - Fetch news/announcements  
- `GetSingleMessage(login, password, url)` - Fetch specific message
- `AnswerMessage(login, password, url, text)` - Reply to message
- `GetAllUpdates(login, password)` - Fetch both messages and news
- `HealthCheck()` - Service health status

### Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./pkg/grpc_client -v

# Integration tests (requires services running)
INTEGRATION_TESTS=true go test ./pkg/grpc_client -v

# Generate protobuf files
make proto-gen
```

### Database

- **MongoDB** stores users, messages, and translation cache
- **Mongo Express** available at http://localhost:12328 (in development)

## Deployment

The project is designed to run in Docker containers. See `docker-compose.yaml` for production configuration.

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `TELEGRAM_TOKEN` | Telegram bot token | Yes |
| `MONGO_EXPRESS_PASSWORD` | MongoDB admin password | Yes |
| `OPEN_AI_KEY` | OpenAI API key for translations | No |
| `GRPC_SCRAPER_ADDRESS` | gRPC scraper service address | No |
| `ATTACHMENTS_BASE_DIR` | Directory for attachments | No |
| `LOG_LEVEL` | Logging level for scraper | No |

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## License

This project is for educational purposes. Please respect Librus terms of service.
