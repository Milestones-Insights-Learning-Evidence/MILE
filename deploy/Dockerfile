# MILE (Milestones, Insight, Learning & Evidence) - Multi-Stage Single-Binary Container Build
# Implementation Target: Phase 5 (Dev Compose) & Phase 8 (Production-ready Community release)
# Reference: docs/adr/ADR-01-go-sqlc-single-binary.md & docs/IMPLEMENTATION_PLAN.md §7 (Phase 8)
#
# Build from the repository root (the build context is the whole module, not deploy/):
#   docker build -f deploy/Dockerfile -t mile:dev .
#
# ---------------------------------------------------------------------------
# Required container configuration
# ---------------------------------------------------------------------------
# The binary's default bind address is 127.0.0.1:8000, which is reachable only from
# inside the container. To accept published traffic, set the listener explicitly:
#
#   MILE_HTTP_ADDR=0.0.0.0:8000
#
# Binding beyond loopback makes MILE_METRICS_TOKEN mandatory: startup validation in
# internal/config rejects a token shorter than 32 non-whitespace characters whenever
# MILE_HTTP_ADDR is non-loopback or MILE_ENV=production. /metrics is only
# unauthenticated for loopback peers in development. Generate and mount one, e.g.:
#
#   openssl rand -hex 32 > ./secrets/mile_metrics_token
#   docker run --rm -p 127.0.0.1:8000:8000 \
#     -e MILE_HTTP_ADDR=0.0.0.0:8000 \
#     -e MILE_METRICS_TOKEN_FILE=/run/secrets/mile_metrics_token \
#     -v "$PWD/secrets/mile_metrics_token:/run/secrets/mile_metrics_token:ro" \
#     mile:dev
#
# Without that token the container exits at startup with a configuration error
# instead of exposing an unauthenticated /metrics endpoint.
#
# ---------------------------------------------------------------------------
# Base image pinning
# ---------------------------------------------------------------------------
# Both stages are pinned to an explicit patch tag AND an immutable multi-arch
# manifest digest, verified against registry-1.docker.io on 2026-11-07. Floating
# tags (`golang:1.26-alpine`, `alpine:3.20`) must not be used here. To bump, resolve
# the new digest first: `docker buildx imagetools inspect <image>:<tag>`.

FROM golang:1.26.10-alpine3.24@sha256:87c65a688975c0b1d4585d204874c45ca78354670e3a5f11368c5a492303d696 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mile ./cmd/mile

FROM alpine:3.24.4@sha256:93b41ef40b1677fce8fcc64ff4af09daddfacf5040afde618bc46c1e2d1cc7d4
RUN adduser -D -u 10001 mile
USER mile
WORKDIR /app
COPY --from=builder /out/mile /usr/local/bin/mile

# Matches the default MILE_HTTP_ADDR port (127.0.0.1:8000). Keep EXPOSE, the compose
# port mappings, and MILE_HTTP_ADDR in agreement when changing this.
EXPOSE 8000

# Liveness only: /healthz stays 200 while the process drains, so orchestrators that
# need drain-aware routing must probe /readyz instead. BusyBox wget is used because
# the runtime image intentionally ships no curl. 127.0.0.1 works whether the listener
# is bound to loopback or 0.0.0.0.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --spider --tries=1 http://127.0.0.1:8000/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/mile"]
CMD ["serve"]
