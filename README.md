# go-amule-webui

Modern web control panel for [aMule](https://github.com/amule-project/amule), written in **Go** + **Vue 3** and deployed as a Docker sidecar.

The aMule Docker image used throughout this project is based on [ngosang/docker-amule](https://github.com/ngosang/docker-amule).

---

## Release

Latest stable: **v0.1.1** — see [releases](https://github.com/neonpc/go-amule-webui/releases).

---

## Description

**go-amule-webui** replaces the legacy aMuleWeb interface with a modern, responsive single-page application. It communicates with aMule over the **ED2K EC (External Connection) protocol** (TCP `:4712`) and exposes a REST API + WebSocket endpoint for real-time browser updates.

### Features

- Real-time dashboard with download/upload speeds, queue status, and KPI cards
- Manage downloads (pause, resume, cancel) with live progress updates
- View active uploads per client
- Browse shared files via a recursive directory tree
- Search the ED2K network (Local, Global, Kad) and download results directly
- Server list management (connect, add, remove)
- Kad network status and connection control
- Statistics overview
- Application log viewer
- Preferences viewer (read-only)
- **Responsive design** — collapsible sidebar drawer on mobile, card layout on small screens
- **PWA** — installable on mobile/home screen, offline-capable via service worker
- Real-time updates via WebSocket with auto-reconnect
- Dark theme UI

### Technology Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24 |
| Frontend | Vue 3, Vue Router 4, Pinia, TypeScript |
| Build | Vite |
| PWA | vite-plugin-pwa, Workbox |
| Charts | Chart.js + vue-chartjs |
| WebSocket | gorilla/websocket |
| Styling | Scoped CSS with CSS custom properties (dark theme) |
| Container | Multi-stage Docker build (Alpine 3.20) |

### Responsive Behaviour

| Viewport | Layout |
|----------|--------|
| >768px | Fixed sidebar, full tables, compact metrics |
| ≤768px | Hamburger menu with slide-in drawer overlay, Downloads switches from table to card layout, touch-friendly button targets (≥38px) |

---

## Docker Images

Images are published to the GitHub Container Registry:

```
ghcr.io/neonpc/go-amule-webui
```

### Docker Tags

| Tag | Description |
|-----|-------------|
| `main` | Latest production build from the `main` branch |
| `dev` | Latest development build from the `dev` branch |
| `v<semver>` | Stable versioned release (e.g. `v0.1.0`, `v0.1.1`) |
| `<major>.<minor>` | Version prefix (e.g. `0.1`) |
| `<sha>` | Specific commit SHA |

### Supported Architectures

| Architecture | Notes |
|-------------|-------|
| `linux/amd64` | Built and tested on GitHub Actions (ubuntu-latest) |

---

## Application Settings

All configuration is passed via environment variables.

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `AMULE_HOST` | `127.0.0.1` | No | aMule EC hostname or IP |
| `AMULE_PORT` | `4712` | No | aMule EC TCP port |
| `AMULE_PASSWORD` | — | **Yes** | aMule EC password |
| `AMULE_WEB_DIR` | `/dist` | No | Path to the built frontend directory inside the container |
| `LISTEN` | `:8080` | No | API/HTTP listen address (port only or `host:port`) |

---

## Usage

### Docker Compose

```yaml
version: "3.8"
services:
  amule:
    image: ngosang/amule:latest
    container_name: amule
    ports:
      - "4712:4712"
      - "4665:4665"
      - "4672:4672"
    volumes:
      - /path/to/downloads:/incoming
      - /path/to/shared:/shared
      - ./amule-data:/home/amule/.aMule
    environment:
      - AMULE_PASSWORD=my_password
      - AMULE_CONFPATH=/home/amule/.aMule
    restart: unless-stopped

  amule-webui:
    image: ghcr.io/neonpc/go-amule-webui:main
    container_name: amule-webui
    ports:
      - "8080:8080"
    environment:
      - AMULE_HOST=amule
      - AMULE_PORT=4712
      - AMULE_PASSWORD=my_password
    depends_on:
      - amule
    restart: unless-stopped
```

### Docker CLI

```bash
docker run -d \
  --name amule-webui \
  -p 8080:8080 \
  -e AMULE_HOST=192.168.1.100 \
  -e AMULE_PORT=4712 \
  -e AMULE_PASSWORD=my_password \
  ghcr.io/neonpc/go-amule-webui:main
```

Then open **http://localhost:8080** in your browser.

---

## Architecture

```
┌──────────────┐   EC protocol    ┌──────────────┐   REST + WS   ┌──────────────┐
│   aMule      │ ◄──────────────► │  amule-api   │ ◄──────────► │ Vue 3 UI    │
│  (daemon)    │    TCP :4712     │  (Go sidecar)│   :8080      │  (browser)   │
└──────────────┘                  └──────────────┘              └──────────────┘
```

- aMule exposes the EC protocol on TCP port `4712` (configurable).
- The Go backend (`amule-api`) translates EC binary commands into a REST/JSON API and pushes real-time updates via WebSocket.
- The Vue 3 frontend is a SPA served by the same Go binary from `/dist`.

---

## Development

```bash
# Backend (requires a running aMule instance)
go run ./cmd/amule-api

# Frontend (separate terminal, hot-reload on :5173)
cd web && npm run dev
```

### Docker Dev Workflow

```bash
# Build and run the dev service
docker compose up -d amule-webui-dev

# After frontend changes, rebuild locally and restart
cd web && npm run build
docker compose restart amule-webui-dev
```
