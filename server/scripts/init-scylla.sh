#!/bin/bash
# ==============================================================================
# ScyllaDB Initialization Script
# ==============================================================================
# This script waits for ScyllaDB to be ready and then enforces the schema.
# Usage: ./init-scylla.sh <host> <cql_file>
# ==============================================================================

HOST=$1
TARGET=$2

if [ -z "$HOST" ] || [ -z "$TARGET" ]; then
    echo "Usage: $0 <host> <cql_file_or_directory>"
    exit 1
fi

echo "Waiting for ScyllaDB at $HOST:9042..."

# Loop until cqlsh can connect
MAX_RETRIES=30
RETRY_COUNT=0

wait_for_scylla() {
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if cqlsh "$HOST" 9042 -e "DESCRIBE CLUSTER" > /dev/null 2>&1; then
            echo "ScyllaDB is ready!"
            return 0
        else
            echo "ScyllaDB not ready yet... retrying ($RETRY_COUNT/$MAX_RETRIES)"
            sleep 5
            RETRY_COUNT=$((RETRY_COUNT+1))
        fi
    done
    return 1
}

apply_schema() {
    local file=$1
    echo "Applying schema from $file..."
    cqlsh "$HOST" 9042 -f "$file"
    if [ $? -eq 0 ]; then
        echo "Schema $file applied successfully."
    else
        echo "Failed to apply schema $file."
        exit 1
    fi
}

if wait_for_scylla; then
    if [ -d "$TARGET" ]; then
        echo "Processing schemas from directory $TARGET..."
        for file in "$TARGET"/*.cql; do
            if [ -f "$file" ]; then
                apply_schema "$file"
            fi
        done
    else
        apply_schema "$TARGET"
    fi
    exit 0
else
    echo "Timeout waiting for ScyllaDB."
    exit 1
fi
