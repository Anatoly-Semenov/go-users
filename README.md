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
- [License](#license)

---

## Project Description

Go-Users is a user management service with authentication and authorization capabilities. The service provides APIs through both HTTP and gRPC protocols, allowing it to be integrated with various systems.

Key features:
- User registration and management
- Authentication using JWT tokens
- Authorization and access control management
- PostgreSQL database integration

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

## License

This project is distributed under the MIT license. See the LICENSE file for more information.
