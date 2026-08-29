import argparse
import logging
import os

import psycopg
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)

logger = logging.getLogger(__name__)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Cleanup ingestion runs.")

    parser.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL. Defaults to DATABASE_URL.",
    )

    return parser.parse_args()


def cleanup(database_url: str) -> None:
    logger.info("Starting cleanup")

    with psycopg.connect(database_url) as conn, conn.transaction():
        # Delete stale courses/classes
        result = conn.execute(
            """
            DELETE FROM courses c
            WHERE c.ingestion_run_id IS NOT NULL
            AND NOT EXISTS (
                SELECT 1
                FROM (
                    SELECT
                        id,
                        ROW_NUMBER() OVER (
                            PARTITION BY term_id
                            ORDER BY created_at DESC
                        ) AS rn
                    FROM ingestion_runs
                ) ir
                WHERE ir.id = c.ingestion_run_id
                AND ir.rn <= 5
            )
            """
        )

        logger.info(f"Cleaned up {result.rowcount} stale courses")

        result = conn.execute(
            """
            DELETE FROM classes cl
            WHERE cl.ingestion_run_id IS NOT NULL
            AND NOT EXISTS (
                SELECT 1
                FROM (
                    SELECT
                        id,
                        ROW_NUMBER() OVER (
                            PARTITION BY term_id
                            ORDER BY created_at DESC
                        ) AS rn
                    FROM ingestion_runs
                ) ir
                WHERE ir.id = cl.ingestion_run_id
                AND ir.rn <= 5
            )
            """
        )

        logger.info(f"Cleaned up {result.rowcount} stale classes")

        # Delete instructors with no classes
        result = conn.execute(
            """
            DELETE FROM instructors i
            WHERE NOT EXISTS (
                SELECT 1
                FROM class_instructors ci
                WHERE ci.instructor_id = i.id
            )
            """
        )

        logger.info(f"Cleaned up {result.rowcount} stale instructors")


def main() -> None:
    args = parse_args()

    if not args.database_url:
        raise RuntimeError(
            "DATABASE_URL must be provided with --database-url "
            "or the DATABASE_URL environment variable"
        )

    cleanup(database_url=args.database_url)


if __name__ == "__main__":
    main()
