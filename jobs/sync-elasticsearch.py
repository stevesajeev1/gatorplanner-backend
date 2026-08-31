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


TERM_CODES = {
    "spring": "1",
    "summer": "5",
    "fall": "8",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sync Elasticsearch.")

    parser.add_argument(
        "term",
        choices=TERM_CODES,
        help="Term to ingest: spring, summer, or fall",
    )

    parser.add_argument(
        "year",
        type=int,
        help="Two-digit year, e.g. 26 for 2026",
    )

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


def get_term_id(term: str, year: int) -> int:
    return int(f"2{year:02d}{TERM_CODES[term]}")


def sync(
    database_url: str,
    es_client: Elasticsearch,
    term: str,
    year: int,
) -> None:
    term_id = get_term_id(term, year)

    logger.info("Starting Elasticsearch sync")

    es_ids = {
        doc["_id"]
        for doc in scan(
            es_client,
            index="classes",
            query={"query": {"term": {"term_id": term_id}}},
        )
    }

    actions = []
    db_ids = set()
    with psycopg.connect(database_url) as conn:
        rows = conn.execute(
            """
            SELECT
                cl.id,
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

            WHERE cl.term_id = %s
            """,
            (term_id,),
        ).fetchall()

        for row in rows:
            class_id = str(row[0])
            document = {
                "term_id": term_id,
                "number": row[1],
                "note": row[2],
                "meet_type": row[3],
                "course_code": row[4],
                "course_code_prefix": row[5],
                "course_level": row[6],
                "course_is_lab": row[7],
                "course_name": row[8],
                "course_description": row[9],
                "course_prerequisites": row[10],
                "course_credits": row[11],
                "course_department": row[12],
                "course_words": row[13],
                "course_gen_eds": row[14],
                "course_quest": row[15],
                "course_is_ai": row[16],
                "course_is_honors": row[17],
                "meet_times": row[18],
                "instructors": row[19],
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

    if not 0 <= args.year <= 99:
        raise ValueError("year must be between 00 and 99")

    es_client = Elasticsearch(args.elasticsearch_url)

    sync(
        database_url=args.database_url,
        es_client=es_client,
        term=args.term,
        year=args.year,
    )


if __name__ == "__main__":
    main()
