# Multi-stage production build for unified full-stack Live Polling application

# Stage 1: Build React Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY backend/go.mod backend/go.sum* ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Stage 3: Unified Production Runtime
FROM alpine:3.20
WORKDIR /app

# Install Redis, CA certificates & tzdata
RUN apk --no-cache add ca-certificates tzdata redis

# Copy Go binary
COPY --from=backend-builder /app/server /app/server

# Copy Frontend static assets for Go SPA serving
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Copy Entrypoint Script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Default environment configuration
ENV PORT=10000
ENV REDIS_URL=redis://127.0.0.1:6379
ENV MONGO_DB=livepolling
ENV GIN_MODE=release

EXPOSE 10000

ENTRYPOINT ["/app/docker-entrypoint.sh"]
