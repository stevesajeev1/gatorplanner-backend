-- name: ListCoursesByID :many
SELECT
    c.code,
    c.is_lab,
    c.name,
    c.description,
    c.syllabus,
    c.prerequisites,
    c.credits_min,
    c.credits_max,
    c.words,
    to_jsonb(c.gen_eds) AS gen_eds,
    c.quest,
    c.is_ai,
    c.is_honors,

    d.name AS department

FROM unnest(@ids::integer[]) WITH ORDINALITY AS ids(id, ord)
JOIN courses c
    ON c.id = ids.id
JOIN departments d
    ON d.id = c.department_id
ORDER BY ids.ord;