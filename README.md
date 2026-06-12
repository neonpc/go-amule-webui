# go-amule-webui

Modern web control panel for aMule, written in Go + Vue 3 and deployed as a Docker sidecar.

## Architecture

```
┌──────────────┐   EC protocol    ┌──────────────┐   REST + WS   ┌──────────────┐
│   aMule      │ ◄──────────────► │  amule-api   │ ◄──────────► │ Vue 3 UI    │
│  (amuleweb)  │    TCP :4712     │  (Go, sidecar)│   :8080      │  (browser)   │
└──────────────┘                  └──────────────┘              └──────────────┘
```

## Usage

```bash
# Run alongside aMule
docker compose -f docker-compose.sidecar.yml up -d

# Open http://localhost:8080
```

## Development

```bash
# Backend
go run ./cmd/amule-api

# Frontend (separate terminal)
cd web && npm run dev
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AMULE_HOST` | `127.0.0.1` | aMule EC host |
| `AMULE_PORT` | `4712` | aMule EC port |
| `AMULE_PASSWORD` | — | aMule EC password (required) |
| `LISTEN` | `:8080` | API listen address |
