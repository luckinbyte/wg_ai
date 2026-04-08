#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="$ROOT_DIR/runtime/pid"
LOG_DIR="$ROOT_DIR/runtime/logs"

mkdir -p "$PID_DIR" "$LOG_DIR"

DB_PID_FILE="$PID_DIR/db.pid"
LOGIN_PID_FILE="$PID_DIR/login.pid"
GAME_PID_FILE="$PID_DIR/game.pid"

DB_LOG_FILE="$LOG_DIR/db.log"
LOGIN_LOG_FILE="$LOG_DIR/login.log"
GAME_LOG_FILE="$LOG_DIR/game.log"
GAME_CONFIG_FILE="$ROOT_DIR/config/game.yaml"

is_pid_running() {
  local pid="$1"
  kill -0 "$pid" 2>/dev/null
}

read_pid() {
  local pid_file="$1"
  [[ -f "$pid_file" ]] || return 1
  local pid
  pid="$(<"$pid_file")"
  [[ -n "$pid" ]] || return 1
  printf '%s\n' "$pid"
}

cleanup_stale_pid() {
  local name="$1"
  local pid_file="$2"
  local pid
  if pid="$(read_pid "$pid_file" 2>/dev/null)" && ! is_pid_running "$pid"; then
    rm -f "$pid_file"
    echo "[$name] removed stale pid $pid"
  fi
}

check_port() {
  local host="$1"
  local port="$2"
  python3 - "$host" "$port" <<'PY'
import socket, sys
host = sys.argv[1]
port = int(sys.argv[2])
sock = socket.socket()
sock.settimeout(0.5)
try:
    sock.connect((host, port))
except OSError:
    sys.exit(1)
finally:
    sock.close()
sys.exit(0)
PY
}

ensure_port_free() {
  local name="$1"
  local host="$2"
  local port="$3"
  if check_port "$host" "$port"; then
    echo "[$name] port $port is already in use"
    return 1
  fi
}

wait_for_port() {
  local name="$1"
  local host="$2"
  local port="$3"
  local pid_file="$4"
  local timeout="$5"
  local start_ts
  start_ts="$(date +%s)"

  while true; do
    local pid
    if ! pid="$(read_pid "$pid_file" 2>/dev/null)" || ! is_pid_running "$pid"; then
      echo "[$name] process exited before becoming ready"
      return 1
    fi

    if check_port "$host" "$port"; then
      echo "[$name] port $port is ready"
      return 0
    fi

    if (( $(date +%s) - start_ts >= timeout )); then
      echo "[$name] readiness timeout on port $port"
      return 1
    fi
    sleep 1
  done
}

game_ws_port() {
  python3 - "$GAME_CONFIG_FILE" <<'PY'
import pathlib
import re
import sys

path = pathlib.Path(sys.argv[1])
try:
    text = path.read_text(encoding='utf-8')
except OSError:
    sys.exit(1)

match = re.search(r'^\s*ws_port\s*:\s*(\d+)\b', text, re.MULTILINE)
if not match:
    sys.exit(1)

port = int(match.group(1))
if port > 0:
    print(port)
    sys.exit(0)

sys.exit(1)
PY
}

has_game_ws() {
  game_ws_port >/dev/null
}

start_db() {
  cleanup_stale_pid db "$DB_PID_FILE"
  local pid
  if pid="$(read_pid "$DB_PID_FILE" 2>/dev/null)" && is_pid_running "$pid"; then
    echo "[db] already running pid=$pid"
    return 0
  fi

  ensure_port_free db 127.0.0.1 50052 || return 1

  nohup "$ROOT_DIR/bin/db" -config "$ROOT_DIR/config/db.yaml" >>"$DB_LOG_FILE" 2>&1 &
  pid=$!
  printf '%s\n' "$pid" > "$DB_PID_FILE"
  echo "[db] started pid=$pid"
  wait_for_port db 127.0.0.1 50052 "$DB_PID_FILE" 15
}

start_login() {
  cleanup_stale_pid login "$LOGIN_PID_FILE"
  local pid
  if pid="$(read_pid "$LOGIN_PID_FILE" 2>/dev/null)" && is_pid_running "$pid"; then
    echo "[login] already running pid=$pid"
    return 0
  fi

  ensure_port_free login 127.0.0.1 50051 || return 1

  nohup "$ROOT_DIR/bin/login" -config "$ROOT_DIR/config/login.yaml" >>"$LOGIN_LOG_FILE" 2>&1 &
  pid=$!
  printf '%s\n' "$pid" > "$LOGIN_PID_FILE"
  echo "[login] started pid=$pid"
  wait_for_port login 127.0.0.1 50051 "$LOGIN_PID_FILE" 15
}

start_game() {
  cleanup_stale_pid game "$GAME_PID_FILE"
  local pid
  if pid="$(read_pid "$GAME_PID_FILE" 2>/dev/null)" && is_pid_running "$pid"; then
    echo "[game] already running pid=$pid"
    return 0
  fi

  ensure_port_free game 127.0.0.1 44445 || return 1

  local ws_port=""
  if ws_port="$(game_ws_port)"; then
    ensure_port_free game-ws 127.0.0.1 "$ws_port" || return 1
  fi

  nohup "$ROOT_DIR/bin/game" -config "$ROOT_DIR/config/game.yaml" >>"$GAME_LOG_FILE" 2>&1 &
  pid=$!
  printf '%s\n' "$pid" > "$GAME_PID_FILE"
  echo "[game] started pid=$pid"
  wait_for_port game 127.0.0.1 44445 "$GAME_PID_FILE" 15
  if [[ -n "$ws_port" ]]; then
    wait_for_port game-ws 127.0.0.1 "$ws_port" "$GAME_PID_FILE" 15
  fi
}

stop_service() {
  local name="$1"
  local pid_file="$2"
  local timeout="$3"

  cleanup_stale_pid "$name" "$pid_file"
  local pid
  if ! pid="$(read_pid "$pid_file" 2>/dev/null)"; then
    echo "[$name] not running"
    return 0
  fi

  echo "[$name] stopping pid=$pid"
  kill -TERM "$pid" 2>/dev/null || true

  local start_ts
  start_ts="$(date +%s)"
  while is_pid_running "$pid"; do
    if (( $(date +%s) - start_ts >= timeout )); then
      echo "[$name] stop timeout, sending SIGKILL"
      kill -KILL "$pid" 2>/dev/null || true
      break
    fi
    sleep 1
  done

  rm -f "$pid_file"
  echo "[$name] stopped"
}

port_status() {
  local host="$1"
  local port="$2"
  if check_port "$host" "$port"; then
    printf 'yes'
  else
    printf 'no'
  fi
}

status_service() {
  local name="$1"
  local pid_file="$2"
  local host="$3"
  local port="$4"

  cleanup_stale_pid "$name" "$pid_file"

  local pid="-"
  local alive="no"
  if pid_value="$(read_pid "$pid_file" 2>/dev/null)"; then
    pid="$pid_value"
    if is_pid_running "$pid_value"; then
      alive="yes"
    fi
  fi

  local port_ready
  port_ready="$(port_status "$host" "$port")"

  local line="[$name] pid=$pid alive=$alive port=$port_ready"
  local warning=""
  if [[ "$alive" == "yes" && "$port_ready" == "no" ]]; then
    warning="port-mismatch"
  elif [[ "$alive" == "no" && "$port_ready" == "yes" ]]; then
    warning="external-port-in-use"
  fi

  if [[ "$name" == "game" ]]; then
    local ws_port
    if ws_port="$(game_ws_port)"; then
      local ws_ready
      ws_ready="$(port_status 127.0.0.1 "$ws_port")"
      line="$line ws=$ws_ready"
      if [[ "$alive" == "yes" && "$ws_ready" == "no" ]]; then
        warning="${warning:+$warning,}ws-port-mismatch"
      elif [[ "$alive" == "no" && "$ws_ready" == "yes" ]]; then
        warning="${warning:+$warning,}external-ws-port-in-use"
      fi
    fi
  fi

  if [[ -n "$warning" ]]; then
    line="$line warning=$warning"
  fi

  echo "$line"
}

start_all() {
  start_db || return 1
  start_login || {
    stop_service db "$DB_PID_FILE" 10
    return 1
  }
  start_game || {
    stop_service game "$GAME_PID_FILE" 30 || true
    stop_service login "$LOGIN_PID_FILE" 10 || true
    stop_service db "$DB_PID_FILE" 10 || true
    return 1
  }
}

stop_all() {
  stop_service game "$GAME_PID_FILE" 30
  stop_service login "$LOGIN_PID_FILE" 10
  stop_service db "$DB_PID_FILE" 10
}

restart_all() {
  stop_all
  start_all
}

status_all() {
  status_service db "$DB_PID_FILE" 127.0.0.1 50052
  status_service login "$LOGIN_PID_FILE" 127.0.0.1 50051
  status_service game "$GAME_PID_FILE" 127.0.0.1 44445
}

case "${1:-}" in
  start) start_all ;;
  stop) stop_all ;;
  restart) restart_all ;;
  status) status_all ;;
  *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac
