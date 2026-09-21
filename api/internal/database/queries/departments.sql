-- name: ListDepartments :many
SELECT DISTINCT name FROM departments ORDER BY name ASC;