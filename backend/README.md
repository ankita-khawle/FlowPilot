# FlowPilot API

Minimal Go HTTP API scaffold. It currently exposes `GET /api/health` and has no external dependencies.

## Run

From the repository root:

```bash
cd backend
go run ./cmd/api
```

Then check `http://localhost:8080/api/health`.

## Test

```bash
cd backend
go test ./...
```

Set `HTTP_ADDR` to override the listen address. Database and queue configuration will be introduced when those integrations are implemented.
