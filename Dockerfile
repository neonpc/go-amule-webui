FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package.json ./
RUN npm install
COPY web/ .
RUN npm run build

FROM golang:1.24-alpine AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /amule-api ./cmd/amule-api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=backend /amule-api /amule-api
COPY --from=frontend /web/dist /dist
EXPOSE 8080
LABEL org.opencontainers.image.title="go-amule-webui" \
      org.opencontainers.image.description="Modern web control panel for aMule — P2P file-sharing client for the eD2K (eDonkey) and Kad networks" \
      org.opencontainers.image.url="https://github.com/neonpc/go-amule-webui" \
      org.opencontainers.image.source="https://github.com/neonpc/go-amule-webui" \
      org.opencontainers.image.licenses="GPL-2.0" \
      org.opencontainers.image.tags="p2p,file-sharing,ed2k,amule,emule,kademlia,edonkey"
ENTRYPOINT ["/amule-api"]
