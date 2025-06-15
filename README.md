# go-users

---

## Table of Contents
- [Project Description](#project-description)
- [Project Structure](#project-structure)
- [Requirements](#requirements)
- [Installation and Setup](#installation-and-setup)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
  - [HTTP Server](#http-server)
  - [gRPC Server](#grpc-server)
  - [Migrations](#migrations)
- [Security Features](#security-features)
  - [IP Blocking](#ip-blocking)
  - [Brute Force Protection](#brute-force-protection)
- [License](#license)

---

## Project Description

Go-Users is a user management service with authentication and authorization capabilities. The service provides APIs through both HTTP and gRPC protocols, allowing it to be integrated with various systems.

Key features:
- User registration and management
- Authentication using JWT tokens
- Authorization and access control management
- PostgreSQL database integration
- IP-based security with brute force protection
- Temporary and permanent IP blocking

---


## Project Structure

```
go-users/
├── api/                   # API definitions and specifications
│   └── v1/                # API version 1 (OpenAPI/Swagger specs, proto files)
├── cmd/                   # Application entry points
│   └── app/               # Main application
│       └── cli/           # CLI commands for server startup and admin operations
├── internal/              # Application-specific internal code
│   ├── app/               # Application initialization and middleware
│   ├── config/            # Configuration parsing and processing
│   ├── domain/            # Domain models and business logic
│   │   ├── auth/          # Authentication domain models and business logic
│   │   └── user/          # User domain models and validation methods
│   ├── infrastructure/    # External systems adapters
│   │   ├── auth/          # Authentication implementations (JWT, sessions)
│   │   ├── repository/    # Data storage implementations (PostgreSQL adapters)
│   │   └── server/        # Server implementations
│   │       ├── http/      # HTTP server (routes, handlers, middleware)
│   │       └── grpc/      # gRPC server (services, interceptors)
│   │           └── gen/   # Auto-generated gRPC code
├── migrations/            # Database migration scripts
├── pkg/                   # Reusable packages
│   └── logger/            # Logging wrapper
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile             # Docker image build instructions
├── go.mod                 # Go dependencies
├── go.sum                 # Dependency hashes
├── LICENSE                # Project license
├── Makefile               # Build and deployment commands
└── README.md              # Project documentation
```

---

## Requirements

To run the application, you need:
- Go 1.18 or higher
- PostgreSQL 13 or higher
- Redis 6 or higher
- Docker (optional)
- Docker Compose (optional)

---

## Installation and Setup

1. Clone the repository:
```bash
git clone https://github.com/anatoly_dev/go-users.git
cd go-users
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
make build
```

Or using Docker:
```bash
make docker-build
```

---

## Configuration

example .env file:
```
# Server
HOST=0.0.0.0
HTTP_PORT=8080
GRPC_PORT=50051

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=users
DB_SSL_MODE=disable

# Redis (for IP blocking)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET_KEY=your-secret-key
JWT_EXPIRATION_HOURS=24

# Application
LOG_LEVEL=info
ENVIRONMENT=development
```

---

## Running the Application

### HTTP Server

```bash
make run-http
```

### gRPC Server

```bash
make run-grpc
```

### Migrations

Apply database migrations:
```bash
make migrate
```

---

## Security Features

### IP Blocking

The service implements a dual-layer IP blocking mechanism:

- **Temporary Blocks**: Stored in Redis with configurable expiry times. These are created automatically when suspicious activity is detected.
- **Permanent Blocks**: Stored in PostgreSQL without expiry. These can be created automatically after multiple violations or manually by administrators.

IP blocking is enforced through middleware that checks each request against both blocking mechanisms.

### Brute Force Protection

The service includes an advanced traffic analyzer to protect against brute force attacks:

- **Login Attempt Tracking**: Redis-based tracking of login attempts with sliding window time frames
- **Automatic Blocking**: IPs exceeding configured thresholds are automatically blocked
- **Progressive Penalties**: Repeated violations lead to longer blocks and potential permanent banning
- **Block Escalation**: System can automatically escalate from temporary to permanent blocks based on repeated violations

Default brute force protection configuration:
- 5 failed login attempts within 5 minutes result in a 30-minute temporary block
- 3 temporary blocks within 24 hours trigger a permanent block recommendation

The protection is active for both HTTP and gRPC interfaces.

---

## License

This project is distributed under the MIT license. See the LICENSE file for more information.
