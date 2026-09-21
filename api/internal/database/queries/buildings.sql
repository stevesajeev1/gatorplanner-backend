-- name: ListBuildings :many
SELECT DISTINCT name FROM buildings ORDER BY name ASC;