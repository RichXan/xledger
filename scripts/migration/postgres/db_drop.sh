#!/bin/bash

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker exec xledger-postgres pg_isready -U admin -d xledger > /dev/null 2>&1; then
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# Drop database
echo "Dropping database..."
docker exec -i xledger-postgres psql -U admin -d postgres -c "DROP DATABASE IF EXISTS xledger;"

echo "Database dropped successfully!"
