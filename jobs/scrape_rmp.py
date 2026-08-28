import argparse
import logging
import os
import threading

import psycopg
from dotenv import load_dotenv
import requests

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)

logger = logging.getLogger(__name__)

API_URL = "https://www.ratemyprofessors.com/graphql"


FETCH_NUM_THREADS = 16
FETCH_BATCH_SIZE = 50


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Ingest RMP data into PostgreSQL."
    )

    parser.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL. Defaults to DATABASE_URL.",
    )

    parser.add_argument(
        "--school-id",
        default=os.environ.get("RMP_SCHOOL_ID"),
        help="RMP School ID. Defaults to RMP_SCHOOL_ID.",
    )

    return parser.parse_args()


def fetch_instructors(
    database_url: str,
    school_id: str
) -> list[dict]:
    logger.info("Fetching instructors from RMP")

    with psycopg.connect(database_url) as conn:
        rows = conn.execute(
            """
            SELECT DISTINCT name FROM instructors;
            """
        ).fetchall()

    instructors_count = len(rows)

    logger.info(f"Fetching information for {instructors_count} instructors")

    all_instructors: list[dict] = []
    lock = threading.Lock()

    def fetch_thread(thread_id: int, control_number_start: int) -> None:
        control_number = control_number_start

        while control_number < instructors_count:
            logger.info(f"Thread-{thread_id}: fetching control numbers {control_number}-{control_number + FETCH_BATCH_SIZE}")

            for i in range(FETCH_BATCH_SIZE):
                try:
                    query = """
                        query SearchTeachers($schoolID: ID!, $name: String!) {
                            newSearch {
                                teachers(
                                    query: {
                                        schoolID: $schoolID
                                        text: $name
                                    }
                                ) {
                                    edges {
                                        node {
                                            id
                                            legacyId
                                            firstName
                                            lastName
                                            department
                                        }
                                    }
                                }
                            }
                        }
                    """

                    variables = {
                        "schoolID": school_id,
                        "name": rows[control_number + i][0],
                    }

                    response = requests.post(
                        API_URL,
                        json={
                            "query": query,
                            "variables": variables,
                        },
                        timeout=30
                    )
                    response.raise_for_status()
                    payload = response.json()
                except requests.RequestException as e:
                    logger.error(f"Thread-{thread_id}: request failed: {e}")
                    raise
                except ValueError as e:
                    logger.error(f"Thread-{thread_id}: invalid JSON response: {e}")
                    raise

            control_number += FETCH_BATCH_SIZE * FETCH_NUM_THREADS

    threads = []

    for thread_id in range(FETCH_NUM_THREADS):
        thread = threading.Thread(
            target=fetch_thread,
            args=(thread_id, FETCH_BATCH_SIZE * thread_id),
        )
        thread.start()
        threads.append(thread)

    for thread in threads:
        thread.join()

    return []

def ingest(
    database_url: str,
    school_id: str
) -> None:
    logger.info(
        "Starting RMP ingestion",
    )

    instructors = fetch_instructors(database_url, school_id)


def main() -> None:
    args = parse_args()

    if not args.database_url:
        raise RuntimeError(
            "DATABASE_URL must be provided with --database-url "
            "or the DATABASE_URL environment variable"
        )

    if not args.school_id:
            raise RuntimeError(
                "RMP_SCHOOL_ID must be provided with --school-id "
                "or the RMP_SCHOOL_ID environment variable"
            )

    ingest(
        database_url=args.database_url,
        school_id=args.school_id,
    )

if __name__ == "__main__":
    main()