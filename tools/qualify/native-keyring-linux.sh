#!/usr/bin/env bash
set -euo pipefail

: "${KEYRING_BIN:?KEYRING_BIN is required}"
: "${KEYRING_EVIDENCE:?KEYRING_EVIDENCE is required}"

# dbus-run-session owns this bus; this foreground process owns its fixture store.
gnome-keyring-daemon --foreground --unlock --components=secrets \
  < /dev/null > "$KEYRING_EVIDENCE/secret-service-start.txt" 2>&1 &
keyring_pid=$!
trap 'kill "$keyring_pid" 2>/dev/null || true; wait "$keyring_pid" 2>/dev/null || true' EXIT
gdbus wait --session --timeout 15 org.freedesktop.secrets
go tool test2json -p github.com/Wolfsblvt/smolpng/internal/secret -t \
  "$KEYRING_BIN/keyring-test" -test.v=test2json \
  -test.run '^TestNativeCredentialsLifecycle$' -test.timeout 120s \
  > "$KEYRING_EVIDENCE/native-tests.jsonl"
