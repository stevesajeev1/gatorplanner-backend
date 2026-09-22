import argparse
import logging
import os
from dataclasses import dataclass
import uuid

import psycopg
import requests
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)

logger = logging.getLogger(__name__)

DATA_URL = "https://campusmap.ufl.edu/library/api/searchBldg"

overrides = {
    'FLAV': '1999',
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Ingest building data into PostgreSQL.")

    parser.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL. Defaults to DATABASE_URL."
    )

    return parser.parse_args()


@dataclass
class Building:
    id: uuid.UUID
    code: str
    latitude: float
    longitude: float


def fetch_buildings(database_url: str) -> list[Building]:
    with psycopg.connect(database_url) as conn:
        rows = conn.execute(
            """
            SELECT id, code FROM buildings;
            """
        ).fetchall()

    buildings_count = len(rows)

    logger.info(f"Fetching information for {buildings_count} buildings")

    all_buildings = []

    response = requests.get(DATA_URL)
    response.raise_for_status()
    data = response.json()

    map: dict[str, tuple[float, float]] = {}
    for building in data:
        code = building["BLDG"]
        lat = float(building["LAT"])
        lon = float(building["LON"])

        map[code] = (float(lat), float(lon))

    for row in rows:
        id = row[0]
        code = overrides.get(row[1], row[1])

        if code not in map:
            continue
        lat, lon = map[code]

        all_buildings.append(Building(id, code, lat, lon))
    return all_buildings


def ingest(database_url: str) -> None:
    logger.info("Starting building ingestion")

    buildings = fetch_buildings(database_url)

    with psycopg.connect(database_url) as conn:
        conn.execute(
            """
            UPDATE buildings AS b
            SET
                code = v.code,
                latitude = v.latitude,
                longitude = v.longitude
            FROM unnest(
                %s::uuid[],
                %s::text[],
                %s::numeric[],
                %s::numeric[]
            ) AS v(id, code, latitude, longitude)
            WHERE b.id = v.id;
            """,
            (
                [b.id for b in buildings],
                [b.code for b in buildings],
                [b.latitude for b in buildings],
                [b.longitude for b in buildings],
            ),
        )

    logger.info(f"Updated information for {len(buildings)} buildings")


def main() -> None:
    args = parse_args()
    
    if not args.database_url:
        raise RuntimeError(
            "DATABASE_URL must be provided with --database-url "
            "or the DATABASE_URL environment variable"
        )
    
    ingest(database_url=args.database_url)


if __name__ == "__main__":
    main()