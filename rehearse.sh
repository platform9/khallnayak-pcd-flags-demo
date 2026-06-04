#!/usr/bin/env bash
# Single-terminal rehearsal / backup screencast script.
# Runs the entire demo end-to-end (flagd + service + flips) in one go.
# Use this to record a backup video before going on stage.
set -e
cd "$(dirname "$0")"

cleanup() {
  kill $FLAGD_PID $SVC_PID 2>/dev/null || true
  wait 2>/dev/null || true
}
trap cleanup EXIT

# Start fresh: flag OFF
./flip.sh off > /dev/null

echo
echo "================================================================"
echo " PCD Feature Flags — live demo"
echo "================================================================"
echo

echo ">>> step 1: starting flagd (per-DU feature-flag engine)"
~/go/bin/flagd start --uri=file:./flags.json --port=8013 --management-port=8014 \
    > /tmp/pcd-flags-demo/flagd.log 2>&1 &
FLAGD_PID=$!
sleep 2
echo "    flagd is up on :8013 — pid $FLAGD_PID"
echo

echo ">>> step 2: starting fake hamgr service"
echo "    (it will poll hamgr.fast_evac once per second)"
go run service.go &
SVC_PID=$!
sleep 6

echo
echo ">>> step 3: operator flips the flag (\`kubectl edit cm/flagd-config\`-equivalent)"
./flip.sh on
sleep 5

echo
echo ">>> step 4: operator rolls back"
./flip.sh off
sleep 5

echo
echo "================================================================"
echo " demo complete — no service restart, no redeploy"
echo "================================================================"
