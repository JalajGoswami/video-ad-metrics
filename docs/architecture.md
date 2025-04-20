# Architecture

## Tech Stack

- Go
- Postgres
- Docker

## Database Design

### Ad

- Table: `ads`

```json
{
  "id": "unique-ad-id",
  "name": "ad-name",
  "description": "ad-description",
  "image_url": "https://.../image.png",
  "target_url": "https://.../target",
  "created_at": "2025-01-01T00:00:00Z",
}
```
### Click

- Table: `clicks`

```json
{
  "id": "unique-click-id",
  "ad_id": "unique-ad-id", // foreign key
  "timestamp": "2025-01-01T00:00:00Z",
  "ip_address": "192.168.1.1",
  "playback_time": 10,
  "created_at": "2025-01-01T00:00:00Z",
}
```
> Note: `clicks` table will only store clicks for the current month. Older data will be archived routinely thus reducing latency for analytics queries.

- Table `archived_clicks`

```json
{
  // schema same as `clicks` table
}
```

> Note: Entries in `clicks` table will be moved to `archived_clicks` table after each month.

### Aggregated Analytics

- Table: `aggregated_analytics`

```json
{
  "id": "unique-aggregated-analytics-id",
  "ad_id": "unique-ad-id", // foreign key
  "total_clicks": 100,
  "total_playback_time": 1000,
  "updated_at": "2025-01-01T00:00:00Z",
  "created_at": "2025-01-01T00:00:00Z",
}
```
> Note: these aggregated analytics are maintained for quick retrieval.

- Table: `monthly_analytics`

```json
{
  "id": "unique-monthly-analytics-id",
  "ad_id": "unique-ad-id", // foreign key
  "month": 4, // 1-12
  "year": 2025,
  "total_clicks": 100, // of that month
  "total_playback_time": 1000, // of that month
  "created_at": "2025-01-01T00:00:00Z",
}
```
> Note: these monthly analytics are maintained for better flexibility and advanced queries.
