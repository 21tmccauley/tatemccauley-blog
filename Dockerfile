# syntax=docker/dockerfile:1

# --- Stage 1: build the static website with Eleventy ---
FROM node:22-alpine AS web
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npx @11ty/eleventy          # → /app/_site

# --- Stage 2: build the SSH/TUI binary (posts embedded via embed.go) ---
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /blog-ssh ./tui

# --- Stage 3: minimal runtime ---
# Alpine (not distroless) keeps a shell so `fly ssh console` is usable for
# debugging. The app runs inside a per-app Firecracker microVM, and SSH
# visitors get only the TUI (no shell/exec), so the surface stays small.
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /blog-ssh /usr/local/bin/blog-ssh
COPY --from=web /app/_site /site
# Host key persists on a Fly volume mounted at /data (see fly.toml).
EXPOSE 8080 2222
CMD ["blog-ssh", "-serve", \
     "-host", "0.0.0.0", \
     "-port", "2222", \
     "-http-port", "8080", \
     "-site-dir", "/site", \
     "-host-key", "/data/host_ed25519"]
