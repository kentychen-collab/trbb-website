#!/usr/bin/env bash
# =============================================================
# SportsPlatform service.sh
# 用法: ./service.sh <command>
#
# Commands:
#   init      初始化：建立 .env、初始化 DB、啟動所有服務
#   start     啟動所有服務
#   stop      停止所有服務
#   restart   重啟所有服務（或指定服務）
#   build     重新編譯 backend 並重啟
#   status    顯示各服務狀態
#   logs      即時顯示 log（預設 backend，可指定服務）
#   db        進入 MySQL CLI
#   shell     進入指定服務的 shell
#   reset-db  ⚠️  清空並重建資料庫（會清除所有資料）
# =============================================================

set -euo pipefail

# ---------- 顏色輸出 ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC}  $*"; }
success() { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }
section() { echo -e "\n${BOLD}${CYAN}=== $* ===${NC}"; }

# ---------- 路徑設定 ----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

ENV_FILE="$SCRIPT_DIR/.env"
COMPOSE="docker-compose"

# 確認 docker-compose 可用
command -v docker-compose &>/dev/null || COMPOSE="docker compose"
command -v docker &>/dev/null || error "找不到 docker，請先安裝 Docker"

# ---------- 指令 ----------
CMD="${1:-help}"
shift || true
EXTRA_ARGS=("$@")

# =============================================================
case "$CMD" in

# -------------------------------------------------------------
init)
  section "初始化 SportsPlatform"

  # 1. 建立 .env
  if [[ ! -f "$ENV_FILE" ]]; then
    if [[ ! -f "$SCRIPT_DIR/.env.example" ]]; then
      error "找不到 .env.example，請確認專案完整"
    fi
    cp "$SCRIPT_DIR/.env.example" "$ENV_FILE"
    warn ".env 已從 .env.example 建立，請編輯填入正確設定後再繼續"
    echo ""
    echo -e "  ${YELLOW}vi $ENV_FILE${NC}"
    echo ""
    read -rp "  已編輯完成？按 Enter 繼續，Ctrl+C 取消... "
  else
    info ".env 已存在，跳過建立"
  fi

  # 2. 讀取 .env 確認必要變數
  source "$ENV_FILE"
  [[ -z "${DB_ROOT_PASS:-}" ]] && error "DB_ROOT_PASS 未設定，請編輯 .env"
  [[ -z "${DB_USER:-}"      ]] && error "DB_USER 未設定，請編輯 .env"
  [[ -z "${DB_PASS:-}"      ]] && error "DB_PASS 未設定，請編輯 .env"
  [[ -z "${JWT_SECRET:-}"   ]] && error "JWT_SECRET 未設定，請編輯 .env"
  [[ "${JWT_SECRET}" == "your-super-secret-jwt-key-change-this-in-production" ]] && \
    warn "JWT_SECRET 使用預設值，建議修改為隨機字串"
  [[ "${ADMIN_INIT_PASSWORD:-}" == "ChangeMe@2025" ]] && \
    warn "ADMIN_INIT_PASSWORD 使用預設值，建議修改"

  # 3. 編譯並啟動
  info "編譯 backend..."
  $COMPOSE build backend

  info "啟動所有服務..."
  $COMPOSE up -d

  # 4. 等待 MySQL 就緒
  info "等待 MySQL 就緒..."
  MAX_WAIT=60
  COUNT=0
  until $COMPOSE exec -T mysql mysqladmin ping -h localhost --silent 2>/dev/null; do
    COUNT=$((COUNT+1))
    [[ $COUNT -ge $MAX_WAIT ]] && error "MySQL 啟動超時（${MAX_WAIT}s）"
    sleep 1
    printf "."
  done
  echo ""
  success "MySQL 已就緒"

  # 5. 完成
  echo ""
  success "初始化完成！"
  echo ""
  echo -e "  前台：${CYAN}http://localhost${NC}"
  echo -e "  後台：${CYAN}http://localhost/admin${NC}  或  ${CYAN}http://admin.yourdomain.com${NC}"
  echo -e "  超管帳號：${YELLOW}${ADMIN_INIT_USERNAME:-superadmin}${NC}"
  echo ""
  ;;

# -------------------------------------------------------------
start)
  section "啟動服務"
  if [[ ! -f "$ENV_FILE" ]]; then
    error ".env 不存在，請先執行：./service.sh init"
  fi
  $COMPOSE up -d "${EXTRA_ARGS[@]}"
  success "服務已啟動"
  echo ""
  $COMPOSE ps
  ;;

# -------------------------------------------------------------
stop)
  section "停止服務"
  $COMPOSE stop "${EXTRA_ARGS[@]}"
  success "服務已停止"
  ;;

# -------------------------------------------------------------
restart)
  section "重啟服務"
  TARGET="${EXTRA_ARGS[0]:-}"
  if [[ -n "$TARGET" ]]; then
    info "重啟 $TARGET..."
    $COMPOSE restart "$TARGET"
  else
    info "重啟所有服務..."
    $COMPOSE restart
  fi
  success "重啟完成"
  $COMPOSE ps
  ;;

# -------------------------------------------------------------
build)
  section "重新編譯 backend"
  info "停止 backend..."
  $COMPOSE stop backend

  info "重新 build..."
  $COMPOSE build backend

  info "啟動 backend..."
  $COMPOSE up -d backend

  # 等待 backend 健康
  sleep 2
  if $COMPOSE ps backend | grep -q "Up"; then
    success "backend 編譯並重啟完成"
  else
    error "backend 啟動失敗，請查看 log：./service.sh logs backend"
  fi
  ;;

# -------------------------------------------------------------
status)
  section "服務狀態"
  $COMPOSE ps

  echo ""
  info "資源使用："
  docker stats --no-stream \
    $($COMPOSE ps -q 2>/dev/null | head -10) \
    --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null || true
  ;;

# -------------------------------------------------------------
logs)
  TARGET="${EXTRA_ARGS[0]:-backend}"
  LINES="${EXTRA_ARGS[1]:-100}"
  section "Log：$TARGET（最近 $LINES 行，Ctrl+C 結束）"
  $COMPOSE logs -f --tail="$LINES" "$TARGET"
  ;;

# -------------------------------------------------------------
db)
  section "進入 MySQL CLI"
  if [[ -f "$ENV_FILE" ]]; then
    source "$ENV_FILE"
  fi
  DB_NAME="${DB_NAME:-sports_platform}"
  DB_USER_VAL="${DB_USER:-root}"
  DB_PASS_VAL="${DB_PASS:-}"
  info "連線到資料庫：$DB_NAME"
  $COMPOSE exec mysql mysql -u"$DB_USER_VAL" -p"$DB_PASS_VAL" "$DB_NAME"
  ;;

# -------------------------------------------------------------
shell)
  TARGET="${EXTRA_ARGS[0]:-backend}"
  section "進入 $TARGET shell"
  case "$TARGET" in
    backend) $COMPOSE exec backend sh ;;
    mysql)   $COMPOSE exec mysql bash ;;
    nginx)   $COMPOSE exec nginx sh ;;
    redis)   $COMPOSE exec redis sh ;;
    *)       $COMPOSE exec "$TARGET" sh ;;
  esac
  ;;

# -------------------------------------------------------------
reset-db)
  section "⚠️  重建資料庫"
  echo -e "${RED}警告：此操作將清除所有資料庫資料！${NC}"
  echo ""
  read -rp "  確定要繼續嗎？輸入 'yes' 確認: " CONFIRM
  [[ "$CONFIRM" != "yes" ]] && { warn "已取消"; exit 0; }

  info "停止所有服務..."
  $COMPOSE down

  info "刪除 MySQL volume..."
  docker volume rm "$(basename "$SCRIPT_DIR")_mysql_data" 2>/dev/null || \
    docker volume rm "sports-platform_mysql_data" 2>/dev/null || \
    warn "找不到 volume，可能已被刪除"

  info "重新啟動（會自動執行 init.sql）..."
  $COMPOSE up -d

  info "等待 MySQL 就緒..."
  MAX_WAIT=60
  COUNT=0
  until $COMPOSE exec -T mysql mysqladmin ping -h localhost --silent 2>/dev/null; do
    COUNT=$((COUNT+1))
    [[ $COUNT -ge $MAX_WAIT ]] && error "MySQL 啟動超時"
    sleep 1
    printf "."
  done
  echo ""
  success "資料庫已重建完成"
  ;;

# -------------------------------------------------------------
help|--help|-h|*)
  echo ""
  echo -e "${BOLD}SportsPlatform service.sh${NC}"
  echo ""
  echo -e "  ${CYAN}./service.sh init${NC}              初始化：建立 .env、啟動所有服務"
  echo -e "  ${CYAN}./service.sh start${NC}             啟動所有服務"
  echo -e "  ${CYAN}./service.sh start nginx${NC}       啟動指定服務"
  echo -e "  ${CYAN}./service.sh stop${NC}              停止所有服務"
  echo -e "  ${CYAN}./service.sh stop backend${NC}      停止指定服務"
  echo -e "  ${CYAN}./service.sh restart${NC}           重啟所有服務"
  echo -e "  ${CYAN}./service.sh restart nginx${NC}     重啟指定服務"
  echo -e "  ${CYAN}./service.sh build${NC}             重新編譯 backend 並重啟"
  echo -e "  ${CYAN}./service.sh status${NC}            顯示各服務狀態與資源使用"
  echo -e "  ${CYAN}./service.sh logs${NC}              顯示 backend log（即時）"
  echo -e "  ${CYAN}./service.sh logs mysql 50${NC}     顯示 mysql 最近 50 行 log"
  echo -e "  ${CYAN}./service.sh db${NC}                進入 MySQL CLI"
  echo -e "  ${CYAN}./service.sh shell backend${NC}     進入指定服務的 shell"
  echo -e "  ${CYAN}./service.sh reset-db${NC}          ⚠️  清空並重建資料庫"
  echo ""
  ;;

esac
