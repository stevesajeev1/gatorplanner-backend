import argparse
import logging
import os

import psycopg
from dotenv import load_dotenv
from elasticsearch import Elasticsearch
from elasticsearch.helpers import bulk, scan

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)
logging.getLogger("elastic_transport").setLevel(logging.WARNING)

logger = logging.getLogger(__name__)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sync Elasticsearch.")

    parser.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL. Defaults to DATABASE_URL.",
    )

    parser.add_argument(
        "--elasticsearch-url",
        default=os.environ.get("ELASTICSEARCH_URL"),
        help="Elasticsearch URL. Defaults to ELASTICSEARCH_URL.",
    )

    return parser.parse_args()


def sync(database_url: str, es_client: Elasticsearch) -> None:
    logger.info("Starting Elasticsearch sync")

    es_ids = {doc["_id"] for doc in scan(es_client, index="classes")}

    actions = []
    db_ids = set()
    with psycopg.connect(database_url) as conn:
        rows = conn.execute(
            """
            SELECT
                cl.id,
                cl.term_id,
                cl.number,
                cl.note,
                cl.meet_type,

                c.code AS course_code,
                c.code_prefix AS course_code_prefix,
                c.level AS course_level,
                c.is_lab AS course_is_lab,
                c.name AS course_name,
                c.description AS course_description,
                c.prerequisites AS course_prerequisites,

                jsonb_build_object(
                    'gte', c.credits_min,
                    'lte', c.credits_max
                ) AS course_credits,

                d.name AS course_department,
                c.words AS course_words,
                c.gen_eds AS course_gen_eds,
                c.quest AS course_quest,
                c.is_ai AS course_is_ai,
                c.is_honors AS course_is_honors,

                COALESCE(
                    (
                        SELECT jsonb_agg(
                            jsonb_build_object(
                                'days', cmt.days,
                                'time', jsonb_build_object(
                                    'gte',
                                        EXTRACT(HOUR FROM cmt.time_begin)::integer * 60
                                        + EXTRACT(MINUTE FROM cmt.time_begin)::integer,
                                    'lte',
                                        EXTRACT(HOUR FROM cmt.time_end)::integer * 60
                                        + EXTRACT(MINUTE FROM cmt.time_end)::integer
                                ),
                                'period', jsonb_build_object(
                                    'gte',
                                    CASE
                                        WHEN cmt.period_begin LIKE 'E%'
                                            THEN 11 + substring(
                                                cmt.period_begin FROM 2
                                            )::integer
                                        ELSE NULLIF(cmt.period_begin, '')::integer
                                    END,
                                    'lte',
                                    CASE
                                        WHEN cmt.period_end LIKE 'E%'
                                            THEN 11 + substring(
                                                cmt.period_end FROM 2
                                            )::integer
                                        ELSE NULLIF(cmt.period_end, '')::integer
                                    END
                                ),
                                'building', cmt.building
                            )
                        )
                        FROM class_meet_times cmt
                        WHERE cmt.class_id = cl.id
                    ),
                    '[]'::jsonb
                ) AS meet_times,

                COALESCE(
                    (
                        SELECT jsonb_agg(
                            jsonb_build_object(
                                'name', i.name,
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

            FROM classes cl

            JOIN courses c
                ON c.id = cl.course_id
                AND c.term_id = cl.term_id

            JOIN departments d
                ON d.id = c.department_id
        """
        ).fetchall()

        for row in rows:
            class_id = str(row[0])
            document = {
                "term_id": row[1],
                "number": row[2],
                "note": row[3],
                "meet_type": row[4],
                "course_code": row[5],
                "course_code_prefix": row[6],
                "course_level": row[7],
                "course_is_lab": row[8],
                "course_name": row[9],
                "course_description": row[10],
                "course_prerequisites": row[11],
                "course_credits": row[12],
                "course_department": row[13],
                "course_words": row[14],
                "course_gen_eds": row[15],
                "course_quest": row[16],
                "course_is_ai": row[17],
                "course_is_honors": row[18],
                "meet_times": row[19],
                "instructors": row[20],
            }

            actions.append(
                {
                    "_index": "classes",
                    "_id": class_id,
                    "_source": document,
                }
            )
            db_ids.add(class_id)

    deleted_ids = es_ids - db_ids
    for class_id in deleted_ids:
        actions.append(
            {
                "_op_type": "delete",
                "_index": "classes",
                "_id": class_id,
            }
        )

    success, failed = bulk(
        es_client,
        actions,
        stats_only=True,
    )

    logger.info(f"Updated {success} documents with {failed} failures")


def main() -> None:
    args = parse_args()

    if not args.database_url:
        raise RuntimeError(
            "DATABASE_URL must be provided with --database-url "
            "or the DATABASE_URL environment variable"
        )

    if not args.elasticsearch_url:
        raise RuntimeError(
            "ELASTICSEARCH_URL must be provided with --elasticsearch-url "
            "or the ELASTICSEARCH_URL environment variable"
        )

    es_client = Elasticsearch(args.elasticsearch_url)

    sync(database_url=args.database_url, es_client=es_client)


if __name__ == "__main__":
    main()
