# Benchphant API Documentation

## API Overview

Benchphant provides a RESTful API for managing database connections and running performance tests. All API endpoints are prefixed with `/api/v1/`.

## Authentication

Currently, the API uses basic authentication. All requests must include valid credentials.

## API Endpoints

### Connections

#### List Connections
```http
GET /api/v1/connections
```

Returns a list of all configured database connections.

**Response**
```json
[
  {
    "id": "string",
    "name": "string",
    "host": "string",
    "port": number,
    "username": "string",
    "database": "string",
    "type": "string",
    "tags": ["string"],
    "created_at": "string",
    "updated_at": "string"
  }
]
```

#### Create Connection
```http
POST /api/v1/connections
```

Creates a new database connection.

**Request Body**
```json
{
  "name": "string",
  "host": "string",
  "port": number,
  "username": "string",
  "password": "string",
  "database": "string",
  "type": "string",
  "tags": ["string"]
}
```

**Response**
```json
{
  "id": "string",
  "name": "string",
  "host": "string",
  "port": number,
  "username": "string",
  "database": "string",
  "type": "string",
  "tags": ["string"],
  "created_at": "string",
  "updated_at": "string"
}
```

#### Test Connection
```http
POST /api/v1/connections/test
```

Tests if a database connection is valid.

**Request Body**
```json
{
  "host": "string",
  "port": number,
  "username": "string",
  "password": "string",
  "database": "string",
  "type": "string"
}
```

**Response**
```json
{
  "success": boolean,
  "message": "string"
}
```

### Benchmarks

#### Start Benchmark
```http
POST /api/v1/benchmarks
```

Starts a new benchmark test.

**Request Body**
```json
{
  "connection_id": "string",
  "type": "string",
  "duration": number,
  "threads": number,
  "parameters": {
    "key": "value"
  }
}
```

**Response**
```json
{
  "id": "string",
  "status": "string",
  "start_time": "string"
}
```

#### Get Benchmark Status
```http
GET /api/v1/benchmarks/{id}
```

Gets the current status and metrics of a running benchmark.

**Response**
```json
{
  "id": "string",
  "status": "string",
  "metrics": {
    "qps": number,
    "latency": {
      "avg": number,
      "p50": number,
      "p95": number,
      "p99": number
    },
    "errors": number
  }
}
```

## Error Responses

All endpoints may return the following error responses:

```json
{
  "error": "string"
}
```

Common HTTP status codes:
- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 404: Not Found
- 500: Internal Server Error

## Rate Limiting

API requests are limited to 100 requests per minute per IP address.

## Versioning

The current API version is v1. The version is included in the URL path.

## Data Types

- `id`: UUID string
- `created_at`, `updated_at`: ISO 8601 datetime strings
- `type`: Enum string ("mysql" | "postgresql")
- `port`: Integer (1-65535)
- `tags`: Array of strings
