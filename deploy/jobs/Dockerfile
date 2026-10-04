# Stage 1: Build the Python application
FROM python:3.13-slim-trixie AS builder
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/
WORKDIR /app

COPY pyproject.toml uv.lock ./

ENV UV_NO_DEV=1
ENV UV_COMPILE_BYTECODE=1
ENV UV_LINK_MODE=copy
ENV UV_PYTHON_DOWNLOADS=0

RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-install-project

# Stage 2: Production runtime
FROM python:3.13-slim-trixie AS runner
WORKDIR /app

# Create non-root user
RUN groupadd --system --gid 1001 jobs
RUN useradd --system --uid 1001 --gid jobs jobs

# Copy the application
COPY --chown=jobs:jobs . .

# Copy production dependencies
COPY --from=builder --chown=jobs:jobs /app/.venv ./.venv

USER jobs

ENV PATH="/app/.venv/bin:$PATH"
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1