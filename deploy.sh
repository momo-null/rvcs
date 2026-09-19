#!/usr/bin/env bash
set -euo pipefail

REDEPLOY_DOCKER=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --redeploy-docker)
      REDEPLOY_DOCKER="${2:-}"
      shift 2
      ;;
    -h|--help)
      echo "Usage: ./deploy.sh [--redeploy-docker Y|N]"
      exit 0
      ;;
    *)
      echo "Unknown arg: $1"
      echo "Usage: ./deploy.sh [--redeploy-docker Y|N]"
      exit 1
      ;;
  esac
done

if [[ "${EUID}" -ne 0 ]]; then
  echo "Please run as root (sudo)."
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker not found."
  exit 1
fi

if [[ -f "infra/docker-compose-livekit.yml" ]]; then
  DOCKER_COMPOSE_FILE="infra/docker-compose-livekit.yml"
else
  DOCKER_COMPOSE_FILE="docker-compose-livekit.yml"
fi

if [[ -f "infra/livekit.yaml" ]]; then
  LIVEKIT_CONFIG_FILE="infra/livekit.yaml"
else
  LIVEKIT_CONFIG_FILE="livekit.yaml"
fi

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose -f "$DOCKER_COMPOSE_FILE" "$@"
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose -f "$DOCKER_COMPOSE_FILE" "$@"
  else
    echo "docker compose / docker-compose not found."
    exit 1
  fi
}

require_file() {
  [[ -f "$1" ]] || { echo "Missing file: $1"; exit 1; }
}

require_file service/rvcs-server
require_file service/configs/config.yaml
require_file "$DOCKER_COMPOSE_FILE"
require_file "$LIVEKIT_CONFIG_FILE"

if [[ -z "$REDEPLOY_DOCKER" ]]; then
  read -r -p "Redeploy Docker services (livekit/coturn/nginx)? [y/N]: " ans
  REDEPLOY_DOCKER="$ans"
fi

REDEPLOY_DOCKER="$(echo "$REDEPLOY_DOCKER" | tr '[:upper:]' '[:lower:]')"
case "$REDEPLOY_DOCKER" in
  y|yes)
    REDEPLOY_DOCKER="yes"
    ;;
  n|no|"")
    REDEPLOY_DOCKER="no"
    ;;
  *)
    echo "--redeploy-docker must be Y or N"
    exit 1
    ;;
esac

mkdir -p /opt/rvcs/logs /opt/rvcs/data

if [[ "$REDEPLOY_DOCKER" == "yes" ]]; then
  compose down || true
  compose up -d
fi

pkill -f "rvcs-server" || true

cp -r service/* /opt/rvcs/
cp -r service/configs /opt/rvcs/
cp -r service/web /opt/rvcs/
chmod +x /opt/rvcs/rvcs-server

cat >/opt/rvcs/start.sh <<'EOS'
#!/usr/bin/env bash
set -euo pipefail
cd /opt/rvcs
mkdir -p logs
if pgrep -f "rvcs-server" >/dev/null 2>&1; then
  echo "rvcs-server already running"
  exit 0
fi
nohup ./rvcs-server > logs/rvcs-server.log 2>&1 &
echo $! > rvcs-server.pid
echo "rvcs-server started: PID $(cat rvcs-server.pid)"
EOS

cat >/opt/rvcs/stop.sh <<'EOS'
#!/usr/bin/env bash
set -euo pipefail
cd /opt/rvcs
if [[ -f rvcs-server.pid ]] && kill -0 "$(cat rvcs-server.pid)" 2>/dev/null; then
  kill "$(cat rvcs-server.pid)" || true
  sleep 1
fi
pkill -f "rvcs-server" || true
rm -f rvcs-server.pid
echo "rvcs-server stopped"
EOS

chmod +x /opt/rvcs/start.sh /opt/rvcs/stop.sh

/opt/rvcs/start.sh

echo "Deploy done."
echo "Docker redeploy: $REDEPLOY_DOCKER"
