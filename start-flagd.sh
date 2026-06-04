#!/usr/bin/env bash
# Terminal 1: run flagd in foreground watching flags.json.
# It will hot-reload on every save.
set -e
cd "$(dirname "$0")"
echo "==> starting flagd on :8013 (mgmt :8014), watching ./flags.json"
exec ~/go/bin/flagd start --uri=file:./flags.json --port=8013 --management-port=8014
