#!/bin/sh
set -e

ES_URL="http://elasticsearch:9200"
INDEX_DIR="/indexes"
STATE_DIR="/state"

mkdir -p "$STATE_DIR"

# Update indexes
echo "Creating indexes"

for CONFIG in "$INDEX_DIR"/*.json; do
    [ -f "$CONFIG" ] || continue

    INDEX=$(basename "$CONFIG" .json)
    HASH_FILE="$STATE_DIR/$INDEX.sha256"

    CURRENT_HASH=$(sha256sum "$CONFIG" | awk '{print $1}')

    if [ -f "$HASH_FILE" ] &&
        [ "$(cat "$HASH_FILE")" = "$CURRENT_HASH" ]; then
        echo "Skipping $INDEX (config unchanged)"
        continue
    fi

    echo "Recreating $INDEX (config changed)"

    # Delete existing index. 404 is fine if it doesn't exist.
    curl -s -o /dev/null \
        -X DELETE "$ES_URL/$INDEX"

    # Create index from config.
    curl -sSf -o /dev/null -X PUT "$ES_URL/$INDEX" \
        -H 'Content-Type: application/json' \
        -d @"$CONFIG"

    printf '%s\n' "$CURRENT_HASH" > "$HASH_FILE"

    echo "Created $INDEX"
done

echo "Elasticsearch initialization complete."
