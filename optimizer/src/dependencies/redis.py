import hashlib
import os
from typing import cast

import betterproto2
from dotenv import load_dotenv
from redis.asyncio import Redis

load_dotenv()

CACHE_TTL_SECONDS = 60 * 60

redis = Redis(
    host=os.environ["REDIS_HOST"],
    port=int(os.environ["REDIS_PORT"]),
    decode_responses=False,
)


class ProtoCache[TRequest: betterproto2.Message, TResponse: betterproto2.Message]:
    def __init__(
        self,
        redis: Redis,
        prefix: str,
        response_type: type[TResponse],
    ):
        self.redis = redis
        self.prefix = prefix
        self.response_type = response_type

    def _get_cache_key(self, request: TRequest) -> str:
        data = request.SerializeToString()
        digest = hashlib.sha256(data).hexdigest()

        return f"{self.prefix}:{digest}"

    async def get(self, request: TRequest) -> TResponse | None:
        key = self._get_cache_key(request)

        data = await self.redis.get(key)
        if data is None:
            return None

        response = self.response_type.FromString(cast(bytes, data))

        return response

    async def set(
        self,
        request: TRequest,
        response: TResponse,
        ttl: int = CACHE_TTL_SECONDS,
    ) -> None:
        key = self._get_cache_key(request)

        data = response.SerializeToString()

        await self.redis.set(
            key,
            data,
            ex=ttl,
        )
