# Notification Service

A microservice responsible for handling all types of notifications (email, SMS, push) in the system.

## Features

- Event-driven notification processing
- Support for multiple notification channels:
  - Email (SendGrid, SMTP)
  - SMS (future)
  - Push Notifications (future)
- Template-based message generation
- Localization support
- Notification history tracking
- Rate limiting and throttling
- Delivery status tracking
- Retry mechanism for failed notifications

## Architecture

```
notification-service/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── domain/            # Domain model and interfaces
│   ├── infrastructure/    # External services implementation
│   ├── application/       # Application services
│   └── interfaces/        # Interface adapters (HTTP, gRPC)
├── pkg/                   # Public libraries
└── templates/             # Message templates
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker
- Kafka
- Redis

### Installation

1. Clone the repository:
```bash
git clone https://github.com/mibrahim2344/notification-service.git
```

2. Install dependencies:
```bash
go mod download
```

3. Set up configuration:
```bash
cp config/default.example.json config/default.json
```

4. Run the service:
```bash
make run
```

## Configuration

The service uses a configuration file located in `config/default.json`. Key configuration options include:

- Kafka connection settings
- Redis connection settings
- Email provider configurations
- Template settings
- Rate limiting parameters

## API Documentation

### Event Subscriptions

The service subscribes to the following Kafka topics:
- `user.registered`
- `user.verified`
- `user.password.reset`
- `user.password.changed`
- `user.deleted`

### REST Endpoints

- `POST /api/v1/notifications/send` - Manual notification sending
- `GET /api/v1/notifications/{id}` - Get notification status
- `GET /api/v1/notifications/history` - Get notification history

## Development

### Running Tests

```bash
make test
```

### Building Docker Image

```bash
make docker-build
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
