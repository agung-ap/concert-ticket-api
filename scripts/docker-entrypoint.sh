#!/bin/sh
set -e

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until nc -z db 5432; do
    echo "PostgreSQL not ready yet - sleeping"
    sleep 1
done
echo "PostgreSQL is ready!"

# Start the application
echo "Starting Concert Ticket API..."
exec "/app/concert-ticket-api"