# AI Later

<div align="center">

![AI Later Logo](static/img/logo.png)

一个基于 **Go + Gin + MySQL** 的 AI 工具导航站，支持用户登录、收藏、后台管理。

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web-Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql&logoColor=white)](https://www.mysql.com/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

</div>

## 功能

- AI 工具搜索与分类筛选
- 用户注册 / 登录 / 退出
- 收藏功能与个人中心
- 后台站点管理
- MySQL 持久化存储
- Go Template + HTMX + Alpine.js 轻量交互

## 技术栈

| 类别 | 方案 |
|---|---|
| 后端 | Go 1.24 + Gin |
| 模板 | Go HTML Templates |
| 前端 | Tailwind CSS + HTMX + Alpine.js |
| 数据库 | MySQL 8+ |
| 鉴权 | JWT |

## 快速开始

**环境要求**: Go 1.24+, MySQL 8.0+

```bash
# 1. 克隆项目
git clone https://github.com/crazykun/ai-later.git
cd ai-later

# 2. 配置
cp config.demo.yaml config.yaml
# 编辑 config.yaml，填入 MySQL 连接信息和 JWT secret

# 3. 初始化数据库
go run scripts/create_db/main.go

# 4. 启动服务
go run main.go
```

访问 http://localhost:8080

## 配置说明

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

## 数据库

migration SQL 位于 `internal/database/migrations/`，应用启动时自动执行。

核心表: `sites` / `tags` / `site_tags` / `users` / `favorites` / `visits`

## 项目结构

```
.
├── main.go                   # 入口
├── config.demo.yaml          # 配置模板
├── internal/
│   ├── config/               # 配置加载
│   ├── database/
│   │   ├── migrations/       # SQL 迁移文件
│   │   └── repository/       # 数据访问层
│   ├── handlers/             # HTTP 处理器
│   ├── middleware/            # 中间件 (JWT 鉴权等)
│   ├── models/               # 数据模型
│   ├── services/             # 业务逻辑
│   └── utils/                # 工具函数
├── scripts/
│   └── create_db/            # 建库脚本
├── static/                   # CSS / JS / 图片
└── templates/                # HTML 模板
```

## License

MIT

---

<div align="center">Made with ❤️ by <a href="https://github.com/crazykun">crazykun</a></div>
