# Stage 1: Build the Go application
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-w -s" -o /main ./cmd

# Stage 2: Production runtime
FROM scratch AS runner

COPY --from=builder /main /main
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 1001:1001

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["/main"]