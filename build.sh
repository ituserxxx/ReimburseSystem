#!/usr/bin/env bash
# 一键打包脚本（Linux / macOS）
# 步骤：构建前端 -> 拷贝 dist 到后端 -> 编译带内嵌前端的单文件后端
set -e

ROOT="$(cd "$(dirname "$0")" && pwd)"
FRONTEND="$ROOT/frontend"
BACKEND="$ROOT/backend"
DIST="$BACKEND/web/dist"

echo "==> 1. 构建前端..."
cd "$FRONTEND"
[ -d node_modules ] || npm install
npm run build
cd "$ROOT"

echo "==> 2. 拷贝 frontend/dist -> backend/web/dist"
rm -rf "$DIST"
cp -r "$FRONTEND/dist" "$DIST"

echo "==> 3. 编译后端（内嵌前端，-tags embed）..."
cd "$BACKEND"
go build -tags embed -o "$ROOT/reimburse-system" .
cd "$ROOT"

echo "完成！运行 ./reimburse-system，浏览器访问 http://localhost:8080"
