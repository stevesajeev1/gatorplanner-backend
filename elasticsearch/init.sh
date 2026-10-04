#!/bin/sh
set -e

INDEX_DIR="./indexes"

echo "Waiting for Elasticsearch..."

until curl -sSf "${ELASTICSEARCH_URL}/_cluster/health" > /dev/null; do
    sleep 2
done

# Update indexes
echo "Creating indexes"

for CONFIG in "$INDEX_DIR"/*.json; do
    [ -f "$CONFIG" ] || continue

    INDEX=$(basename "$CONFIG" .json)

    # Delete existing index. 404 is fine if it doesn't exist.
    curl -s -o /dev/null \
        -X DELETE "${ELASTICSEARCH_URL}/$INDEX"

    # Create index from config.
    curl -sSf -o /dev/null -X PUT "${ELASTICSEARCH_URL}/$INDEX" \
        -H 'Content-Type: application/json' \
        -d @"$CONFIG"

    echo "Created $INDEX"
done

echo "Elasticsearch initialization complete."
