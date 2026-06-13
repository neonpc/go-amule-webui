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
ENTRYPOINT ["/amule-api"]
