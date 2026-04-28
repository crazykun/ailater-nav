# AI Later

<div align="center">

![AI Later Logo](static/img/logo.png)

# AI Later 创造美好生活

一个基于 **Go + Gin + MySQL** 的 AI 工具导航站，支持 **用户登录、收藏、后台管理**，并提供从历史 `data/ai.json` 到 MySQL 的一次性迁移能力。

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web-Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

</div>

---

## ✨ 功能概览

- 🔍 **AI 工具搜索与分类筛选**
- 👤 **用户注册 / 登录 / 退出**
- ⭐ **收藏功能与个人中心**
- 🛠️ **后台站点管理**
- 🗄️ **MySQL 持久化存储**
- 🚚 **旧版 `data/ai.json` 一次性迁移到 MySQL**
- 🎨 **Go Template + HTMX + Alpine.js 驱动的轻量交互界面**

## 🧰 技术栈

| 类别 | 方案 |
|---|---|
| 后端 | Go 1.24 + Gin |
| 模板 | Go HTML Templates |
| 前端 | Tailwind CSS + HTMX + Alpine.js |
| 数据库 | MySQL 8+ |
| 数据迁移 | Bash + Go 脚本 |
| 鉴权 | JWT |

## 📦 环境要求

- Go 1.24 或更高版本
- MySQL 8.0 或兼容版本

## 🚀 快速开始

### 1) 克隆项目

```bash
git clone https://github.com/crazykun/ai-later.git
cd ai-later
```

### 2) 准备配置文件

```bash
cp config.demo.yaml config.yaml
```

编辑 `config.yaml`，至少确认以下配置正确：

```yaml
port: 8080

mysql:
  host: localhost
  port: 3306
  username: root
  password: "your-password"
  database: ai_later

jwt:
  secret: your-jwt-secret-change-me
  expire_days: 7
```

### 3) 初始化 MySQL 并导入历史 JSON 数据

如果你是从旧版 JSON 数据迁移，直接运行：

```bash
bash scripts/init_mysql_from_json.sh
```

这个脚本会按顺序执行：

1. 检查 `go`、`config.yaml`、`data/ai.json`
2. 根据 `config.yaml` 创建数据库
3. 执行 `internal/database/migrations/` 目录下全部 SQL migration 文件
4. 读取 `data/ai.json` 并导入 MySQL

> 💡 迁移脚本具备幂等性：已存在的站点会按名称跳过，重复执行不会重复插入同名站点。

### 4) 启动服务

```bash
go run ./main.go
```

默认访问地址：

```text
http://localhost:8080
```

---

## 🧭 线上迁移流程

推荐线上初始化步骤：

```bash
cp config.demo.yaml config.yaml
# 修改 mysql / jwt / admin 配置
bash scripts/init_mysql_from_json.sh
go run ./main.go
```

### 迁移建议

- ✅ 首次迁移前先备份 `data/ai.json`
- ✅ 使用独立 MySQL 账号，不要在线上继续使用 root
- ✅ 将 `JWT_SECRET`、数据库密码改成强随机值
- ✅ 迁移完成后以 **MySQL 为准**，不要再把 JSON 作为主数据源

如果线上已经完成过迁移，之后只需要正常启动应用，不必重复执行 JSON 导入。

---

## 🗄️ 数据库初始化与迁移说明

### `scripts/init_mysql_from_json.sh`

一键初始化脚本，适用于首次从 JSON 版本切到 MySQL 版本。

执行内容：

```bash
go run ./scripts/create_db/main.go
go run ./scripts/migrate/main.go
```

### `scripts/create_db/main.go`

根据 `config.yaml` 中的 MySQL 配置创建数据库。

### `scripts/migrate/main.go`

负责两件事：

1. 执行 SQL migration
2. 将 `data/ai.json` 中的站点导入到 MySQL

当导入过程中存在失败记录时，脚本会返回非零退出状态，避免“部分成功但整体显示成功”的情况。

### 旧版 JSON 数据结构

```json
{
  "name": "站点名称",
  "url": "https://example.com",
  "description": "描述",
  "logo": "/static/img/logo.png",
  "tags": ["标签1", "标签2"],
  "category": "分类",
  "rating": 4.5,
  "featured": true
}
```

---

## 🧱 数据库结构

当前 migration SQL 位于：

- `internal/database/migrations/`

应用会按文件名排序执行该目录下所有 `*.sql` 文件。

核心表包括：

- `sites`
- `tags`
- `site_tags`
- `users`
- `favorites`
- `visits`

---

## 🧪 测试

运行全部测试：

```bash
go test ./...
```

模板相关回归测试位于：

- `test/templ_test.go`

---

## 📁 项目结构

```text
.
├── config.demo.yaml
├── config.yaml
├── data/
│   └── ai.json
├── internal/
│   ├── config/
│   ├── database/
│   │   ├── migrations/
│   │   └── repository/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── services/
│   ├── utils/
│   └── web/
├── scripts/
│   ├── create_db/
│   │   └── main.go
│   ├── migrate/
│   │   └── main.go
│   └── init_mysql_from_json.sh
├── static/
├── templates/
├── test/
│   └── templ_test.go
├── main.go
└── README.md
```

---

## 🔧 仅初始化数据库结构

如果你只想建库，不导入 JSON 数据：

```bash
go run ./scripts/create_db/main.go
go run ./main.go
```

`main.go` 启动时会自动执行数据库 migration。

---

## 📜 License

MIT License


---

<div align="center">

Made with ❤️ by [crazykun](https://github.com/crazykun)

</div>