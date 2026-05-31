#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# TRBB 鐵人三項運動社團 — Service Management Script
# Usage: ./service.sh [command]
# ═══════════════════════════════════════════════════════════════

set -euo pipefail

PROJECT_NAME="trbb"
NODE_IMAGE="node:20-alpine"
GO_IMAGE="golang:1.22-alpine"
BOLD="\033[1m"
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
CYAN="\033[36m"
RESET="\033[0m"

# ── Helpers ──────────────────────────────────────────────────
# 全部輸出到 stderr，避免被 $() command substitution 捕獲
log_info()    { echo -e "${GREEN}[INFO]${RESET}  $*" >&2; }
log_warn()    { echo -e "${YELLOW}[WARN]${RESET}  $*" >&2; }
log_error()   { echo -e "${RED}[ERROR]${RESET} $*" >&2; }
log_section() { echo -e "\n${BOLD}${CYAN}══ $* ══${RESET}\n" >&2; }

# ── Detect docker compose command ────────────────────────────
# Priority: 1) docker compose (plugin)  2) docker-compose (standalone)
# ── Version comparison helper ────────────────────────────────
# version_gte "2.20.1" "2.17.0" → 0 (true) if first >= second
version_gte() {
  local a b
  a=$(echo "$1" | sed 's/^v//')
  b=$(echo "$2" | sed 's/^v//')
  [ "$(printf '%s\n%s' "$a" "$b" | sort -V | head -1)" = "$b" ]
}

# ── Detect docker compose command ────────────────────────────
# Priority:
#   1) docker compose (plugin) >= 2.17.0  AND  buildx >= 0.17.0
#   2) docker-compose (standalone) >= 1.29.0  (no buildx required)
detect_compose() {
  # ── 1. docker compose plugin ────────────────────────────────
  if docker compose version &>/dev/null 2>&1; then
    local cver bxver
    cver=$(docker compose version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    bxver=$(docker buildx version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)

    if [ -n "$cver" ] && version_gte "$cver" "2.17.0"; then
      if [ -n "$bxver" ] && version_gte "$bxver" "0.17.0"; then
        echo "docker compose"
        return 0
      else
        log_warn "'docker compose' v${cver} OK, but buildx ${bxver:-not found} < 0.17.0"
        log_warn "Falling back to 'docker-compose' standalone (no buildx needed)..."
      fi
    else
      log_warn "'docker compose' version ${cver:-unknown} < 2.17.0"
    fi
  fi

  # ── 2. docker-compose standalone ────────────────────────────
  if command -v docker-compose &>/dev/null; then
    local ver
    ver=$(docker-compose version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    if [ -n "$ver" ] && version_gte "$ver" "1.29.0"; then
      echo "docker-compose"
      return 0
    else
      log_warn "'docker-compose' version ${ver:-unknown} < 1.29.0"
    fi
  fi

  echo ""
  return 1
}

# ── Lazy-init COMPOSE ────────────────────────────────────────
COMPOSE=""
init_compose() {
  if [ -z "$COMPOSE" ]; then
    if ! command -v docker &>/dev/null; then
      log_error "Docker not found. Please install Docker first."
      exit 1
    fi

    COMPOSE=$(detect_compose)

    if [ -z "$COMPOSE" ]; then
      log_error "No usable Docker Compose found."
      log_error ""
      log_error "Current versions:"
      docker compose  version 2>/dev/null && true
      docker buildx   version 2>/dev/null && true
      docker-compose  version 2>/dev/null && true
      log_error ""
      log_error "Requirements (pick one option):"
      log_error ""
      log_error "  Option A — upgrade buildx to >= 0.17.0:"
      log_error "    BXVER=\$(curl -s https://api.github.com/repos/docker/buildx/releases/latest | grep tag_name | cut -d'\"' -f4)"
      log_error "    sudo mkdir -p /usr/local/lib/docker/cli-plugins"
      log_error "    sudo curl -SL \"https://github.com/docker/buildx/releases/download/\${BXVER}/buildx-\${BXVER}.linux-amd64\" \\"
      log_error "         -o /usr/local/lib/docker/cli-plugins/docker-buildx"
      log_error "    sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-buildx"
      log_error ""
      log_error "  Option B — install docker-compose standalone >= 1.29 (no buildx needed):"
      log_error "    sudo curl -SL \"https://github.com/docker/compose/releases/download/v1.29.2/docker-compose-linux-x86_64\" \\"
      log_error "         -o /usr/local/bin/docker-compose"
      log_error "    sudo chmod +x /usr/local/bin/docker-compose"
      exit 1
    fi

    log_info "Using compose: ${BOLD}${COMPOSE}${RESET}"
  fi
}

check_env() {
  if [ ! -f ".env" ]; then
    log_warn ".env not found — copying from .env.example"
    [ -f ".env.example" ] && cp .env.example .env || true
  fi
  if [ ! -f "backend/.env" ]; then
    log_warn "backend/.env not found — copying from backend/.env.example"
    [ -f "backend/.env.example" ] && cp backend/.env.example backend/.env || true
  fi
}

# ── go mod tidy ───────────────────────────────────────────────
go_mod_tidy() {
  log_info "Running go mod tidy via Docker (${GO_IMAGE})..."
  docker pull "${GO_IMAGE}" --quiet 2>/dev/null || true
  docker run --rm \
    -v "$(pwd)/backend:/workspace" \
    -w /workspace \
    -e GONOSUMDB="*" \
    -e GOFLAGS="-mod=mod" \
    -e GOPROXY="direct" \
    "${GO_IMAGE}" \
    sh -c "apk add --no-cache git && go mod download && go mod tidy && echo '[go] go.sum ready'"
  log_info "✓ go mod tidy done"
}

# ── Build frontend (Docker Node) ─────────────────────────────
build_frontend_app() {
  local app="$1"
  local dir="frontend/${app}"
  if [ ! -d "$dir" ]; then
    log_warn "Directory $dir not found, skipping."
    return 0
  fi
  log_info "[$app] npm install..."
  docker run --rm \
    -v "$(pwd)/${dir}:/workspace" \
    -w /workspace \
    "${NODE_IMAGE}" \
    sh -c "npm install --prefer-offline 2>/dev/null || npm install"
  log_info "[$app] npm run build..."
  docker run --rm \
    -v "$(pwd)/${dir}:/workspace" \
    -w /workspace \
    "${NODE_IMAGE}" \
    sh -c "npm run build"
  log_info "✓ [$app] built → ${dir}/dist"
}

build_frontend() {
  log_section "Building Frontend (via Docker ${NODE_IMAGE})"
  log_info "Pulling ${NODE_IMAGE}..."
  docker pull "${NODE_IMAGE}" --quiet
  for app in web app admin; do
    build_frontend_app "$app"
  done
  log_info "✓ All frontends built."
}

# ── init ──────────────────────────────────────────────────────
cmd_init() {
  log_section "Initialising TRBB Project"
  init_compose
  check_env
  mkdir -p nginx/logs logs mysql/data redis/data minio/data
  log_info "Pulling Docker images..."
  $COMPOSE pull --quiet || true
  go_mod_tidy
  build_frontend
  log_info "Building backend Docker image..."
  $COMPOSE build backend
  log_info "Starting all services..."
  $COMPOSE up -d
  log_info "Waiting for MySQL to be ready..."
  local tries=0
  until $COMPOSE exec mysql mysqladmin ping -h localhost --silent 2>/dev/null; do
    tries=$((tries + 1))
    if [ "$tries" -ge 30 ]; then
      log_error "MySQL did not become ready. Check: ./service.sh logs mysql"
      exit 1
    fi
    sleep 2
  done
  log_info "✓ All services initialised."
  _print_urls
}

# ── build ─────────────────────────────────────────────────────
cmd_build() {
  log_section "Building Project"
  init_compose
  check_env
  go_mod_tidy
  build_frontend
  log_info "Building backend Docker image..."
  $COMPOSE build backend
  log_info "✓ Build complete."
}

# ── build-app ─────────────────────────────────────────────────
cmd_build_app() {
  local app="${2:-}"
  if [ -z "$app" ]; then
    log_error "Usage: ./service.sh build-app [web|app|admin]"
    exit 1
  fi
  init_compose
  log_section "Building Frontend: $app"
  docker pull "${NODE_IMAGE}" --quiet
  build_frontend_app "$app"
}

# ── build-backend ─────────────────────────────────────────────
cmd_build_backend() {
  log_section "Building Backend"
  init_compose
  log_info "Pulling latest base image..."
  docker pull golang:1.22-alpine
  log_info "Building backend Docker image (no cache)..."
  $COMPOSE build --no-cache backend
  log_info "✓ Backend built."
}

# ── start ─────────────────────────────────────────────────────
cmd_start() {
  log_section "Starting Services"
  init_compose
  check_env
  $COMPOSE up -d
  log_info "✓ Services started."
  cmd_status
}

# ── stop ──────────────────────────────────────────────────────
cmd_stop() {
  log_section "Stopping Services"
  init_compose
  $COMPOSE stop
  log_info "✓ Services stopped."
}

# ── restart ───────────────────────────────────────────────────
cmd_restart() {
  log_section "Restarting Services"
  init_compose
  check_env
  $COMPOSE restart
  log_info "✓ Services restarted."
}

# ── rebuild ───────────────────────────────────────────────────
cmd_rebuild() {
  log_section "Rebuilding & Restarting"
  init_compose
  check_env

  # 1. Stop
  log_info "Stopping containers..."
  $COMPOSE down

  # 2. Start DB only (needed for migrate)
  log_info "Starting database for migration..."
  $COMPOSE up -d mysql
  log_info "Waiting for MySQL to be ready..."
  local tries=0
  until $COMPOSE exec -T mysql mysqladmin ping -u root \
      "-p$(grep MYSQL_ROOT_PASSWORD .env | cut -d= -f2)" \
      --silent 2>/dev/null; do
    tries=$((tries+1))
    if [ $tries -ge 30 ]; then
      log_warn "MySQL not ready after 30s, skipping wait"
      break
    fi
    sleep 2
  done

  # 3. Migrate
  log_info "Running migrations..."
  local root_pw; root_pw=$(grep MYSQL_ROOT_PASSWORD .env | cut -d= -f2)
  local db_name; db_name=$(grep MYSQL_DATABASE .env | cut -d= -f2)
  for f in mysql/init/*.sql; do
    $COMPOSE exec -T mysql mysql -u root "-p${root_pw}" "${db_name}" \
      < "$f" 2>/dev/null && \
      log_info "  ✓ $f" || log_warn "  ⚠ $f (skipped or already applied)"
  done

  # 4. go mod tidy + build frontend
  go_mod_tidy
  build_frontend

  # 5. Build backend (pull fresh base image first)
  log_info "Pulling latest base image..."
  docker pull golang:1.22-alpine
  log_info "Rebuilding backend (no cache)..."
  $COMPOSE build --no-cache backend

  # 6. Start all
  log_info "Starting all services..."
  $COMPOSE up -d
  log_info "✓ Rebuild complete."
  cmd_status
}

# ── status ────────────────────────────────────────────────────
cmd_status() {
  log_section "Service Status"
  init_compose
  $COMPOSE ps
}

# ── logs ──────────────────────────────────────────────────────
cmd_logs() {
  init_compose
  local service="${2:-}"
  if [ -n "$service" ]; then
    $COMPOSE logs -f --tail=100 "$service"
  else
    $COMPOSE logs -f --tail=50
  fi
}

# ── migrate ───────────────────────────────────────────────────
cmd_migrate() {
  log_section "Running DB Migrations"
  init_compose
  check_env
  local root_pw; root_pw=$(grep MYSQL_ROOT_PASSWORD .env | cut -d= -f2)
  local db_name; db_name=$(grep MYSQL_DATABASE .env | cut -d= -f2)
  local applied=0
  for f in mysql/init/*.sql; do
    log_info "Applying $f ..."
    $COMPOSE exec -T mysql mysql -u root "-p${root_pw}" "${db_name}" < "$f" 2>/dev/null && \
      log_info "  ✓ $f" || log_warn "  ⚠ $f (skipped or already applied)"
    applied=$((applied+1))
  done
  log_info "✓ $applied file(s) processed."
}

# ── fix-admin ─────────────────────────────────────────────────
cmd_fix_admin() {
  log_section "Fixing Super Admin Account"
  init_compose
  check_env
  local root_pw; root_pw=$(grep MYSQL_ROOT_PASSWORD .env | cut -d= -f2)
  local db_name; db_name=$(grep MYSQL_DATABASE .env | cut -d= -f2)
  log_info "Applying mysql/init/02_fix_admin.sql..."
  $COMPOSE exec -T mysql mysql -u root "-p${root_pw}" "${db_name}" \
    < mysql/init/02_fix_admin.sql
  log_info "✓ Done."
  log_info "  Email:    admin@trbbtw.com"
  log_info "  Password: Trbb@Super2024!"
}

# ── db ────────────────────────────────────────────────────────
cmd_db() {
  init_compose
  check_env
  local root_pw; root_pw=$(grep MYSQL_ROOT_PASSWORD .env | cut -d= -f2)
  local db_name; db_name=$(grep MYSQL_DATABASE .env | cut -d= -f2)
  local sql_file="${2:-}"
  if [ -n "$sql_file" ]; then
    log_info "Executing ${sql_file}..."
    $COMPOSE exec -T mysql mysql -u root "-p${root_pw}" "${db_name}" < "${sql_file}"
  else
    $COMPOSE exec mysql mysql -u root "-p${root_pw}" "${db_name}"
  fi
}

# ── rm ────────────────────────────────────────────────────────
# 相當於 docker compose down -v
# 刪除所有容器 + 所有 named volumes（資料庫、快取、物件儲存全部清除）
cmd_rm() {
  log_section "Remove All Services & Data"
  init_compose

  echo -e "${RED}${BOLD}⚠ 警告：此操作將刪除所有容器及資料（MySQL / Redis / MinIO），無法復原！${RESET}"
  echo -n "輸入 'yes' 確認刪除："
  read -r confirm
  if [ "$confirm" != "yes" ]; then
    log_info "已取消。"
    exit 0
  fi

  log_info "Stopping and removing containers + volumes..."
  $COMPOSE down -v --remove-orphans
  log_info "✓ All containers and volumes removed."
  log_warn "資料已清除。若要重新啟動請執行: ./service.sh init"
}

# ── strava-webhook ────────────────────────────────────────────
cmd_strava_webhook() {
  local action="${1:-subscribe}"
  check_env

  local client_id;     client_id=$(grep STRAVA_CLIENT_ID     backend/.env | cut -d= -f2)
  local client_secret; client_secret=$(grep STRAVA_CLIENT_SECRET backend/.env | cut -d= -f2)
  local verify_token;  verify_token=$(grep STRAVA_WEBHOOK_VERIFY_TOKEN backend/.env 2>/dev/null | cut -d= -f2)
  local callback_url;  callback_url=$(grep STRAVA_REDIRECT_URI backend/.env | cut -d= -f2 | sed 's|/strava/callback|/strava/webhook|')

  if [ -z "$verify_token" ]; then
    verify_token="trbb_strava_webhook_2024"
  fi

  if [ -z "$client_id" ] || [ -z "$client_secret" ]; then
    log_error "STRAVA_CLIENT_ID / STRAVA_CLIENT_SECRET 未設定，請先填入 backend/.env"
    exit 1
  fi

  case "$action" in
    subscribe)
      log_section "Strava Webhook — 訂閱"
      log_info "Callback URL : $callback_url"
      log_info "Verify Token : $verify_token"
      echo ""
      curl -s -X POST https://www.strava.com/api/v3/push_subscriptions \
        -F "client_id=$client_id" \
        -F "client_secret=$client_secret" \
        -F "callback_url=$callback_url" \
        -F "verify_token=$verify_token" | python3 -m json.tool 2>/dev/null || echo "(raw response above)"
      echo ""
      log_info "若看到 {\"id\": ...} 表示訂閱成功，Strava 將自動推送活動事件。"
      ;;

    list)
      log_section "Strava Webhook — 查詢現有訂閱"
      curl -s -G https://www.strava.com/api/v3/push_subscriptions \
        -d "client_id=$client_id" \
        -d "client_secret=$client_secret" | python3 -m json.tool 2>/dev/null
      ;;

    delete)
      local sub_id="$2"
      if [ -z "$sub_id" ]; then
        log_error "請提供 subscription_id，用法：./service.sh strava-webhook delete <id>"
        exit 1
      fi
      log_section "Strava Webhook — 刪除訂閱 $sub_id"
      curl -s -X DELETE "https://www.strava.com/api/v3/push_subscriptions/$sub_id" \
        -F "client_id=$client_id" \
        -F "client_secret=$client_secret"
      echo ""
      log_info "訂閱 $sub_id 已刪除。"
      ;;

    *)
      log_error "未知動作：$action"
      echo "用法："
      echo "  ./service.sh strava-webhook subscribe   # 建立訂閱"
      echo "  ./service.sh strava-webhook list        # 查詢訂閱"
      echo "  ./service.sh strava-webhook delete <id> # 刪除訂閱"
      exit 1
      ;;
  esac
}


cmd_help() {
  echo -e "${BOLD}TRBB Service Manager${RESET}"
  echo ""
  echo -e "Usage: ${CYAN}./service.sh${RESET} [command] [options]"
  echo ""
  echo -e "${BOLD}Commands:${RESET}"
  echo -e "  ${GREEN}init${RESET}                        First-time setup"
  echo -e "  ${GREEN}build${RESET}                       Build all frontends + backend"
  echo -e "  ${GREEN}build-app${RESET} [web|app|admin]   Build a single frontend"
  echo -e "  ${GREEN}build-backend${RESET}               Rebuild Go backend only"
  echo -e "  ${GREEN}start${RESET}                       Start all services"
  echo -e "  ${GREEN}stop${RESET}                        Stop all services"
  echo -e "  ${GREEN}restart${RESET}                     Restart all services"
  echo -e "  ${GREEN}rebuild${RESET}                     Full rebuild + restart"
  echo -e "  ${GREEN}status${RESET}                      Show container status"
  echo -e "  ${GREEN}logs [service]${RESET}              Tail logs"
  echo -e "  ${GREEN}migrate${RESET}                     Apply all SQL migrations"
  echo -e "  ${GREEN}fix-admin${RESET}                   Reset super admin password"
  echo -e "  ${GREEN}db [file.sql]${RESET}               MySQL shell or run SQL file"
  echo -e "  ${GREEN}strava-webhook${RESET} [subscribe|list|delete <id>]  Manage Strava Webhook"
  echo -e "  ${RED}rm${RESET}                          ${RED}Remove ALL containers + data (irreversible)${RESET}"
  echo ""
  echo -e "${BOLD}Services:${RESET} mysql, redis, minio, backend, nginx"
  echo ""
  echo -e "${BOLD}Examples:${RESET}"
  echo -e "  ./service.sh init"
  echo -e "  ./service.sh build-app admin"
  echo -e "  ./service.sh logs backend"
  echo -e "  ./service.sh rm           ${RED}# wipes all data${RESET}"
}

_print_urls() {
  echo -e "\n${BOLD}Service URLs:${RESET}"
  echo -e "  Web Frontend  : http://trbbtw.com  (or http://localhost)"
  echo -e "  Mobile        : http://m.trbbtw.com"
  echo -e "  Admin         : http://admin.trbbtw.com"
  echo -e "  Images CDN    : http://images.trbbtw.com"
  echo -e "  MinIO Console : http://localhost:9001"
  echo -e "  Backend API   : http://localhost:8080/health\n"
}

# ── Router ────────────────────────────────────────────────────
COMMAND="${1:-help}"
case "$COMMAND" in
  init)           cmd_init          ;;
  build)          cmd_build         ;;
  build-app)      cmd_build_app     "$@" ;;
  build-backend)  cmd_build_backend ;;
  start)          cmd_start         ;;
  stop)           cmd_stop          ;;
  restart)        cmd_restart       ;;
  rebuild)        cmd_rebuild       ;;
  status)         cmd_status        ;;
  logs)           cmd_logs          "$@" ;;
  migrate)        cmd_migrate       ;;
  fix-admin)      cmd_fix_admin     ;;
  db)             cmd_db            "$@" ;;
  strava-webhook) shift; cmd_strava_webhook "$@" ;;
  rm)             cmd_rm            ;;
  help|--help|-h) cmd_help          ;;
  *)
    log_error "Unknown command: $COMMAND"
    echo ""
    cmd_help
    exit 1
    ;;
esac
