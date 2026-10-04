# Stage 1: Build the Python application
FROM python:3.12-slim-trixie AS builder
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
FROM python:3.12-slim-trixie AS runner
WORKDIR /app

# Create non-root user
RUN groupadd --system --gid 1001 optimizer
RUN useradd --system --uid 1001 --gid optimizer optimizer

# Copy the application
COPY --chown=optimizer:optimizer . .

# Copy production dependencies
COPY --from=builder --chown=optimizer:optimizer /app/.venv ./.venv

USER optimizer

EXPOSE 50051
ENV PORT=50051

ENV PATH="/app/.venv/bin:$PATH"
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Start the optimizer server
CMD ["python", "src/server.py"]