# go-notification-service

> Part of the SaaS Framework - Extracted from monorepo

## Description

[Add service description here]

## Features

- Feature 1
- Feature 2
- Feature 3

## Prerequisites

- Go 1.24+
- MongoDB 4.4+ (if applicable)
- Redis 6.0+ (if applicable)
- RabbitMQ 3.9+ (if applicable)

## Installation

```bash
# Clone the repository
git clone https://github.com/vhvcorp/go-notification-service.git
cd go-notification-service

# Install dependencies
go mod download
```

## Configuration

Copy the example environment file and update with your values:

```bash
cp .env.example .env
```

See [DEPENDENCIES.md](docs/DEPENDENCIES.md) for a complete list of environment variables.

## Development

### Running Locally

```bash
# Run the service
make run

# Or with go run
go run cmd/main.go
```

### Running with Docker

```bash
# Build and run
make docker-build
make docker-run
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

### Linting

```bash
# Run linters
make lint

# Format code
make fmt
```

## API Documentation

See [docs/API.md](docs/API.md) for API documentation.

## Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for deployment instructions.

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for architecture details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Related Repositories

- [go-shared](https://github.com/vhvcorp/go-shared) - Shared Go libraries

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

- Documentation: [Wiki](https://github.com/vhvcorp/go-notification-service/wiki)
- Issues: [GitHub Issues](https://github.com/vhvcorp/go-notification-service/issues)
- Discussions: [GitHub Discussions](https://github.com/vhvcorp/go-notification-service/discussions)
