import argparse
from collections import defaultdict
from dataclasses import dataclass
import logging
import os
import random
import threading
import time

import curl_cffi
import psycopg
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)

logger = logging.getLogger(__name__)

API_URL = "https://www.ratemyprofessors.com/graphql"

SEARCH_TEACHERS_QUERY = """
    query SearchTeachers($schoolID: ID!, $name: String!) {
        newSearch {
            teachers(query: {schoolID: $schoolID, text: $name}) {
                edges {
                    node {
                        legacyId
                        firstName
                        lastName
                        avgRatingRounded
                        avgDifficultyRounded
                        wouldTakeAgainPercentRounded
                        courseCodes {
                            courseName
                        }
                    }
                }
            }
        }
    }
"""


FETCH_NUM_THREADS = 16


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


@dataclass
class Instructor:
    rmp_id: int
    rating: float
    difficulty: float
    take_again: float
    courses: list[str]


threads_rate_limited = set()
rate_limit_lock = threading.Lock()


def request_with_retries(
    session: curl_cffi.Session,
    url: str,
    *,
    thread_id: int,
    max_retries: int = 5,
    **kwargs,
) -> curl_cffi.Response:
    global threads_rate_limited

    attempt = 0
    while True:
        try:
            # Check if we should globally wait
            is_current_thread_rate_limited = False
            global_delay = 0
            with rate_limit_lock:
                if len(threads_rate_limited) > 0:
                    is_current_thread_rate_limited = thread_id in threads_rate_limited

                    global_delay = min(30 * 2 ** len(threads_rate_limited), 10 * 60)
            if global_delay > 0:
                # jitter
                global_delay = random.uniform(0.5 * global_delay, global_delay)

                logger.warning(
                    f"Thread-{thread_id}: Respecting global wait time of {global_delay:.1f} seconds",
                )

                # Reset attempts
                attempt = 0
            time.sleep(global_delay)

            response = session.post(url, **kwargs)

            if response.status_code != 429:
                if is_current_thread_rate_limited:
                    with rate_limit_lock:
                        threads_rate_limited.remove(thread_id)

                response.raise_for_status()
                return response

            delay = 2**attempt

            # jitter
            delay = random.uniform(0.5 * delay, delay)

            if attempt == max_retries:
                if not is_current_thread_rate_limited:
                    with rate_limit_lock:
                        threads_rate_limited.add(thread_id)
                continue

            logger.warning(
                f"Thread-{thread_id}: Rate limited (429), retrying in {delay:.1f} seconds",
            )
            time.sleep(delay)

        except curl_cffi.requests.exceptions.Timeout:
            pass

        attempt += 1


def fetch_instructors(
    database_url: str,
    school_id: str
) -> dict[str, list[Instructor]]:
    with psycopg.connect(database_url) as conn:
        rows = conn.execute(
            """
            SELECT DISTINCT name FROM instructors;
            """
        ).fetchall()

    instructors_count = len(rows)

    logger.info(f"Fetching information for {instructors_count} instructors")

    all_instructors: dict[str, list[Instructor]] = defaultdict(list[Instructor])
    lock = threading.Lock()

    def fetch_thread(thread_id: int) -> None:
        session = curl_cffi.Session()

        i = thread_id
        while i < instructors_count:
            if i % (FETCH_NUM_THREADS * 10) == thread_id:
                logger.info(f"Thread-{thread_id}: fetching instructor {i}/{instructors_count}")

            instructor_name = rows[i][0]

            try:
                variables = {
                    "schoolID": school_id,
                    "name": instructor_name,
                }

                response = request_with_retries(
                    session,
                    API_URL,
                    json={
                        "query": SEARCH_TEACHERS_QUERY,
                        "variables": variables,
                    },
                    timeout=30,
                    impersonate="chrome",
                    thread_id=thread_id
                )

                payload = response.json()
            except curl_cffi.requests.RequestsError as e:
                logger.error(f"Thread-{thread_id}: request failed: {e}")
                raise
            except ValueError as e:
                logger.error(f"Thread-{thread_id}: invalid JSON response: {e}")
                raise

            for edge in payload["data"]["newSearch"]["teachers"]["edges"]:
                node = edge["node"]

                name = node["firstName"] + " " + node["lastName"]
                if name != instructor_name:
                    continue

                instructor = Instructor(
                    rmp_id=node["legacyId"],
                    rating=node["avgRatingRounded"],
                    difficulty=node["avgDifficultyRounded"],
                    take_again=node["wouldTakeAgainPercentRounded"],
                    courses=[course["courseName"].upper() for course in node["courseCodes"]]
                )

                with lock:
                    all_instructors[instructor_name].append(instructor)

            i += FETCH_NUM_THREADS

    threads = []

    for thread_id in range(FETCH_NUM_THREADS):
        thread = threading.Thread(
            target=fetch_thread,
            args=(thread_id,),
        )
        thread.start()
        threads.append(thread)

    for thread in threads:
        thread.join()

    return all_instructors


def update_instructor(
    conn: psycopg.Connection,
    instructor_name: str,
    instructors: list[Instructor],
) -> None:
    for instructor in instructors:
        row = conn.execute(
            """
            UPDATE instructors i
            SET
                rmp_id = %s,
                rating = NULLIF(%s, -1),
                difficulty = NULLIF(%s, -1),
                take_again = NULLIF(%s, -1)
            WHERE i.id = (
                SELECT i2.id
                FROM instructors i2
                WHERE i2.name = %s
                    AND EXISTS (
                        SELECT 1
                        FROM class_instructors ci
                        JOIN classes cl
                            ON cl.id = ci.class_id
                        JOIN courses c
                            ON c.id = cl.course_id
                        WHERE ci.instructor_id = i2.id
                            AND c.code = ANY(%s)
                    )
                ORDER BY i2.created_at DESC
                LIMIT 1
            )
            RETURNING i.id;
            """,
            (
                instructor.rmp_id,
                instructor.rating,
                instructor.difficulty,
                instructor.take_again,
                instructor_name,
                instructor.courses,
            ),
        ).fetchone()

        if row is None:
            continue

        first_instructor_id = row[0]

        conn.execute(
            """
            UPDATE class_instructors ci
            SET instructor_id = %s
            WHERE instructor_id IN (
                SELECT i.id
                FROM instructors i
                WHERE i.name = %s
                    AND i.id <> %s
                    AND EXISTS (
                        SELECT 1
                        FROM class_instructors ci2
                        JOIN classes cl
                            ON cl.id = ci2.class_id
                        JOIN courses c
                            ON c.id = cl.course_id
                        WHERE ci2.instructor_id = i.id
                            AND c.code = ANY(%s)
                    )
            )
            """,
            (
                first_instructor_id,
                instructor_name,
                first_instructor_id,
                instructor.courses,
            ),
        )


def ingest(
    database_url: str,
    school_id: str
) -> None:
    logger.info(
        "Starting RMP ingestion",
    )

    instructors = fetch_instructors(database_url, school_id)

    with psycopg.connect(database_url) as conn, conn.transaction():
        for index, (name, instructor) in enumerate(instructors.items(), start=1):
            logger.info(
                "[%d/%d] Processing %s",
                index,
                len(instructors),
                name,
            )

            update_instructor(
                conn,
                name,
                instructor
            )


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