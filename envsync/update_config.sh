#!/usr/bin/env bash

DB_CONTAINER="postgres"
DB_USER="postgres"
DB_NAME="envsync_db"

KEY=$1
VALUE=$2

if [ -z "$KEY" ] || [ -z "$VALUE" ]; then
    echo "Missing key or value"
    exit 1
fi

echo "Pushing $KEY=$VALUE to the database..."

if docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<EOF
INSERT INTO configs (key, value)
VALUES ('$KEY', '$VALUE')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    updated_at = NOW();
EOF
then
    echo "Successfully updated $KEY"
else
    echo "Failed to update config"
fi
