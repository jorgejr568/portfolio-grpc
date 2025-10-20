# Portfolio gRPC Service

A high-performance gRPC service built in Go for managing portfolio data including skills, experiences, and education. Features dual interfaces with native gRPC and HTTP/REST endpoints via gRPC-Gateway.

## Features

- **gRPC API** - High-performance RPC endpoints for portfolio data
- **REST API** - HTTP/JSON endpoints via gRPC-Gateway
- **Protocol Buffers** - Type-safe API definitions
- **Metrics Collection** - StatsD integration for monitoring
- **Structured Logging** - Configurable logging with Zap
- **Health Checks** - Built-in gRPC health checking
- **CORS Support** - Configurable cross-origin resource sharing
- **Graceful Shutdown** - Clean server termination handling
- **Dependency Injection** - Organized service management with Uber Dig

## Architecture

```
portfolio-grpc/
├── protos/              # Protocol Buffer definitions
├── gen/                 # Generated code from protobuf
│   ├── go/             # Go gRPC/protobuf code
│   └── openapi/        # OpenAPI specifications
├── internal/
│   ├── server/         # gRPC server implementation
│   ├── repositories/   # Data access layer
│   ├── interceptors/   # gRPC middleware
│   ├── client/         # External clients (StatsD)
│   └── utils/          # Shared utilities
└── main.go             # Application entry point
```

## Prerequisites

- Go 1.25.3 or higher
- PostgreSQL database
- (Optional) StatsD server for metrics
- (Optional) Buf CLI for protobuf generation

## Installation

1. Clone the repository:
```bash
git clone https://github.com/jorgejr568/portfolio-grpc.git
cd portfolio-grpc
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables (create `.env` file):
```bash
DATABASE_URL=postgres://user:password@localhost:5432/portfolio?sslmode=disable
STATSD_ADDRESS=localhost:8125  # Optional
LOG_LEVEL=info                 # Optional: debug, info, warn, error
ALLOWED_ORIGIN=*              # Optional: CORS origin
```

## Running the Service

### Local Development

```bash
go run main.go
```

The service will start:
- gRPC server on `:50051`
- HTTP/REST gateway on `:8080`

### With Docker Compose

```bash
docker-compose up -d
```

This starts the StatsD service for metrics collection.

## API Endpoints

### REST API (HTTP/JSON)

Base URL: `http://localhost:8080`

#### Skills
- `GET /v1/skills` - List all skills
- `GET /v1/skills/{id}` - Get skill by ID

#### Experiences
- `GET /v1/experiences` - List all experiences
- `GET /v1/experiences/{id}` - Get experience by ID

#### Education
- `GET /v1/educations` - List all educations
- `GET /v1/educations/{id}` - Get education by ID

### gRPC API

Connect to `localhost:50051`

Services:
- `PortfolioService.GetAllSkills`
- `PortfolioService.GetSkill`
- `PortfolioService.GetAllExperiences`
- `PortfolioService.GetExperience`
- `PortfolioService.GetAllEducations`
- `PortfolioService.GetEducation`

## Development

### Regenerate Protocol Buffers

If you modify `.proto` files, regenerate the code:

```bash
buf generate
```

This generates:
- Go gRPC/protobuf code
- gRPC-Gateway HTTP mappings
- OpenAPI specifications

### Database Setup

The service expects a PostgreSQL database with the appropriate schema. Refer to the repository implementations in `internal/repositories/` for the expected table structures.

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `STATSD_ADDRESS` | StatsD server address (host:port) | Optional |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` |
| `ALLOWED_ORIGIN` | CORS allowed origin | `*` |

### Ports

- `:50051` - gRPC server
- `:8080` - HTTP/REST gateway

## Monitoring

The service includes:

- **StatsD Metrics** - Method calls, latency, and error rates
- **Structured Logging** - JSON-formatted logs with correlation IDs
- **Health Checks** - Standard gRPC health checking protocol
- **Reflection** - gRPC server reflection for debugging

## Testing

Test with grpcurl:

```bash
# List services
grpcurl -plaintext localhost:50051 list

# Call method
grpcurl -plaintext localhost:50051 jorgejr568.portfolio_grpc.PortfolioService/GetAllSkills
```

Test REST API:

```bash
curl http://localhost:8080/v1/skills
```

## License

This project is part of a personal portfolio.

## Author

Jorge Junior (@jorgejr568)