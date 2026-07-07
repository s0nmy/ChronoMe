#!/bin/sh
set -eu

export SERVER_ADDRESS="${SERVER_ADDRESS:-:8081}"

/usr/local/bin/chronome-server &
BACKEND_PID="$!"

term() {
    kill -TERM "$BACKEND_PID" 2>/dev/null || true
    nginx -s quit 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
}

trap term INT TERM

nginx -g "daemon off;" &
NGINX_PID="$!"

set +e
wait -n "$BACKEND_PID" "$NGINX_PID"
STATUS="$?"
set -e
term
exit "$STATUS"
