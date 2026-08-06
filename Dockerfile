FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS build
RUN apk add --no-cache build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/static/dist ./static/dist
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/recommendli ./main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates su-exec tzdata \
    && addgroup -S recommendli \
    && adduser -S -G recommendli recommendli
WORKDIR /app
COPY --from=build --chown=recommendli:recommendli /out/recommendli ./recommendli
COPY --from=build --chown=recommendli:recommendli /src/migrations ./migrations
COPY --from=build --chown=recommendli:recommendli /src/static/dist ./static/dist
COPY --chmod=755 deploy/docker-entrypoint.sh /usr/local/bin/docker-entrypoint
EXPOSE 9999
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O - "http://127.0.0.1:${PORT:-9999}/status" >/dev/null || exit 1
ENTRYPOINT ["docker-entrypoint"]
