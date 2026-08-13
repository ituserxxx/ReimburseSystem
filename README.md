# 报销记录管理系统

基于 **Vue 3 + Gin + GORM + MySQL** 实现的报销记录管理系统。

## 功能概览

- **报销记录管理**：新增 / 编辑 / 删除 / 查看报销记录
- **多人报销**：支持多人报销，每人独立金额分配和审批单号
- **报销缘由**：支持多项报销缘由（交通 / ETC / 油费 / 误餐 / 住宿 / 办公用品 / 其他，可自定义新增）
- **状态管理**：整体状态（待报销 / 已到账）+ 单人状态独立管理
- **统计面板**：总记录数、待报销笔数 / 金额、已到账金额一览
- **筛选查询**：按报销人、状态、报销类型、日期范围筛选
- **人员自动补全**：历史人员姓名自动提示

## 技术栈

| 层     | 技术                                      |
| ------ | ----------------------------------------- |
| 前端   | Vue 3 + Element Plus + Vite + Axios       |
| 后端   | Go + Gin + GORM                           |
| 数据库 | MySQL 5.7+ / 8.0                          |

## 项目结构

```
ReimburseSystem/
├── backend/                  # 后端 (Go/Gin)
│   ├── config/
│   │   └── database.go       # 数据库配置
│   ├── models/
│   │   └── reimbursement.go  # 数据模型定义
│   ├── handlers/
│   │   └── reimbursement.go  # API 处理器
│   ├── routers/
│   │   └── router.go         # 路由配置
│   ├── web/                  # 前端内嵌（通过 embed 构建标签）
│   │   ├── embed.go          # 启用 -tags embed 时：内嵌 frontend/dist 并提供 SPA 静态服务
│   │   └── static_off.go     # 未启用 embed 时：空实现（开发模式）
│   ├── init.sql              # 数据库初始化脚本
│   ├── main.go               # 入口文件
│   └── go.mod
├── frontend/                 # 前端 (Vue3)
│   ├── src/
│   │   ├── api/
│   │   │   └── index.js      # API 封装
│   │   ├── components/
│   │   │   ├── RecordList.vue    # 记录列表
│   │   │   ├── RecordForm.vue    # 新增/编辑表单
│   │   │   └── RecordDetail.vue  # 详情弹窗
│   │   ├── App.vue           # 主布局
│   │   └── main.js           # 入口
│   ├── index.html
│   ├── vite.config.js
│   └── package.json
└── README.md
```

## 快速开始

### 1. 准备数据库

> **无需手动建库建表。** 后端启动时，`config.InitDB()` 会自动 `CREATE DATABASE IF NOT EXISTS`（字符集 utf8mb4），随后 GORM AutoMigrate 自动创建三张表。只要 MySQL 服务可达、账号密码正确即可直接运行。

如需手动初始化（例如想先建好库/表再跑），可执行 `backend/init.sql`：

```sql
SOURCE backend/init.sql;
```

> 手动方式同样无需先 `CREATE DATABASE`，脚本内已包含建库语句。

### 2. 启动后端

```bash
cd backend

# 设置数据库连接（可选，默认值如下）
# Windows PowerShell:
$env:DB_HOST="127.0.0.1"
$env:DB_PORT="3306"
$env:DB_USER="root"
$env:DB_PASSWORD="your_password"
$env:DB_NAME="reimburse_system"
$env:SERVER_PORT="8080"

# Linux/Mac:
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=reimburse_system
export SERVER_PORT=8080

# 运行
go run main.go
```

后端默认运行在 `http://localhost:8080`。

### 3. 启动前端（开发模式，可选）

前后端分离开发时，前端可用 dev server 独立运行（已配置 `/api` 代理转发到后端 8080）：

```bash
cd frontend
npm install
npm run dev
```

前端默认运行在 `http://localhost:5173`。开发模式下后端无需内嵌前端（用 `go run main.go` 即可，仅提供 API）。

### 4. 打包为单文件（生产 / 一键部署）

使用仓库根目录的打包脚本，自动完成「构建前端 → 内嵌到后端 → 编译单文件」：

```powershell
# Windows (PowerShell)
.\build.ps1
```

```bash
# Linux / macOS
./build.sh
```

脚本会生成 **单文件可执行程序** `reimburse-system(.exe)`，把前端静态资源一并编译进二进制。直接运行即可，无需再启动独立的前端服务：

```bash
# Windows
.\reimburse-system.exe
# Linux / macOS
./reimburse-system
```

然后浏览器访问 **http://localhost:8080** 即可同时使用前端界面与后端 API。

> 原理：打包时使用 Go build tag `-tags embed`，由 `backend/web/embed.go` 将 `frontend/dist` 通过 `//go:embed` 嵌入二进制，并在 `router.go` 中注册 SPA 静态服务（含 `index.html` 回退）。
> 若不指定 `-tags embed`，则 `web.Register` 为空实现，方便开发时只用后端 API。

**配置文件（可选）**：打包后的程序会优先读取**可执行文件同目录**下的 `.env`。可复制仓库根目录的 `.env.example` 为 `.env` 并修改数据库连接等参数：

```bash
cp .env.example .env   # 与本程序放在同一目录
```

可配置项同 `.env.example`：`DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME / SERVER_PORT / GIN_MODE`。不配置则使用代码内置默认值（root/root@127.0.0.1:3306/reimburse_system，端口 8080）。

## 环境变量配置

| 变量名       | 默认值              | 说明           |
| ------------ | ------------------- | -------------- |
| DB_HOST      | 127.0.0.1           | 数据库地址     |
| DB_PORT      | 3306                | 数据库端口     |
| DB_USER      | root                | 数据库用户名   |
| DB_PASSWORD  | root                | 数据库密码     |
| DB_NAME      | reimburse_system    | 数据库名       |
| SERVER_PORT  | 8080                | 后端服务端口   |

## API 接口

| 方法   | 路径                                  | 说明                 |
| ------ | ------------------------------------- | -------------------- |
| GET    | /api/reimbursements                   | 获取报销记录列表     |
| GET    | /api/reimbursements/:id               | 获取报销记录详情     |
| POST   | /api/reimbursements                   | 创建报销记录         |
| PUT    | /api/reimbursements/:id               | 更新报销记录         |
| DELETE | /api/reimbursements/:id               | 删除报销记录         |
| PATCH  | /api/reimbursements/:id/status        | 更新整体状态         |
| PATCH  | /api/reimbursements/persons/:person_id/status | 更新单人状态  |
| GET    | /api/persons/names                    | 获取人员姓名列表     |
| GET    | /api/reasons/options                  | 获取报销缘由选项     |
| GET    | /api/statistics                       | 获取统计数据         |

## 数据模型

```
Reimbursement (报销记录主表)
├── submitter        报销人
├── is_multi_person  是否多人
├── total_amount     总金额
├── submit_date      提交日期
├── status           状态 (pending/received)
├── remark           备注
├── Persons[]        人员分配 (1:N)
│   ├── person_name      姓名
│   ├── amount           分配金额
│   ├── approval_number  审批单号
│   └── status           个人状态
└── Reasons[]        报销缘由 (1:N)
    └── reason           缘由
```
