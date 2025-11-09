# Goff Examples

This directory contains example applications demonstrating how to use the goff feature flag library.

## Admin Server

The admin server (`examples/admin/`) provides a web UI and REST API to manage feature flags stored in SQLite. Changes are automatically synced to a YAML file that can be consumed by the goff client.

### Features

- **Web UI** - Beautiful, modern interface for managing flags
- SQLite database for flag persistence
- REST API for CRUD operations on flags
- Automatic YAML sync for goff client consumption
- Initial seed data with example flags

### Usage

```bash
cd examples/admin
go run . -db flags.db -yaml flags.yaml -port 8081
```

Then open your browser to **http://localhost:8081** to access the admin UI.

### API Endpoints

- `GET /flags` - List all flags
- `GET /flags/{key}` - Get a specific flag
- `POST /flags` - Create a new flag
- `PUT /flags/{key}` - Update a flag
- `DELETE /flags/{key}` - Delete a flag

### Example: Create a Flag

```bash
curl -X POST http://localhost:8081/flags \
  -H "Content-Type: application/json" \
  -d '{
    "key": "new_feature",
    "enabled": true,
    "type": "bool",
    "variants": "{\"true\": 50, \"false\": 50}",
    "rules": "[]",
    "default": "false"
  }'
```

### Example: Update a Flag

```bash
curl -X PUT http://localhost:8081/flags/enable_logging \
  -H "Content-Type: application/json" \
  -d '{
    "key": "enable_logging",
    "enabled": true,
    "type": "bool",
    "variants": "{\"true\": 100, \"false\": 0}",
    "rules": "[]",
    "default": "true"
  }'
```

## App Server

The app server (`examples/app/`) demonstrates how to use goff with the `With...` methods to evaluate flags and change application behavior.

### Features

- Uses `WithFile` to load flags from YAML
- Uses `WithAutoReload` for automatic flag updates
- Uses `WithHooks` for observability
- Multiple endpoints with different flag-based behaviors

### Usage

```bash
cd examples/app
go run . -yaml ../admin/flags.yaml -port 8080
```

### Endpoints

- `GET /checkout?user={id}&plan={plan}` - Checkout flow with flag-based behavior
- `GET /api/users?user={id}` - User API with flag-based logging
- `GET /api/features?user={id}` - Feature availability endpoint
- `GET /health` - Health check

### Example Requests

```bash
# Checkout with new flow
curl "http://localhost:8080/checkout?user=123&plan=pro"

# Users API with logging
curl "http://localhost:8080/api/users?user=456"

# Feature flags
curl "http://localhost:8080/api/features?user=789"
```

## Running Both Servers

1. Start the admin server:
```bash
cd examples/admin
go run . &
```

2. Start the app server:
```bash
cd examples/app
go run . &
```

3. Modify flags via the admin API, and watch the app server automatically pick up changes (thanks to `WithAutoReload`).

## Flag Structure

Flags in the database use JSON strings for complex fields:

- `variants`: JSON object mapping variant names to percentages (0-100)
- `rules`: JSON array of rule objects with `when` and `then` conditions
- `default`: JSON value (boolean for bool flags, string for string flags)

Example flag:
```json
{
  "key": "new_checkout",
  "enabled": true,
  "type": "bool",
  "variants": "{\"true\": 50, \"false\": 50}",
  "rules": "[{\"when\": {\"all\": [{\"attr\": \"plan\", \"op\": \"eq\", \"value\": \"pro\"}]}, \"then\": {\"variants\": {\"true\": 90, \"false\": 10}}}]",
  "default": "false"
}
```

