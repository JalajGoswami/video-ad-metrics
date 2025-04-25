# Video Ad Metrics

This is a Go based service that allows you to manage video ads and track their performance. It supports Ad Management, Click Tracking, Analytics and Long Term Reports.

## Setup/Run Instructions

### Prerequisites

- Git
- Docker

### Instructions

1. Clone the repository

```bash
git clone https://github.com/JalajGoswami/video-ad-metrics.git
```

2. Build Docker Images

```bash
docker compose build
```

3. Run Docker Containers (Postgres and Service)

```bash
docker compose up -d
```

4. Test the service

```bash
curl -X GET http://localhost:5000/health
```

## Development Environment Setup

### Prerequisites

- Git
- Go
- Postgres
- Air (for hot reloading)

### Instructions

1. Clone the repository

```bash
git clone https://github.com/JalajGoswami/video-ad-metrics.git
```

2. Install Dependencies

```bash
go mod tidy
```

3. Install Air

```bash
go install github.com/air-verse/air@latest
```

4. Run postgres database

```bash
docker compose up -d postgres
```

5. Create a Postgres database (first time only)

```bash
go run cmd/create-db/main.go
# creates a new database with mock data
# skip this step if database already exists as it will drop the existing database
```


6. Run the service (in dev mode)

```bash
air
```

7. Test the service

```bash
curl -X GET http://localhost:5000/health
```

## API Documentation

See [API Documentation](docs/api-specs.md)

## Architecture

See [Architecture](docs/architecture.md)
