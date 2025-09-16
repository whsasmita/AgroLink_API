# syntax=docker/dockerfile:1.7

# --- Stage 1: Builder ---
# [PERBAIKAN] Ganti versi Go dari 1.22 menjadi 1.24
FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder

ARG TARGETOS TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/agrolink-api ./main.go


# --- Stage 2: Final (Menggunakan Debian Slim) ---
# Tahap ini tidak perlu diubah
FROM debian:bookworm-slim AS final

RUN apt-get update && \
    apt-get install -y --no-install-recommends wkhtmltopdf ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN groupadd --system nonroot && \
    useradd --system --gid nonroot nonroot

WORKDIR /app

COPY --from=builder /out/agrolink-api /app/agrolink-api
COPY templates /app/templates

RUN chown -R nonroot:nonroot /app

ENV GIN_MODE=release
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/agrolink-api"]