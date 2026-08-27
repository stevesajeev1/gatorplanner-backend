import argparse
import logging
import os
from datetime import datetime
import threading
from typing import Any

import psycopg
import requests
from dotenv import load_dotenv


load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)

logger = logging.getLogger(__name__)

API_URL = "https://one.ufl.edu/apix/soc/schedule/"


TERM_CODES = {
    "spring": "1",
    "summer": "5",
    "fall": "8",
}

FETCH_NUM_THREADS = 16
FETCH_BATCH_SIZE = 50


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Ingest UF course data into PostgreSQL."
    )

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
        "--authenticated",
        type=bool,
        default=False,
        help="Makes authenticated requests, requires ONEUF_SESSION to be set. Defaults to False."
    )

    parser.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL. Defaults to DATABASE_URL.",
    )

    return parser.parse_args()


def get_term_id(term: str, year: int) -> int:
    return int(f"2{year:02d}{TERM_CODES[term]}")


def fetch_courses(term_id: int, authenticated: bool) -> list[dict]:
    logging.info(f"Fetching UF courses for term {term_id}")

    all_courses: list[dict] = []
    lock = threading.Lock()

    def fetch_thread(thread_id: int, control_number_start: int) -> None:
        control_number = control_number_start

        while True:
            logging.info(
                f"Thread-{thread_id}: fetching control number "
                f"{control_number}"
            )

            params = {
                "category": "RES",
                "term": term_id,
                "last-control-number": control_number
            }

            # TODO: if authenticated, pass session cookie
            if authenticated:
                _ = os.environ["ONEUF_SESSION"]

            try:
                response = requests.get(API_URL, params=params, timeout=30)
                response.raise_for_status()
                payload = response.json()
            except requests.RequestException as e:
                logging.error(
                    f"Thread-{thread_id}: request failed: {e}"
                )
                raise
            except ValueError as e:
                logging.error(
                    f"Thread-{thread_id}: invalid JSON response: {e}"
                )
                raise

            if not isinstance(payload, list):
                raise ValueError(
                    f"Thread-{thread_id}: unexpected response type: "
                    f"{type(payload).__name__}"
                )

            retrieved_rows = 0

            for page in payload:
                if not isinstance(page, dict):
                    continue

                page_courses = page.get("COURSES", [])

                if not isinstance(page_courses, list):
                    raise ValueError(
                        f"Thread-{thread_id}: COURSES is not a list"
                    )

                with lock:
                    all_courses.extend(page_courses)

                value = page.get("RETRIEVEDROWS", 0)

                try:
                    retrieved_rows += int(value or 0)
                except (TypeError, ValueError):
                    pass

            # No more results for this control number.
            if retrieved_rows == 0:
                break

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

    logging.info(f"Received {len(all_courses)} courses from UF")

    return all_courses


def get_or_create_department(
    conn: psycopg.Connection,
    code: int,
    name: str,
) -> str:
    row = conn.execute(
        """
        INSERT INTO departments (code, name)
        VALUES (%s, %s)
        ON CONFLICT (code)
        DO UPDATE SET name = EXCLUDED.name
        RETURNING id
        """,
        (code, name),
    ).fetchone()

    assert row is not None

    return row[0]


def get_or_create_instructor(
    conn: psycopg.Connection,
    name: str,
) -> str:
    """
    Find an instructor by name.

    Names are intentionally NOT unique. If multiple instructors have
    the same name, this function does not attempt to distinguish them.
    That distinction can be handled later by the RMP ingestion job
    using the classes they teach.
    """

    rows = conn.execute(
        """
        SELECT id
        FROM instructors
        WHERE name = %s
        ORDER BY id
        """,
        (name,),
    ).fetchall()

    if len(rows) == 1:
        return rows[0][0]

    if len(rows) > 1:
        logger.warning(
            "Multiple instructors already exist with name %r; "
            "using the first existing instructor",
            name,
        )
        return rows[0][0]

    row = conn.execute(
        """
        INSERT INTO instructors (name)
        VALUES (%s)
        RETURNING id
        """,
        (name,),
    ).fetchone()

    assert row is not None

    return row[0]


def parse_meet_type(value: str) -> str:
    """
    Convert UF's sectWeb value into the class_meet_type enum.
    """

    mapping = {
        "PC": "Primarily Classroom",
        "HB": "Hybrid",
        "PD": "Online (80-99%)",
        "AD": "Online (100%)",
    }

    try:
        return mapping[value]
    except KeyError:
        raise ValueError(f"Unknown UF sectWeb value: {value!r}")


def parse_quest(value) -> str | None:
    if not value:
        return None

    if isinstance(value, list):
        if not value:
            return None

        if len(value) > 1:
            raise ValueError(
                f"Expected at most one quest value, got: {value!r}"
            )

        value = value[0]

    value = str(value).strip()

    if not value:
        return None

    return value.title()


def parse_words(section: dict) -> int:
    value = section.get("grWriting")

    if value is None:
        return 0

    value = str(value).strip()

    if not value or value.upper() == "N":
        return 0

    try:
        return int(value) * 1000
    except ValueError as exc:
        raise ValueError(
            f"Invalid grWriting value: {value!r}"
        ) from exc


def parse_building(meet_time: dict[str, Any]) -> str:
    """
    building is:

        meetBuilding + meetRoom

    if both are present.

    If both are empty, use meetBldgCode.
    """

    building = (meet_time.get("meetBuilding") or "").strip()
    room = (meet_time.get("meetRoom") or "").strip()

    if building and room:
        return f"{building} {room}"

    if building:
        return building

    if room:
        return room

    return (meet_time.get("meetBldgCode") or "").strip()


def parse_time(value: str):
    return datetime.strptime(value, "%I:%M %p").time()


# TODO: reorganize this to perform validation upfront, then only do anything
# this might be unnecessary since we do everything in a transaction
# TODO: we should collect skipped classes somewhere
def upsert_course(
    conn: psycopg.Connection,
    course: dict[str, Any],
    term_id: int,
    ingestion_run_id: str,
) -> None:
    sections = course.get("sections", [])

    if not sections:
        logger.warning(
            "Course %s has no sections; skipping",
            course.get("code"),
        )
        return

    first_section = sections[0]

    dept_code = first_section.get("deptCode")

    if not dept_code:
        logging.warning(
            "Course %s has no deptCode; skipping",
            course.get("code"),
        )
        return

    try:
        dept_code = int(dept_code)
    except (TypeError, ValueError):
        logging.warning(
            "Course %s has invalid deptCode %r; skipping",
            course.get("code"),
            dept_code,
        )
        return

    department_id = get_or_create_department(
        conn,
        dept_code,
        first_section["deptName"],
    )

    credits_min = float(first_section["credits_min"])

    credits_max = float(first_section["credits_max"])

    quest_values = [
        parse_quest(section.get("quest"))
        for section in sections
        if section.get("quest")
    ]

    quest = quest_values[0] if quest_values else None

    is_ai = any(
        bool(section.get("isAICourse", False))
        for section in sections
    )

    is_honors = any(
        bool(section.get("isHonorsClass", False))
        for section in sections
    )

    conn.execute(
        """
        INSERT INTO courses (
            id,
            term_id,
            code,
            name,
            description,
            syllabus,
            prerequisites,
            credits_min,
            credits_max,
            department_id,
            words,
            gen_eds,
            quest,
            is_ai,
            is_honors,
            ingestion_run_id
        )
        VALUES (
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s
        )
        ON CONFLICT (id, term_id)
        DO UPDATE SET
            code = EXCLUDED.code,
            name = EXCLUDED.name,
            description = EXCLUDED.description,
            syllabus = EXCLUDED.syllabus,
            prerequisites = EXCLUDED.prerequisites,
            credits_min = EXCLUDED.credits_min,
            credits_max = EXCLUDED.credits_max,
            department_id = EXCLUDED.department_id,
            words = EXCLUDED.words,
            gen_eds = EXCLUDED.gen_eds,
            quest = EXCLUDED.quest,
            is_ai = EXCLUDED.is_ai,
            is_honors = EXCLUDED.is_honors,
            ingestion_run_id = EXCLUDED.ingestion_run_id
        """,
        (
            int(course["courseId"]),
            term_id,
            course["code"],
            course["name"],
            course.get("description", ""),
            first_section.get("simpleSyllabusParams", ""),
            course.get("prerequisites", ""),
            credits_min,
            credits_max,
            department_id,
            parse_words(first_section),
            quest_values and list(
                {
                    value
                    for section in sections
                    for value in section.get("genEd", [])
                }
            ) or [],
            quest,
            is_ai,
            is_honors,
            ingestion_run_id,
        ),
    )

    for section in sections:
        upsert_class(
            conn,
            course,
            section,
            term_id,
            ingestion_run_id,
        )


def upsert_class(
    conn: psycopg.Connection,
    course: dict[str, Any],
    section: dict[str, Any],
    term_id: int,
    ingestion_run_id: str,
) -> None:
    course_id = int(course["courseId"])
    class_number = int(section["classNumber"])

    meet_type = parse_meet_type(section["sectWeb"])

    row = conn.execute(
        """
        INSERT INTO classes (
            course_id,
            term_id,
            number,
            note,
            meet_type,
            ingestion_run_id
        )
        VALUES (
            %s,
            %s,
            %s,
            %s,
            %s,
            %s
        )
        ON CONFLICT (course_id, term_id, number)
        DO UPDATE SET
            note = EXCLUDED.note,
            meet_type = EXCLUDED.meet_type,
            ingestion_run_id = EXCLUDED.ingestion_run_id
        RETURNING id
        """,
        (
            course_id,
            term_id,
            class_number,
            section.get("note") or None,
            meet_type,
            ingestion_run_id,
        ),
    ).fetchone()

    assert row is not None

    class_id = row[0]

    # Replace meeting times for this class.
    conn.execute(
        """
        DELETE FROM class_meet_times
        WHERE class_id = %s
        """,
        (class_id,),
    )

    for meet_time in section.get("meetTimes", []):
        insert_meet_time(
            conn,
            class_id,
            meet_time,
        )

    # Replace instructor relationships for this class.
    conn.execute(
        """
        DELETE FROM class_instructors
        WHERE class_id = %s
        """,
        (class_id,),
    )

    for instructor in section.get("instructors", []):
        name = (instructor.get("name") or "").strip()

        if not name:
            continue

        instructor_id = get_or_create_instructor(
            conn,
            name,
        )

        conn.execute(
            """
            INSERT INTO class_instructors (
                class_id,
                instructor_id
            )
            VALUES (%s, %s)
            ON CONFLICT DO NOTHING
            """,
            (class_id, instructor_id),
        )


def insert_meet_time(
    conn: psycopg.Connection,
    class_id: str,
    meet_time: dict[str, Any],
) -> None:
    conn.execute(
        """
        INSERT INTO class_meet_times (
            class_id,
            number,
            days,
            time_begin,
            time_end,
            period_begin,
            period_end,
            building
        )
        VALUES (
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s,
            %s
        )
        """,
        (
            class_id,
            int(meet_time["meetNo"]),
            meet_time.get("meetDays", []),
            parse_time(meet_time["meetTimeBegin"]),
            parse_time(meet_time["meetTimeEnd"]),
            int(meet_time["meetPeriodBegin"]),
            int(meet_time["meetPeriodEnd"]),
            parse_building(meet_time),
        ),
    )


def ingest(
    authenticated: bool,
    database_url: str,
    term: str,
    year: int,
) -> None:
    term_id = get_term_id(term, year)

    logger.info(
        "Starting ingestion for %s %02d (term_id=%d)",
        term,
        year,
        term_id,
    )

    courses = fetch_courses(term_id, authenticated)

    with psycopg.connect(database_url) as conn:
        with conn.transaction():
            row = conn.execute(
                """
                INSERT INTO ingestion_runs (term_id)
                VALUES (%s)
                RETURNING id
                """,
                (term_id,),
            ).fetchone()

            assert row is not None

            ingestion_run_id = row[0]

            logger.info(
                "Created ingestion run %s",
                ingestion_run_id,
            )

            # TODO: make this threaded
            for index, course in enumerate(courses, start=1):
                logger.info(
                    "[%d/%d] Processing %s",
                    index,
                    len(courses),
                    course.get("code"),
                )

                upsert_course(
                    conn,
                    course,
                    term_id,
                    ingestion_run_id,
                )

            logger.info(
                "Ingestion run %s completed successfully",
                ingestion_run_id,
            )


def main() -> None:
    args = parse_args()

    if not args.database_url:
        raise RuntimeError(
            "DATABASE_URL must be provided with --database-url "
            "or the DATABASE_URL environment variable"
        )

    if not 0 <= args.year <= 99:
        raise ValueError("year must be between 00 and 99")

    if args.authenticated and "ONEUF_SESSION" not in os.environ:
        raise RuntimeError(
            "ONEUF_SESSION environment variable must be set "
            "with --authenticated"
        )

    ingest(
        authenticated = args.authenticated,
        database_url=args.database_url,
        term=args.term,
        year=args.year,
    )


if __name__ == "__main__":
    main()