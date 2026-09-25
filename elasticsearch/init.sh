#!/bin/sh
set -e

ES_URL="http://localhost:9200"
INDEX_DIR="./indexes"

# Update indexes
echo "Creating indexes"

for CONFIG in "$INDEX_DIR"/*.json; do
    [ -f "$CONFIG" ] || continue

    INDEX=$(basename "$CONFIG" .json)

    # Delete existing index. 404 is fine if it doesn't exist.
    curl -s -o /dev/null \
        -X DELETE "$ES_URL/$INDEX"

    # Create index from config.
    curl -sSf -o /dev/null -X PUT "$ES_URL/$INDEX" \
        -H 'Content-Type: application/json' \
        -d @"$CONFIG"

    echo "Created $INDEX"
done

echo "Elasticsearch initialization complete."
