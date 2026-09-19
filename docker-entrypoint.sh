#!/bin/sh
set -e

# If REDIS_URL points to localhost/127.0.0.1 or is empty, start local Redis server
if [ -z "$REDIS_URL" ] || [ "$REDIS_URL" = "redis://127.0.0.1:6379" ] || [ "$REDIS_URL" = "redis://localhost:6379" ]; then
    echo ">> [ENTRYPOINT] Starting internal Redis server on 127.0.0.1:6379..."
    redis-server --daemonize yes --protected-mode no
    export REDIS_URL="redis://127.0.0.1:6379"
fi

echo ">> [ENTRYPOINT] Launching Live Polling Full-Stack Service on port ${PORT:-10000}..."
exec /app/server
