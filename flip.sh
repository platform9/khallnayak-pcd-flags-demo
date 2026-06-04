#!/usr/bin/env bash
# Terminal 3: toggle the flag without editing the file in $EDITOR.
# Usage: ./flip.sh on   or   ./flip.sh off
set -e
cd "$(dirname "$0")"
target="${1:-}"
case "$target" in
  on|off) ;;
  *)
    echo "usage: $0 on|off" >&2
    exit 1
    ;;
esac

# Toggle defaultVariant in flags.json (BSD sed-safe, macOS).
sed -i.bak -E "s/\"defaultVariant\": \"(on|off)\"/\"defaultVariant\": \"$target\"/" flags.json
rm -f flags.json.bak
echo "==> hamgr.fast_evac defaultVariant -> \"$target\""
grep '"defaultVariant"' flags.json
