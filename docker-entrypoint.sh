#!/bin/sh
set -e

# If REDIS_URL points to localhost/127.0.0.1 or is empty, start local Redis server
if [ -z "$REDIS_URL" ] || [ "$REDIS_URL" = "redis://127.0.0.1:6379" ] || [ "$REDIS_URL" = "redis://localhost:6379" ]; then
    echo ">> [ENTRYPOINT] Starting internal Redis server on 127.0.0.1:6379..."
    redis-server --daemonize yes --protected-mode no
    export REDIS_URL="redis://127.0.0.1:6379"
fi

# Diagnostic check for MONGO_URI in cloud container
if [ -z "$MONGO_URI" ] || [ "$MONGO_URI" = "mongodb://127.0.0.1:27017" ] || [ "$MONGO_URI" = "mongodb://localhost:27017" ]; then
    echo ">> [ENTRYPOINT] NOTICE: MONGO_URI is set to local address ($MONGO_URI)."
    echo ">> [ENTRYPOINT] If deploying on Render or in the cloud, please set MONGO_URI in your dashboard environment variables"
    echo ">> [ENTRYPOINT] to your MongoDB Atlas connection string: mongodb+srv://<user>:<password>@cluster0.mongodb.net/livepolling?retryWrites=true&w=majority"
fi

echo ">> [ENTRYPOINT] Launching Live Polling Full-Stack Service on port ${PORT:-10000}..."
exec /app/server
