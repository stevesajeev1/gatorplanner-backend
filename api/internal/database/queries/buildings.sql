-- name: ListBuildings :many
SELECT
    name,
    code,
    latitude,
    longitude
FROM buildings
ORDER BY name ASC;