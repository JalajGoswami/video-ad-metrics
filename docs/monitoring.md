# Basic Prometheus Monitoring

This project includes a minimal Prometheus monitoring setup to track the most important application metrics.

## Available Metrics

The following metrics are exposed through the `/metrics` endpoint:

### HTTP Metrics
- `http_requests_total` - Total number of HTTP requests by method, path, and status code
- `http_request_duration_seconds` - Duration of HTTP requests in seconds by method and path

### Database Metrics
- `database_connections` - Number of active database connections
- `clicks_logged_total` - Total number of ad clicks logged (can be used to get rate of clicks logged)

## Prometheus Configuration

To scrape these metrics with Prometheus, add the following job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'video-ad-metrics'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:5000']
```

## Grafana Dashboard

You can create a simple Grafana dashboard to visualize these metrics. Here's a basic example:

1. Add Prometheus as a data source in Grafana
2. Create a new dashboard
3. Add panels for:
   - HTTP request rate (counter)
   - HTTP response time (histogram)
   - Database connections (gauge)
   - Click logging rate (counter)