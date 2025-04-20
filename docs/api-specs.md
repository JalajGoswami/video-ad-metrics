## API Documentation

### Ad Management

#### Create Ad

- Endpoint: `POST /ads`
- Request Body:

```json
{
  "name": "ad-name",
  "description": "ad-description", // optional
  "image_url": "https://.../image.png",
  "target_url": "https://.../target"
}
```

- Response:

```json
{
  "success": true,
  "message": "Ad created successfully",
  "trace_id": "unique-trace-id", // for tracing the flow of request
  "result": {
    "id": "unique-ad-id",
    "name": "ad-name",
    "description": "ad-description",
    "image_url": "https://.../image.png",
    "target_url": "https://.../target",
    "created_at": "2025-01-01T00:00:00Z",
  }
}
```

#### Get All Ads

- Endpoint: `GET /ads`
- Query Params:
  - `page`: number - default: 1
  - `rows`: number - default: 25 (max: 100)
  - `order`: `asc` or `desc` - default: `desc` (by created_at)
  - `search`: string - optional (search by name case insensitive)
- Response:

```json
{
  "success": true,
  "message": "Request successful",
  "trace_id": "unique-trace-id", // for tracing the flow of request
  "result": {
    "pages": {
      "page_number": 1,
      "total_pages": 10,
      "page_size": 25,
    },
    "values": [
        {
            "id": "unique-ad-id",
            "name": "ad-name",
            "description": "ad-description",
            "image_url": "https://.../image.png",
            "target_url": "https://.../target",
            "created_at": "2025-01-01T00:00:00Z"
        }
    ]
  }
}
```

#### Get Ad by ID

- Endpoint: `GET /ads/:id`
- Response:

```json
{
  "success": true,
  "message": "Request successful",
  "trace_id": "unique-trace-id",
  "result": {
    "id": "unique-ad-id",
    "name": "ad-name",
    "description": "ad-description",
    "image_url": "https://.../image.png",
    "target_url": "https://.../target",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

### Click Tracking

#### Track Click

- Endpoint: `POST /ads/clicks`
- Request Body:

```json
{
  "ad_id": "unique-ad-id",
  "timestamp": "2025-01-01T00:00:00Z",
  "ip_address": "192.168.1.1",
  "playback_time": 10 // in seconds
}
```

- Response:

```json
{
  "success": true,
  "message": "Click tracked successfully",
  "trace_id": "unique-trace-id"
}

### Ads Performance & Analytics

#### Get Ads Analytics

- Endpoint: `GET /ads/analytics`
- Query Params:
  - `period`: `minute`, `hour`, `day`, `week`, `month` - default: `hour`
- Response:

```json
{
  "success": true,
  "message": "Request successful",
  "trace_id": "unique-trace-id",
  "result": {
    "total_clicks": 100, // click count so far of all ads
    "average_clicks_per_ad": 10, // average click count per ad
    "total_playback_time": 1000, // total playback time of all ads
    "average_playback_time": 10, // average playback time of all ads
    "period": "hour", // range of the analytics
    "total_clicks_in_range": 40, // click count in the given range (e.g. last hour)
    "average_clicks_per_ad_in_range": 4, // average clicks per ad in the given range
    "total_playback_time_in_range": 400, // total playback time of ads in the given range
    "average_playback_time_per_ad_in_range": 4, // average playback time per ad in the given range
  }
}
```

#### Get Ad Analytics

- Endpoint: `GET /ads/analytics/:id`
- Query Params:
  - `period`: `minute`, `hour`, `day`, `week`, `month` - default: `hour`
- Response:

```json
{
  "success": true,
  "message": "Request successful",
  "trace_id": "unique-trace-id",
  "result": {
    "total_clicks": 100, // click count so far of this ad
    "total_playback_time": 1000, // total playback time of this ad
    "average_playback_time": 10, // average playback time of this ad
    "period": "hour", // range of the analytics
    "total_clicks_in_range": 40, // click count in the given range (e.g. last hour)
    "total_playback_time_in_range": 400, // total playback time of this ad in the given range
    "average_playback_time_in_range": 4, // average playback time of this ad in the given range
  }
}
```
