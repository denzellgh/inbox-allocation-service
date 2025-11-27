#!/bin/bash
set -e

# This script runs when the PostgreSQL container starts for the first time
# It enables the uuid-ossp extension for UUID support

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    ALTER USER $POSTGRES_USER WITH PASSWORD '$POSTGRES_PASSWORD';
EOSQL

echo "Database initialized with uuid-ossp extension and user password set"
