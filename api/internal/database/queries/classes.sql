-- name: ListClassesByID :many
SELECT
    cl.id,
    cl.number,
    cl.note,
    cl.meet_type,

    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'days', cmt.days,
                    'time_start', cmt.time_begin,
                    'time_end', cmt.time_end,
                    'period_start', cmt.period_begin,
                    'period_end', cmt.period_end,
                    'building',
                    CASE
                        WHEN b.id IS NULL THEN NULL
                        ELSE jsonb_build_object(
                            'name', b.name,
                            'code', b.code,
                            'room', NULLIF(cmt.room, '')
                        )
                    END
                )
            )
            FROM class_meet_times cmt
            LEFT JOIN buildings b
                ON b.id = cmt.building_id
            WHERE cmt.class_id = cl.id
        ),
        '[]'::jsonb
    ) AS meet_times,

    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'name', i.name,
                    'rmp_id', i.rmp_id,
                    'rating', i.rating,
                    'difficulty', i.difficulty,
                    'take_again', i.take_again
                )
            )
            FROM class_instructors ci
            JOIN instructors i
                ON i.id = ci.instructor_id
            WHERE ci.class_id = cl.id
        ),
        '[]'::jsonb
    ) AS instructors

FROM unnest(@ids::uuid[]) WITH ORDINALITY AS ids(id, ord)
JOIN classes cl
    ON cl.id = ids.id
ORDER BY ids.ord;