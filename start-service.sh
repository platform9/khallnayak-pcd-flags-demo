#!/usr/bin/env bash
# Terminal 2: run the fake hamgr service. Polls the flag once per second
# and prints which evac path it would take.
set -e
cd "$(dirname "$0")"
echo "==> starting fake hamgr service"
exec go run service.go
