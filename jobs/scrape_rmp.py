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
    parser = argparse.ArgumentParser(
        description="Ingest RMP data into PostgreSQL."
    )

    parser.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL. Defaults to DATABASE_URL.",
    )

    return parser.parse_args()


def ingest(database_url: str) -> None:
    logger.info(
        "Starting RMP ingestion",
    )

    with psycopg.connect(database_url) as conn, conn.transaction():
        pass


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