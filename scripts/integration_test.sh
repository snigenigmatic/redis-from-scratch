#!/usr/bin/env bash
set -euo pipefail

# Integration test for redis-from-scratch server
# Starts the server on PORT (default 6380), runs a sequence of commands (SET/GET/EXISTS/KEYS/DEL)
# and asserts expected RESP replies. Uses redis-cli if available, otherwise falls back to netcat.

PORT=${1:-6380}
LOG=/tmp/rfs-integration.log

echo "Starting integration test on port $PORT"

rm -f "$LOG"

go run cmd/server/main.go --port "$PORT" >"$LOG" 2>&1 &
SERVER_PID=$!
trap 'kill ${SERVER_PID} 2>/dev/null || true; echo "Server log:"; sed -n "1,200p" "$LOG"' EXIT

# wait for server to listen (timeout 5s)
for i in {1..25}; do
  if ss -ltn | grep -q ":${PORT}"; then
    break
  fi
  sleep 0.2
done

if ! ss -ltn | grep -q ":${PORT}"; then
  echo "Server did not start or port $PORT not listening"
  echo "---- server log ----"
  sed -n '1,200p' "$LOG" || true
  exit 2
fi

use_redis_cli=false
if command -v redis-cli >/dev/null 2>&1; then
  use_redis_cli=true
fi

echo "Using $( $use_redis_cli && echo redis-cli || echo nc ) for testing"

run_cmd() {
  # run a command and print output
  if [ "$use_redis_cli" = true ]; then
    redis-cli -p "$PORT" "$@"
  else
    case "$1" in
      SET)
        printf '*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n' ${#2} "$2" ${#3} "$3" | nc -w 2 localhost "$PORT"
        ;;
      GET)
        printf '*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT"
        ;;
      EXISTS)
        printf '*2\r\n$6\r\nEXISTS\r\n$%d\r\n%s\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT"
        ;;
      KEYS)
        printf '*2\r\n$4\r\nKEYS\r\n$%d\r\n%s\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT"
        ;;
      DEL)
        printf '*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT"
        ;;
      *)
        echo "Unsupported fallback command: $@"; return 2
        ;;
    esac
  fi
}

expect() {
  got="$1"
  want="$2"
  if [ "$got" != "$want" ]; then
    echo "Assertion failed:\n  got: '$got'\n  want: '$want'"
    exit 3
  fi
}

echo "SET mykey hello -> expect OK"
out=$(run_cmd SET mykey hello)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "OK"
else
  # nc returns +OK with CRLF; normalize
  expect "$(echo "$out" | tr -d '\r')" "+OK"
fi

echo "GET mykey -> expect hello"
out=$(run_cmd GET mykey)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "hello"
else
  # nc returns bulk string: $5\r\nhello\r\n
  # extract the value line
  val=$(echo "$out" | sed -n '2p' | tr -d '\r')
  expect "$val" "hello"
fi

echo "EXISTS mykey -> expect 1"
out=$(run_cmd EXISTS mykey)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "1"
else
  expect "$(echo "$out" | tr -d '\r')" ":1"
fi

echo "KEYS * -> expect mykey in array"
out=$(run_cmd KEYS "*")
if [ "$use_redis_cli" = true ]; then
  # redis-cli prints keys one per line
  if ! echo "$out" | grep -q '^mykey$'; then
    echo "KEYS did not return mykey:\n$out"; exit 4
  fi
else
  # raw array; verify second line contains key
  key=$(echo "$out" | sed -n '3p' | tr -d '\r')
  if [ "$key" != "mykey" ]; then
    echo "KEYS did not return mykey: got='$key'"; exit 4
  fi
fi

echo "DEL mykey -> expect 1"
out=$(run_cmd DEL mykey)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "1"
else
  expect "$(echo "$out" | tr -d '\r')" ":1"
fi

echo "GET mykey after DEL -> expect (nil) / $-1"
out=$(run_cmd GET mykey)
if [ "$use_redis_cli" = true ]; then
  # redis-cli may print (nil) or nothing for a missing key; accept either
  if [ -z "$out" ] || [ "$out" = "(nil)" ]; then
    : # ok
  else
    echo "Assertion failed: expected (nil) or empty, got: '$out'"; exit 3
  fi
else
  expect "$(echo "$out" | tr -d '\r')" "$\-1"
fi

echo "Integration test passed"

exit 0
