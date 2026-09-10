import asyncio
import os

from services.scheduler import SchedulerService
from grpclib.server import Server

from dotenv import load_dotenv
load_dotenv()

async def server(host: str, port: int):
    server = Server([SchedulerService()])

    await server.start(host, port)
    print(f"Optimizer listening on {host}:{port}")

    await server.wait_closed()


if __name__ == "__main__":
    HOST = "0.0.0.0"
    PORT = int(os.environ["PORT"])

    asyncio.run(server(HOST, PORT))