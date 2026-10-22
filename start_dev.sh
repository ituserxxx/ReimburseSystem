#!/usr/bin/env bash
# 报销系统开发环境管理脚本：start / stop / status
#
# 端口约定（前后端共用 SERVER_PORT，保持一致）：
#   后端  -> backend/main.go 监听 SERVER_PORT
#   前端  -> frontend/vite.config.js 的 /api 代理目标指向 SERVER_PORT
#   前端开发服务器自身端口固定为 5173（浏览器访问地址）
#
# SERVER_PORT 取值优先级：shell 环境变量 > backend/.env > 8080（与 main.go 默认值一致）
# 脚本会把解析出的 SERVER_PORT 导出给前后端子进程，确保两端永远一致。
#
# 用法：
#   ./start_dev.sh start      # 后台启动前后端
#   ./start_dev.sh stop       # 停止前后端（按 PID 整树清理 + 按端口兜底清理）
#   ./start_dev.sh restart    # 先 stop 再 start
#   ./start_dev.sh status     # 查看状态（PID + 端口监听情况）
#   ./start_dev.sh logs       # 实时跟踪前后端日志（Ctrl+C 退出跟踪，不影响服务）
#   ./start_dev.sh help       # 查看帮助

set -u

# ------------------------- 基础配置 -------------------------

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

case "$(uname -s)" in
  *MINGW* | *MSYS* | *CYGWIN*) IS_WINDOWS=1 ;;
  *) IS_WINDOWS=0 ;;
esac

PID_FILE="$ROOT/.dev_pids"
BACKEND_LOG="$ROOT/.dev_backend.log"
FRONTEND_LOG="$ROOT/.dev_frontend.log"
FRONTEND_PORT=5173

# 从 backend/.env 解析 SERVER_PORT（去引号/空白）
read_env_port() {
  local f="$ROOT/backend/.env"
  [ -f "$f" ] || return 0
  grep -E '^[[:space:]]*SERVER_PORT[[:space:]]*=' "$f" 2>/dev/null \
    | head -n1 \
    | sed -E 's/^[^=]*=[[:space:]]*//; s/[[:space:]]*$//; s/^["'"'"']//; s/["'"'"']$//'
}

# 解析最终 SERVER_PORT：shell 环境变量优先，其次 backend/.env，最后默认 8080
if [ -z "${SERVER_PORT:-}" ]; then
  SERVER_PORT="$(read_env_port)"
  [ -z "$SERVER_PORT" ] && SERVER_PORT=8080
fi
export SERVER_PORT

# ------------------------- 工具函数 -------------------------

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "❌ 未找到命令: $1，请先在 PATH 中安装后再运行本脚本。"
    exit 1
  fi
}

pid_alive() {
  local pid=$1
  [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null
}

# 从 PID 文件安全读取某个值（只接受纯数字，忽略被污染/异常的行，绝不 source）
read_pid() {
  local key=$1
  [ -f "$PID_FILE" ] || { echo ""; return 0; }
  grep -E "^${key}=[0-9]+$" "$PID_FILE" 2>/dev/null | tail -n1 | cut -d= -f2
}

port_listening() {
  local port=$1
  (echo > "/dev/tcp/localhost/$port") >/dev/null 2>&1
}

wait_for_port() {
  local port=$1 tries=${2:-60} i=0
  while [ "$i" -lt "$tries" ]; do
    port_listening "$port" && return 0
    sleep 1
    i=$((i + 1))
  done
  return 1
}

kill_tree() {
  local pid=$1
  [ -z "$pid" ] && return 0
  if [ "$IS_WINDOWS" = "1" ]; then
    taskkill //F //T //PID "$pid" >/dev/null 2>&1 || true
  else
    # setsid 使服务自成进程组（PGID=PID），先杀整组（覆盖 go run 的临时编译进程、
    # npm 的 vite/esbuild 子进程），再杀单进程及其直接子进程兜底
    kill -- -"$pid" 2>/dev/null || true
    kill "$pid" 2>/dev/null || true
    pkill -P "$pid" 2>/dev/null || true
  fi
}

# 按端口兜底清理：杀掉所有监听该端口的进程
kill_port() {
  local port=$1
  [ -z "$port" ] && return 0
  if [ "$IS_WINDOWS" = "1" ]; then
    local pids
    pids=$(netstat -ano 2>/dev/null | awk -v p=":$port" '$1 ~ /TCP/ && $2 ~ p && $4=="LISTENING"{print $5}')
    for p in $pids; do taskkill //F //PID "$p" >/dev/null 2>&1 || true; done
  else
    if command -v fuser >/dev/null 2>&1; then
      fuser -k "${port}/tcp" >/dev/null 2>&1 || true
    fi
    if command -v lsof >/dev/null 2>&1; then
      lsof -ti "tcp:${port}" 2>/dev/null | xargs -r kill >/dev/null 2>&1 || true
    elif command -v ss >/dev/null 2>&1; then
      # 从 ss 输出解析监听该端口的进程 PID
      ss -ltnpH "sport = :$port" 2>/dev/null \
        | grep -oE 'pid=[0-9]+' | cut -d= -f2 | sort -u \
        | while read -r p; do kill "$p" 2>/dev/null || true; done
    fi
  fi
}

# 后台启动一个进程，返回其 PID（纯数字）。
# 日志创建失败时降级到 /dev/null，保证服务照常启动；
# setsid 让进程脱离脚本进程组（Ctrl+C 不误杀后台服务）。
launch() {
  local dir=$1 cmd=$2 log=$3
  mkdir -p "$(dirname "$log")" 2>/dev/null
  [ -d "$log" ] && rm -rf "$log" 2>/dev/null
  # 尝试创建/截断日志；失败则删除后重试一次；仍失败则降级到 /dev/null
  if ! : > "$log" 2>/dev/null; then
    rm -f "$log" 2>/dev/null
    if ! : > "$log" 2>/dev/null; then
      echo "⚠️  日志无法写入（可能被占用），该服务日志将输出到 /dev/null: $log" >&2
      log=/dev/null
    fi
  fi
  if command -v setsid >/dev/null 2>&1; then
    setsid bash -c "cd '$dir' && exec $cmd" >> "$log" 2>&1 </dev/null &
  else
    ( cd "$dir" && exec $cmd ) >> "$log" 2>&1 </dev/null &
  fi
  echo $!
}

# ------------------------- 子命令 -------------------------

do_start() {
  need_cmd go
  need_cmd npm

  # 1) 已有服务在跑 -> 拒绝重复启动
  local bp fp
  bp=$(read_pid BACKEND_PID)
  fp=$(read_pid FRONTEND_PID)
  if pid_alive "$bp" || pid_alive "$fp"; then
    echo "⚠️  前后端已在运行（PID 文件有效）。先执行 ./start_dev.sh stop 再启动。"
    exit 1
  fi
  rm -f "$PID_FILE"

  # 2) 清理可能残留的旧进程/端口占用（孤儿服务），避免占用端口或持有日志文件锁
  kill_port "$SERVER_PORT"
  kill_port "$FRONTEND_PORT"
  sleep 1

  # 3) 清理后端口仍被占用 -> 说明被外部程序占用，提示用户
  if port_listening "$SERVER_PORT"; then
    echo "⚠️  端口 $SERVER_PORT 仍被占用（非本脚本管理的进程），请手动释放后重试。"
    exit 1
  fi

  # 4) 前端依赖自检：缺失则自动安装
  if [ ! -d "$ROOT/frontend/node_modules" ]; then
    echo "ℹ️  前端依赖缺失，正在自动安装 (npm install) ..."
    ( cd "$ROOT/frontend" && npm install ) > "$FRONTEND_LOG.install" 2>&1
    if [ $? -ne 0 ]; then
      echo "❌ 前端依赖安装失败，请检查网络或手动执行: cd frontend && npm install"
      exit 1
    fi
    echo "✅ 前端依赖安装完成。"
  fi

  echo "=============================================="
  echo " 报销系统开发环境启动 (后台)"
  echo " SERVER_PORT : $SERVER_PORT  (后端监听 + 前端 /api 代理目标)"
  echo " 前端访问    : http://localhost:$FRONTEND_PORT"
  echo " 日志        : $BACKEND_LOG / $FRONTEND_LOG"
  echo "=============================================="

  # 5) 启动前后端
  BACKEND_PID=$(launch "$ROOT/backend" "go run main.go" "$BACKEND_LOG")
  FRONTEND_PID=$(launch "$ROOT/frontend" "npm run dev" "$FRONTEND_LOG")

  # 校验 PID 为纯数字（防止启动失败时把错误信息写进 PID 文件）
  if ! [[ "$BACKEND_PID" =~ ^[0-9]+$ ]]; then
    echo "❌ 后端启动失败。"
    exit 1
  fi
  if ! [[ "$FRONTEND_PID" =~ ^[0-9]+$ ]]; then
    echo "❌ 前端启动失败。"
    kill_tree "$BACKEND_PID"
    exit 1
  fi

  {
    echo "BACKEND_PID=$BACKEND_PID"
    echo "FRONTEND_PID=$FRONTEND_PID"
    echo "SERVER_PORT=$SERVER_PORT"
  } > "$PID_FILE"

  echo "已后台启动。后端 PID=$BACKEND_PID 前端 PID=$FRONTEND_PID"
  echo "等待服务就绪 ..."

  if wait_for_port "$SERVER_PORT" 60; then
    echo "✅ 后端已就绪 (端口 $SERVER_PORT)"
  else
    echo "⚠️  后端在 60s 内未就绪，请查看日志: $BACKEND_LOG"
  fi

  if wait_for_port "$FRONTEND_PORT" 60; then
    echo "✅ 前端已就绪: http://localhost:$FRONTEND_PORT"
  else
    echo "⚠️  前端在 60s 内未就绪，请查看日志: $FRONTEND_LOG"
  fi

  echo "----------------------------------------------"
  echo "停止请执行 : ./start_dev.sh stop"
  echo "查看状态   : ./start_dev.sh status"
  echo "查看日志   : ./start_dev.sh logs"
}

do_stop() {
  local bp fp
  bp=$(read_pid BACKEND_PID)
  fp=$(read_pid FRONTEND_PID)

  echo "正在停止前后端 ..."
  [ -n "$bp" ] && kill_tree "$bp"
  [ -n "$fp" ] && kill_tree "$fp"

  local i=0
  while [ "$i" -lt 10 ]; do
    if ! pid_alive "$bp" && ! pid_alive "$fp"; then
      break
    fi
    sleep 1
    i=$((i + 1))
  done

  # 按端口兜底清理，防止孤儿服务占用
  kill_port "${SERVER_PORT:-$(read_env_port)}"
  kill_port "$FRONTEND_PORT"

  rm -f "$PID_FILE"
  echo "✅ 已停止。"
}

do_status() {
  local bp fp
  bp=$(read_pid BACKEND_PID)
  fp=$(read_pid FRONTEND_PID)

  local b f
  if port_listening "$SERVER_PORT" || pid_alive "$bp"; then b="运行中"; else b="已停止"; fi
  if port_listening "$FRONTEND_PORT" || pid_alive "$fp"; then f="运行中"; else f="已停止"; fi

  echo "状态:"
  echo "  后端 : $b  (PID=${bp:-无}  监听端口 SERVER_PORT=$SERVER_PORT)"
  echo "  前端 : $f  (PID=${fp:-无}  开发服务器 $FRONTEND_PORT，/api 代理到 SERVER_PORT=$SERVER_PORT)"
  echo "  访问 : http://localhost:$FRONTEND_PORT"
}

do_logs() {
  echo "实时跟踪日志 (Ctrl+C 退出跟踪，不影响服务)"
  echo "===== 后端 (末尾 20 行) ====="
  tail -n 20 "$BACKEND_LOG" 2>/dev/null
  echo "===== 前端 (末尾 20 行) ====="
  tail -n 20 "$FRONTEND_LOG" 2>/dev/null
  echo "===== 实时 ====="
  tail -f "$BACKEND_LOG" "$FRONTEND_LOG" 2>/dev/null
}

usage() {
  sed -n '2,18p' "$0" | sed 's/^# \{0,1\}//'
}

# ------------------------- 入口分发 -------------------------

case "${1:-}" in
  start)   do_start ;;
  stop)    do_stop ;;
  restart) do_stop; do_start ;;
  status)  do_status ;;
  logs)    do_logs ;;
  help|-h|--help) usage ;;
  *)
    echo "用法: $0 {start|stop|restart|status|logs|help}"
    echo "详见: $0 help"
    exit 1
    ;;
esac
