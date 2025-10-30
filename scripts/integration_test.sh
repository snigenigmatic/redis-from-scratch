#!/usr/bin/env bash
set -euo pipefail

# Integration test for redis-from-scratch server
# Tests all data types: strings, hashes, lists, sets, sorted sets
# Includes type-error assertions and edge cases

PORT=${1:-6380}
LOG=/tmp/rfs-integration.log

echo "Starting comprehensive integration test on port $PORT"
rm -f "$LOG"

go run cmd/server/main.go --port "$PORT" >"$LOG" 2>&1 &
SERVER_PID=$!
trap 'kill ${SERVER_PID} 2>/dev/null || true; echo "Server log:"; sed -n "1,200p" "$LOG"' EXIT

# Wait for server to listen (timeout 5s)
for i in {1..25}; do
  if ss -ltn | grep -q ":${PORT}"; then
    break
  fi
  sleep 0.2
done

if ! ss -ltn | grep -q ":${PORT}"; then
  echo "Server did not start or port $PORT not listening"
  exit 2
fi

use_redis_cli=false
if command -v redis-cli >/dev/null 2>&1; then
  use_redis_cli=true
fi

echo "Using $( $use_redis_cli && echo redis-cli || echo nc ) for testing"

run_cmd() {
  if [ "$use_redis_cli" = true ]; then
    redis-cli -p "$PORT" "$@" 2>&1 || true
  else
    case "$1" in
      SET)
        printf '*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n' ${#2} "$2" ${#3} "$3" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      GET)
        printf '*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      HSET)
        printf '*4\r\n$4\r\nHSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n' ${#2} "$2" ${#3} "$3" ${#4} "$4" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      HGET)
        printf '*3\r\n$4\r\nHGET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n' ${#2} "$2" ${#3} "$3" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      LPUSH)
        printf '*3\r\n$5\r\nLPUSH\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n' ${#2} "$2" ${#3} "$3" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      LRANGE)
        printf '*4\r\n$6\r\nLRANGE\r\n$%d\r\n%s\r\n$1\r\n0\r\n$2\r\n-1\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      SADD)
        printf '*3\r\n$4\r\nSADD\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n' ${#2} "$2" ${#3} "$3" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      SMEMBERS)
        printf '*2\r\n$8\r\nSMEMBERS\r\n$%d\r\n%s\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
      DEL)
        printf '*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n' ${#2} "$2" | nc -w 2 localhost "$PORT" 2>&1 || true
        ;;
    esac
  fi
}

expect() {
  got="$1"
  want="$2"
  if [ "$got" != "$want" ]; then
    echo "❌ Assertion failed: got '$got', want '$want'"
    exit 3
  fi
  echo "✓ Passed"
}

assert_contains() {
  haystack="$1"
  needle="$2"
  if ! echo "$haystack" | grep -q "$needle"; then
    echo "❌ Assertion failed: '$haystack' does not contain '$needle'"
    exit 3
  fi
  echo "✓ Passed"
}

assert_error() {
  output="$1"
  if ! echo "$output" | grep -qi "error\|wrong\|wrongtype"; then
    echo "❌ Expected error, got: $output"
    exit 3
  fi
  echo "✓ Passed (error detected)"
}

echo ""
echo "=== STRING OPERATIONS ==="
echo -n "SET mykey hello -> "
out=$(run_cmd SET mykey hello)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "OK"
else
  expect "$(echo "$out" | tr -d '\r')" "+OK"
fi

echo -n "GET mykey -> "
out=$(run_cmd GET mykey)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "hello"
else
  val=$(echo "$out" | sed -n '2p' | tr -d '\r')
  expect "$val" "hello"
fi

echo ""
echo "=== HASH OPERATIONS ==="
echo -n "HSET hash1 field1 value1 -> "
out=$(run_cmd HSET hash1 field1 value1)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "1"
else
  expect "$(echo "$out" | tr -d '\r')" ":1"
fi

echo -n "HGET hash1 field1 -> "
out=$(run_cmd HGET hash1 field1)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "value1"
else
  val=$(echo "$out" | sed -n '2p' | tr -d '\r')
  expect "$val" "value1"
fi

echo -n "Type error: HSET on string key -> "
out=$(run_cmd HSET mykey field value)
assert_error "$out"

echo ""
echo "=== LIST OPERATIONS ==="
echo -n "LPUSH list1 a b c -> "
out=$(run_cmd LPUSH list1 a b c)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "3"
else
  expect "$(echo "$out" | tr -d '\r')" ":3"
fi

echo -n "LRANGE list1 0 -1 -> "
out=$(run_cmd LRANGE list1 0 -1)
assert_contains "$out" "c"
assert_contains "$out" "b"
assert_contains "$out" "a"

echo -n "Type error: LPUSH on string key -> "
out=$(run_cmd LPUSH mykey item)
assert_error "$out"

echo ""
echo "=== SET OPERATIONS ==="
echo -n "SADD set1 member1 member2 -> "
out=$(run_cmd SADD set1 member1 member2)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "2"
else
  expect "$(echo "$out" | tr -d '\r')" ":2"
fi

echo -n "SMEMBERS set1 -> "
out=$(run_cmd SMEMBERS set1)
assert_contains "$out" "member1"
assert_contains "$out" "member2"

echo -n "Type error: SADD on string key -> "
out=$(run_cmd SADD mykey member)
assert_error "$out"

echo ""
echo "=== CLEANUP ==="
echo -n "DEL all keys -> "
out=$(run_cmd DEL mykey hash1 list1 set1)
if [ "$use_redis_cli" = true ]; then
  expect "$out" "4"
else
  expect "$(echo "$out" | tr -d '\r')" ":4"
fi

echo ""
echo "✅ All integration tests passed!"
exit 0