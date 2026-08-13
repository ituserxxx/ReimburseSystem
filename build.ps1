# 一键打包脚本（Windows / PowerShell）
# 步骤：构建前端 -> 拷贝 dist 到后端 -> 编译带内嵌前端的单文件后端
$ErrorActionPreference = "Stop"

$root     = Split-Path -Parent $MyInvocation.MyCommand.Definition
$frontend = Join-Path $root "frontend"
$backend  = Join-Path $root "backend"
$dist     = Join-Path $backend "web" "dist"

Write-Host "==> 1. 构建前端..." -ForegroundColor Cyan
Push-Location $frontend
if (-not (Test-Path "node_modules")) { npm install }
npm run build
Pop-Location

Write-Host "==> 2. 拷贝 frontend/dist -> backend/web/dist" -ForegroundColor Cyan
if (Test-Path $dist) { Remove-Item -Recurse -Force $dist }
Copy-Item -Recurse (Join-Path $frontend "dist") $dist

Write-Host "==> 3. 编译后端（内嵌前端，-tags embed）..." -ForegroundColor Cyan
Push-Location $backend
go build -tags embed -o (Join-Path $root "reimburse-system.exe") .
Pop-Location

Write-Host "完成！运行 reimburse-system.exe，浏览器访问 http://localhost:8080" -ForegroundColor Green
