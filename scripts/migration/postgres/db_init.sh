#!/bin/bash

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker exec xledger-postgres pg_isready -U admin -d postgres > /dev/null 2>&1; then
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# Check PostgreSQL version and configuration
echo "Checking PostgreSQL configuration..."
docker exec xledger-postgres psql -U admin -d postgres -c "SELECT version();"
docker exec xledger-postgres psql -U admin -d postgres -c "SHOW server_encoding;"
docker exec xledger-postgres psql -U admin -d postgres -c "SHOW client_encoding;"

# Create database (if not exists)
echo "Creating database..."
docker exec xledger-postgres psql -U admin -d postgres -c "DROP DATABASE IF EXISTS xledger;"
docker exec xledger-postgres psql -U admin -d postgres -c "CREATE DATABASE xledger WITH ENCODING 'UTF8' LC_COLLATE='en_US.utf8' LC_CTYPE='en_US.utf8';"

# Initialize database schema
echo "Initializing database schema..."
docker exec -i xledger-postgres psql -U admin -d xledger < scripts/migration/postgres/init.sql

# Verify database tables
echo "Verifying database tables..."
docker exec xledger-postgres psql -U admin -d xledger -c "\dt"

echo "Database initialization completed!"