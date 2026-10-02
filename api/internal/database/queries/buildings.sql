-- name: ListBuildings :many
SELECT
    id,
    name,
    code,
    latitude,
    longitude
FROM buildings
ORDER BY name ASC;