# ⚡ Practice Problems - 极简全栈题库刷题系统

<p align="center">
  <img src="https://img.shields.io/badge/Backend-Go%201.20%2B-blue" alt="Go">
  <img src="https://img.shields.io/badge/Frontend-Vue%203%20%2B%20TS-green" alt="Vue">
  <img src="https://img.shields.io/badge/Database-SQLite-lightgrey" alt="SQLite">
  <img src="https://img.shields.io/badge/License-MIT-orange" alt="License">
</p>

> 一个轻量级、高性能、易于部署的个人知识库与刷题系统。支持私有化部署，也提供在线服务。

## 🔗 在线体验 (Live Demo)

👉 **访问地址：[https://pp.yugams.com](https://pp.yugams.com)**

*   **账号**：可直接注册新账号体验
*   **特性**：全站 HTTPS 加密，数据安全，访问极速。

---

## ✨ 项目亮点 (Features)

### 核心功能
- **📚 多层级知识管理**：支持 `科目` -> `分类` -> `知识点` -> `题目` 的四层结构管理。
- **🔐 灵活的权限控制**：
    - **创建者权限**：只有作者可以修改、删除自己的题库。
    - **分享码机制**：独创的 **Share Code** 系统，支持生成有效期（如 3天、永久）的分享码，一键分享给他人订阅。
    - **订阅者权限**：通过分享码绑定的用户，拥有只读刷题权限。
- **📢 公告系统**：支持针对分享码发布特定公告。
- **🖼️ 图片管理**：支持知识点/题目图片上传，自动压缩与本地存储。

### 技术特性
- **🚀 纯 Go 后端**：基于 Gin 框架，高性能 API 服务。
- **📂 单文件即可运行，数据迁移极其方便。
- **⚡ Vue 3 前端**：使用 Vite + Element Plus + TypeScript 构建，极致的加载速度和丝滑的交互体验。
- **🛡️ 安全加固**：全站 JWT 鉴权，密码 Bcrypt 加密，支持 Nginx 反向代理与 HTTPS。

---

## 🛠️ 技术栈 (Tech Stack)

| 模块 | 技术选型 | 说明 |
| :--- | :--- | :--- |
| **后端** | Golang | 核心语言 |
| **Web框架** | Gin | 高性能 HTTP Web 框架 |
| **数据库** | SQLite | 轻量级嵌入式数据库 (无需安装 MySQL) |
| **日志** | Zap + Lumberjack | 高性能结构化日志 & 日志轮转 |
| **鉴权** | JWT (JSON Web Token) | 无状态认证 |
| **前端** | Vue 3 + TypeScript | 组合式 API 开发 |
| **UI库** | Element Plus | 饿了么 UI 组件库 |
| **构建工具** | Vite | 下一代前端构建工具 |

---

## 🚀 本地开发与运行 (Getting Started)

### 1. 环境准备
*   **Go** >= 1.18
*   **Node.js** >= 16
*   **Git**

### 2. 后端启动 (Backend)

```bash
# 1. 克隆项目
git clone https://github.com/YYMi/practice_problems.git
cd practice_problems

# 2. 下载依赖
go mod tidy

# 3. 运行项目
# 程序会自动在 /uploads 目录下初始化 data.db 数据库文件
go run main.go
后端默认监听端口：19527 (可在代码中修改)

3. 前端启动 (Frontend)
Copy# 进入前端目录
cd practice_problems_web

# 1. 安装依赖
npm install

# 2. 启动开发服务器
npm run dev
前端默认地址：http://localhost:19528

📦 服务器部署指南 (Deployment)
本项目支持跨平台编译，推荐部署在 Linux 服务器上。

1. 编译后端
在 Windows/Mac 上交叉编译 Linux 可执行文件：

Copy# Windows PowerShell
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o practiceProblem

# Mac/Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o practiceProblem
2. 打包前端
Copycd practice_problems_web
npm run build
# 生成的 dist 目录即为静态资源
3. 上传与运行
将 practiceProblem 文件和 dist 目录上传至服务器（例如 /dev/data/）。

使用 Systemd 管理服务 (推荐):

Copy# /etc/systemd/system/practice.service
[Unit]
Description=Practice Problem Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/dev/data/go-practiceProblem
ExecStart=/dev/data/go-practiceProblem/practiceProblem
Restart=always

[Install]
WantedBy=multi-user.target
4. Nginx 配置 (HTTPS + 反向代理)
推荐使用 Nginx 托管前端静态资源，并反向代理后端 API。

Copyserver {
    listen 443 ssl;
    server_name pp.yugams.com;

    # SSL 配置
    ssl_certificate /etc/nginx/cert/pp.pem;
    ssl_certificate_key /etc/nginx/cert/pp.key;

    # 前端静态资源
    location / {
        root /dev/data/vue-frontend/dist;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:19527;
        proxy_set_header Host $host;
    }

    # 图片资源代理
    location /uploads/ {
        alias /dev/data/go-practiceProblem/uploads/;
    }
}
📂 项目目录结构
Copypractice_problems/
├── api/                  # 业务逻辑 Controller
├── global/               # 全局变量 (DB, Log)
├── initialize/           # 初始化逻辑 (Logger, SQLite建表)
├── middleware/           # 中间件 (JWT, CORS, RequestID, Logger)
├── model/                # 结构体定义
├── router/               # 路由配置
├── uploads/              # 资源存储 (图片 + data.db)
├── practice_problems_web/ # 前端 Vue 项目
├── main.go               # 入口文件
└── README.md             # 说明文档
🤝 贡献 (Contributing)

```
如果你觉得这个项目对你有帮助，请给一个 ⭐️ Star！

## 📬 联系方式 (Contact)

如果你在使用过程中遇到问题，或者有功能建议，欢迎通过以下方式联系：

- 🐛 **技术支持**: [GitHub Issues](https://github.com/YYMi/practice_problems/issues)
- 📧 **Email**: [yusongsong1993@gmail.com](mailto:yusongsong1993@gmail.com)

📄 协议 (License)
MIT License © 2025 YuBaiBai


