# Workflow Rules

## Two-Branch Strategy

- **main** — Production branch. Runs on port **8080**. Only merge stable, tested code.
- **dev** — Development branch. Runs on port **8081**. **All changes go here first.**

Never commit or push directly to `main`. Always work on `dev`, test, then merge to `main`.

## Repositories

| Branch | Local path | Docker service | Port |
|--------|-----------|----------------|------|
| main   | `/root/go-amule-webui`     | `amule-webui`     | 8080 |
| dev    | `/root/go-amule-webui-dev` | `amule-webui-dev` | 8081 |

## Docker

Both branches run simultaneously via `/root/docker-compose.yml`:

```bash
# Start both
docker compose -f /root/docker-compose.yml up -d

# Rebuild dev after changes
docker compose -f /root/docker-compose.yml build amule-webui-dev
docker compose -f /root/docker-compose.yml up -d amule-webui-dev

# Rebuild main after merge
docker compose -f /root/docker-compose.yml build amule-webui
docker compose -f /root/docker-compose.yml up -d amule-webui
```

## Development Workflow

1. Work in `/root/go-amule-webui-dev` (already on `dev` branch)
2. Make changes, commit, push to `origin/dev`
3. The CI builds the dev image automatically on push to `dev`
4. Test on http://localhost:8081
5. When stable, create a PR from `dev` → `main` and merge
6. After merge, rebuild `amule-webui` on host to update production at http://localhost:8080
7. Tag the release on `main` (e.g. `v0.1.2`) for GitHub releases

## CI/CD

`.github/workflows/ci.yml` triggers on:
- Push to `main` — builds + pushes Docker image to `ghcr.io`
- Push to `dev` — builds only (no push)
- Tag `v*` — builds + pushes + creates GitHub Release
