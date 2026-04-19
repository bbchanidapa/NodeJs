# HTTP API (Go)

Small service using the standard library:

- `GET /` → `Hello World!` (plain text)
- `GET /health` → `{"status":"ok"}` (JSON)

## Run locally

```bash
go run ./cmd/server
```

Optional: `PORT` (default `3000`).

## Test

```bash
go test ./...
```

## Docker

```bash
docker build -t api .
docker run --rm -p 3000:3000 api
```
