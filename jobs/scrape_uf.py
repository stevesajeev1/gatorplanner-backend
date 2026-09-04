import argparse
import logging
import os
import re
import threading
from datetime import datetime, time
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
        action="store_true",
        help="Makes authenticated requests, requires ONEUF_SESSION to be set. Defaults to False.",
    )

    parser.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL. Defaults to DATABASE_URL.",
    )

    return parser.parse_args()


def get_term_id(term: str, year: int) -> int:
    return int(f"2{year:02d}{TERM_CODES[term]}")


def get_term(term_id: int) -> str | None:
    term_code = term_id % 10
    for term, code in TERM_CODES.items():
        if term_code == int(code):
            return term
    return None


def fetch_courses(term_id: int, authenticated: bool) -> list[dict]:
    logger.info(f"Fetching UF courses for term {term_id}")

    all_courses: list[dict] = []
    lock = threading.Lock()

    def fetch_thread(thread_id: int, control_number_start: int) -> None:
        control_number = control_number_start

        while True:
            logger.info(f"Thread-{thread_id}: fetching control number {control_number}")

            params = {
                "category": "RES",
                "term": term_id,
                "last-control-number": control_number,
            }

            cookies = None
            if authenticated:
                cookies = {"ONEUF_SESSION": os.environ["ONEUF_SESSION"]}

            try:
                response = requests.get(
                    API_URL, params=params, cookies=cookies, timeout=30
                )
                response.raise_for_status()
                payload = response.json()
            except requests.RequestException as e:
                logger.error(f"Thread-{thread_id}: request failed: {e}")
                raise
            except ValueError as e:
                logger.error(f"Thread-{thread_id}: invalid JSON response: {e}")
                raise

            if not isinstance(payload, list):
                raise TypeError(
                    f"Thread-{thread_id}: unexpected response type: "
                    f"{type(payload).__name__}"
                )

            retrieved_rows = 0

            for page in payload:
                if not isinstance(page, dict):
                    continue

                page_courses = page["COURSES"]

                with lock:
                    all_courses.extend(page_courses)

                value = page["RETRIEVEDROWS"]

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

    logger.info(f"Received {len(all_courses)} courses from UF")

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


def parse_meet_type(value: str) -> str:
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


def parse_quest(value: list[str]) -> str:
    if len(value) > 1:
        raise ValueError(f"Expected at most one quest value, got: {value!r}")

    return str(value[0]).strip().title()


def parse_words(section: dict) -> int:
    value = section["grWriting"]

    if value == "N":
        return 0

    try:
        return int(value) * 1000
    except ValueError as exc:
        raise ValueError(f"Invalid grWriting value: {value!r}") from exc


def parse_times_periods(meet_time: dict[str, Any], term_id: int) -> tuple[time, time, str, str] | tuple[None, None, None, None]:
    time_begin = parse_time(meet_time["meetTimeBegin"])
    time_end = parse_time(meet_time["meetTimeEnd"])
    period_begin = meet_time["meetPeriodBegin"]
    period_end = meet_time["meetPeriodEnd"]

    if time_begin and time_end and period_begin and period_end:
        return (time_begin, time_end, period_begin, period_end)

    if time_begin and time_end:
        term = get_term(term_id)
        if not term:
            return None, None, None, None

        is_summer = term == "summer"
        period_begin = parse_period(time_begin, is_summer)
        period_end = parse_period(time_end, is_summer)

        return (time_begin, time_end, period_begin, period_end)

    return (None, None, None, None)


def parse_time(value: str) -> time:
    return datetime.strptime(value, "%I:%M %p").time()  # noqa: DTZ007


FALL_SPRING_PERIOD_STARTS = {
    time(7, 25): "1",
    time(8, 30): "2",
    time(9, 35): "3",
    time(10, 40): "4",
    time(11, 45): "5",
    time(12, 50): "6",
    time(13, 55): "7",
    time(15, 0): "8",
    time(16, 5): "9",
    time(17, 10): "10",
    time(18, 15): "11",
    time(19, 20): "E1",
    time(20, 20): "E2",
    time(21, 20): "E3",
}

SUMMER_PERIOD_STARTS = {
    time(8, 0): "1",
    time(9, 30): "2",
    time(11, 0): "3",
    time(12, 30): "4",
    time(14, 0): "5",
    time(15, 30): "6",
    time(17, 0): "7",
    time(19, 0): "E1",
    time(20, 30): "E2",
}

def parse_period(time_: time, summer: bool) -> str:
    PERIOD_STARTS = SUMMER_PERIOD_STARTS if summer else FALL_SPRING_PERIOD_STARTS

    for start, period in reversed(PERIOD_STARTS.items()):
        if time_ >= start:
            return period
    raise ValueError(f"Time {time_} is before the first period")


def upsert_building(conn: psycopg.Connection, meet_time: dict[str, Any]) -> tuple[str, str | None] | tuple[None, None]:
    name = meet_time["meetBuilding"].strip()
    code = meet_time["meetBldgCode"].strip()
    if not name and not code:
        return None, None

    row = conn.execute(
        """
        INSERT INTO buildings (name, code)
        VALUES (%s, %s)
        ON CONFLICT (code)
        DO UPDATE SET name = EXCLUDED.name
        WHERE EXCLUDED.name <> ''
        RETURNING id
        """,
        (name, code),
    ).fetchone()

    if row is None:
        row = conn.execute(
            """
            SELECT id
            FROM buildings
            WHERE code = %s
            """,
            (code,),
        ).fetchone()

    assert row is not None

    building_id = row[0]

    room = meet_time["meetRoom"].strip()
    return (building_id, room or None)


def clean_text(value: str) -> str:
    # Remove zero-width and other invisible formatting characters
    value = re.sub(r"[\u200B-\u200D\uFEFF]", "", value)

    # Normalize non-breaking hyphen to regular hyphen
    value = value.replace("\u2011", "-")

    return value


def upsert_course(
    authenticated: bool,
    conn: psycopg.Connection,
    course: dict[str, Any],
    term_id: int,
    ingestion_run_id: str,
) -> None:
    code = course["code"]
    sections = course["sections"]

    if not sections:
        logger.warning(
            "Course %s has no sections; skipping",
            code,
        )
        return

    first_section = sections[0]

    dept_code = first_section.get("deptCode")

    if not dept_code:
        logger.warning(
            "Course %s has no deptCode; skipping",
            code,
        )
        return

    try:
        dept_code = int(dept_code)
    except (TypeError, ValueError):
        logger.warning(
            "Course %s has invalid deptCode %r; skipping",
            code,
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

    gen_eds = list({value for section in sections for value in section["genEd"]})

    quest_values = [
        parse_quest(section["quest"]) for section in sections if section["quest"]
    ]

    quest = quest_values[0] if quest_values else None

    is_ai = first_section.get("isAICourse", False)

    is_honors = first_section.get("isHonorsClass", False)

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
            clean_text(course["description"]),
            first_section["simpleSyllabusParams"],
            course["prerequisites"],
            credits_min,
            credits_max,
            department_id,
            parse_words(first_section),
            gen_eds,
            quest,
            is_ai,
            is_honors,
            ingestion_run_id,
        ),
    )

    for section in sections:
        upsert_class(
            authenticated,
            conn,
            course,
            section,
            term_id,
            ingestion_run_id,
        )


def upsert_class(
    authenticated: bool,
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
            clean_text(section["note"]),
            meet_type,
            ingestion_run_id,
        ),
    ).fetchone()

    assert row is not None

    class_id = row[0]

    if authenticated:
        update_class_meet_times(conn, section, class_id, term_id)

    update_class_instructors(conn, section, class_id)


def update_class_meet_times(
    conn: psycopg.Connection,
    section: dict[str, Any],
    class_id: str,
    term_id: int
) -> None:
    conn.execute(
        """
        DELETE FROM class_meet_times
        WHERE class_id = %s
        """,
        (class_id,),
    )

    meet_times = section["meetTimes"]

    for meet_time in meet_times:
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
                building_id,
                room
            )
            VALUES (
                %s, %s, %s, %s, %s, %s, %s, %s, %s
            )
            """,
            (
                class_id,
                int(meet_time["meetNo"]),
                meet_time["meetDays"],
                *parse_times_periods(meet_time, term_id),
                *upsert_building(conn, meet_time)
            )
        )


def update_class_instructors(
    conn: psycopg.Connection, section: dict[str, Any], class_id: str
) -> None:
    # Get existing instructors for class
    rows = conn.execute(
        """
        SELECT i.id, i.name
        FROM instructors i
        JOIN class_instructors ci
            ON ci.instructor_id = i.id
        WHERE ci.class_id = %s;
        """,
        (class_id,),
    ).fetchall()

    existing_instructor_names = {row[1] for row in rows}

    # Determine instructors that were added/removed
    instructor_names = {value["name"].title() for value in section["instructors"]}

    new_instructor_names = []
    for instructor_name in instructor_names:
        if instructor_name not in existing_instructor_names:
            new_instructor_names.append(instructor_name)

    removed_instructor_ids = []
    for instructor_id, instructor_name in rows:
        if instructor_name not in instructor_names:
            removed_instructor_ids.append(instructor_id)

    # Delete removed instructors
    conn.execute(
        """
        DELETE FROM class_instructors
        WHERE class_id = %s
        AND instructor_id = ANY(%s)
        """,
        (class_id, removed_instructor_ids),
    )

    # Create new instructors
    if new_instructor_names:
        # Always create a new instructor (even if an instructor with the same name exists)
        # This is because two instructors can have the same name but teach separate classes.
        # The "merging" of instructors will be handled in a later job.
        rows = conn.execute(
            """
            INSERT INTO instructors (name)
            SELECT name
            FROM unnest(%s::text[]) AS name
            RETURNING id
            """,
            (new_instructor_names,),
        ).fetchall()

        conn.execute(
            """
            INSERT INTO class_instructors (class_id, instructor_id)
            SELECT %s, id
            FROM unnest(%s::uuid[]) AS id
            """,
            (class_id, [row[0] for row in rows]),
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

    with psycopg.connect(database_url) as conn, conn.transaction():
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

        for index, course in enumerate(courses, start=1):
            logger.info(
                "[%d/%d] Processing %s",
                index,
                len(courses),
                course["code"],
            )

            upsert_course(
                authenticated,
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
            "ONEUF_SESSION environment variable must be set with --authenticated"
        )

    ingest(
        authenticated=args.authenticated,
        database_url=args.database_url,
        term=args.term,
        year=args.year,
    )


if __name__ == "__main__":
    main()
