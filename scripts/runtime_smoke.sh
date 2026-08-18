#!/usr/bin/env bash
set -euo pipefail

cleanup() {
  docker rm -f workorder-runtime-smoke >/dev/null 2>&1 || true
}

trap cleanup EXIT INT TERM
cleanup
docker build -f benzhi.Dockerfile -t workorder-runtime-smoke:local .
docker run --rm --name workorder-runtime-smoke -p 18078:8080 workorder-runtime-smoke:local
