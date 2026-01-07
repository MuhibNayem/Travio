#!/bin/bash
# ==============================================================================
# ScyllaDB Initialization Script
# ==============================================================================
# This script waits for ScyllaDB to be ready and then enforces the schema.
# Usage: ./init-scylla.sh <host> <cql_file>
# ==============================================================================

HOST=$1
CQL_FILE=$2

if [ -z "$HOST" ] || [ -z "$CQL_FILE" ]; then
    echo "Usage: $0 <host> <cql_file>"
    exit 1
fi

echo "Waiting for ScyllaDB at $HOST:9042..."

# Loop until cqlsh can connect
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if cqlsh "$HOST" 9042 -e "DESCRIBE CLUSTER" > /dev/null 2>&1; then
        echo "ScyllaDB is ready!"
        echo "Applying schema from $CQL_FILE..."
        cqlsh "$HOST" 9042 -f "$CQL_FILE"
        if [ $? -eq 0 ]; then
            echo "Schema applied successfully."
            exit 0
        else
            echo "Failed to apply schema."
            exit 1
        fi
    else
        echo "ScyllaDB not ready yet... retrying ($RETRY_COUNT/$MAX_RETRIES)"
        sleep 5
        RETRY_COUNT=$((RETRY_COUNT+1))
    fi
done

echo "Timeout waiting for ScyllaDB."
exit 1
