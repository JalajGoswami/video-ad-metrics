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

2. Build Docker Image

```bash
docker build -t video-ad-metrics .
```

3. Run Docker Container

```bash
docker run -d -p 5000:5000 video-ad-metrics
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

### Instructions

1. Clone the repository

```bash
git clone https://github.com/JalajGoswami/video-ad-metrics.git
```

2. Create a Postgres database (first time only)

```bash
make create-db
# creates a new database with mock data
# skip this step if database already exists as it will drop the existing database
```

3. Install Dependencies

```bash
go mod tidy
```

4. Run the service (in dev mode)

```bash
make run
```

5. Test the service

```bash
make test
```

## API Documentation

See [API Documentation](docs/api-specs.md)

## Architecture

See [Architecture](docs/architecture.md)
