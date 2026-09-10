# Build the two Svelte frontends.
FROM node:20-alpine AS frontend
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./frontend/
COPY dm-frontend/package.json dm-frontend/package-lock.json ./dm-frontend/
RUN npm --prefix frontend ci && npm --prefix dm-frontend ci
COPY frontend/ ./frontend/
COPY dm-frontend/ ./dm-frontend/
RUN npm --prefix frontend run build && npm --prefix dm-frontend run build

# Build the Go server (CGO required for go-sqlite3).
FROM golang:1.25 AS backend
WORKDIR /build
COPY backend/ ./
RUN CGO_ENABLED=1 go build -o /out/server ./cmd/server

# Minimal runtime image with glibc for the CGO sqlite binary.
FROM debian:bookworm-slim AS runtime
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*
COPY --from=backend /out/server ./server
COPY --from=frontend /src/frontend/dist ./static/player
COPY --from=frontend /src/dm-frontend/dist ./static/dm
ENV PLAYER_DIST=/app/static/player \
    DM_DIST=/app/static/dm
EXPOSE 8000
CMD ["./server"]