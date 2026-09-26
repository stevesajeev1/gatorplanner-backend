import asyncio
import os

from dotenv import load_dotenv
from grpclib.server import Server

from dependencies.redis import redis
from services.scheduler import SchedulerService

load_dotenv()


async def server(host: str, port: int):
    server = Server([SchedulerService(redis)])

    await server.start(host, port)
    print(f"Optimizer listening on {host}:{port}")

    await server.wait_closed()


if __name__ == "__main__":
    HOST = "0.0.0.0"
    PORT = int(os.environ["PORT"])

    asyncio.run(server(HOST, PORT))
