没问题！这是加上了联系方式的**最终完整版 `README.md`**。

你可以直接复制粘贴，保存为 `README.md` 文件即可。

```markdown
# 🧠 个人知识护城河构建系统 (Knowledge & Question Bank System)

![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)
![Gin](https://img.shields.io/badge/Gin-Framework-008ECF?style=flat&logo=go)
![SQLite](https://img.shields.io/badge/SQLite-Embedded-003B57?style=flat&logo=sqlite)
![Vue3](https://img.shields.io/badge/Vue-3.0+-4FC08D?style=flat&logo=vue.js)
![ElementPlus](https://img.shields.io/badge/Element--Plus-UI-409EFF?style=flat&logo=element)

## 📖 项目简介

**告别低效勤奋，重新定义你的学习方式。**

在这个信息爆炸、知识碎片化的时代，你是否也陷入了这样的**“学习焦虑”**：
*   🤯 **收藏夹吃灰**：看到好文章就收藏，但永远找不到，知识永远零散地躺在各个角落。
*   😫 **题海战术失效**：想刷题巩固，却需要在各大网站像无头苍蝇一样乱撞，题目质量参差不齐，且与你学的知识点完全脱节。
*   📉 **学了就忘**：理论看了一大堆，没有对应的实战练习，两天就忘得一干二净。

**STOP！是时候结束这种低效的循环了。**

本项目不是一个简单的管理系统，它是为你量身打造的 **“个人知识护城河构建平台”**，是你对抗遗忘、征服复杂技术栈的**终极武器**！

它完美解决了**“学”与“练”割裂**的历史难题。在这里，每一滴知识都有它的归宿，每一道题目都有它的根基。我们通过 **"科目 -> 分类 -> 知识点 -> 题目"** 的严密逻辑，帮你把脑海中一团乱麻的知识线索，编织成一张坚不可摧的**知识图谱**。

**它能为你带来什么？**
*   💎 **打造专属第二大脑**：把分散在全网的精华收录进来，构建完全属于你的私有知识库，谁也拿不走。
*   🚀 **实现降维打击**：看完知识点立马刷题，理论与实践毫秒级闭环，让你的学习效率产生复利效应。
*   🧘 **找回掌控感**：不再为“去哪找题”而发愁，不再为“学过但忘了”而焦虑。一切尽在掌握。

**这不只是一个工具，这是你技术进阶路上的加速引擎。**

---

## 🌟 核心特性

*   🚀 **极速启动 (Zero Config)**：后端底层架构已升级为 **SQLite**。无需安装笨重的 MySQL，无需配置复杂的数据库服务，下载即运行，数据随身带。
*   ⚡ **高性能架构**：基于 **Go + Gin** 研发，数据库开启 **WAL (Write-Ahead Logging)** 模式，配合 **FULL** 同步机制，兼顾极速响应与数据安全。
*   📚 **四层级知识体系**：严格遵循“科目-分类-知识点-题目”的结构，强制让你把知识梳理得井井有条。
*   📝 **所见即所得**：集成现代化富文本编辑器，支持代码高亮、图片本地存储与管理，让记录知识成为一种享受。
*   🛡️ **数据智能维护**：内置 SQLite 触发器 (Triggers)，自动维护数据创建与更新时间；开启外键约束，保证数据的完整性与逻辑严密性。

## 🛠️ 技术栈

### 后端 (Backend)
*   **语言**: Golang
*   **Web 框架**: [Gin](https://github.com/gin-gonic/gin)
*   **数据交互**: 原生 SQL 优化封装，拒绝臃肿的 ORM，追求极致性能
*   **文件服务**: 原生静态资源服务

### 前端 (Frontend)
*   **框架**: Vue 3 (Composition API)
*   **构建工具**: Vite
*   **UI 组件库**: Element Plus
*   **富文本**: WangEditor
*   **路由**: Vue Router 4

## 📂 目录结构

├── api/                  # 核心业务逻辑 (Controller)
│   ├── point.go          # 知识点管理 (CRUD、富文本处理)
│   ├── category.go       # 分类管理
│   ├── question.go       # 题目管理 (关联知识点)
│   └── subject.go        # 科目管理
├── common/               # 公共工具箱 (文件删除、路径处理)
├── global/               # 全局对象容器 (DB 连接池)
├── initialize/           # 系统初始化核心
│   ├── db.go             # SQLite 连接、WAL模式配置、表结构自动生成、数据迁移
│   └── router.go         # 路由注册与中间件
├── model/                # 数据模型定义 (Struct)
├── uploads/              # 持久化数据存储区
│   ├── data.db           # 你的核心资产 (SQLite 数据库文件)
│   ├── data.db-wal       # 预写日志 (运行时自动生成)
│   └── images/           # 知识点图片仓库
├── main.go               # 程序入口
├── go.mod                # 依赖管理
└── README.md             # 项目文档
```

## 💾 数据库设计 (Schema)

系统自动维护以下核心表结构 (启动时自动创建)：

1.  **subjects (科目)**: 知识的顶层容器（如：Java、架构设计）。
2.  **knowledge_categories (分类)**: 知识的细分领域（如：集合源码、分布式锁）。
3.  **knowledge_points (知识点)**: 核心内容载体，包含详细的富文本笔记。
4.  **questions (题目)**: 挂载在具体知识点下的练习题，支持图片选项和解析。

> **技术细节**: 系统利用 SQLite 的 `TRIGGER` (触发器) 完美复刻了 MySQL 的 `ON UPDATE CURRENT_TIMESTAMP` 功能，确保时间字段永远精准。

## 🚀 5分钟快速开始

### 1. 后端启动 (Server)

无需安装数据库软件，确保有 Go 环境即可。

```bash
# 1. 下载依赖
go mod tidy

# 2. 启动服务
# 首次启动时，系统会自动在 uploads 目录下生成 data.db 并初始化表结构
go run main.go
```

看到 `[GIN-debug] Listening and serving HTTP on :8080` 即代表启动成功。

### 2. 前端启动 (Client)

确保安装了 Node.js。

```bash
cd frontend

# 1. 安装依赖
npm install

# 2. 启动开发模式
npm run dev
```

访问浏览器显示的地址（通常是 `http://localhost:5173`），开始构建你的知识帝国！

## 🔌 API 接口速查

| 模块 | 方法 | 路径 | 说明 |
| :--- | :--- | :--- | :--- |
| **科目** | GET | `/api/v1/subjects` | 获取所有启用科目 |
| | POST | `/api/v1/subjects` | 新建科目 |
| **分类** | GET | `/api/v1/categories` | 获取分类 (支持 `?subject_id=x` 筛选) |
| **知识点** | GET | `/api/v1/points` | 获取知识点列表 (只返回 ID 和 标题，极速) |
| | GET | `/api/v1/points/:id` | 获取详情 (包含富文本内容) |
| **题目** | GET | `/api/v1/questions` | 获取题目 (支持按知识点或分类筛选) |

## 💡 常见问题 (FAQ)

**Q: 数据库文件在哪里？**
A: 在项目根目录的 `uploads/data.db`。你可以直接复制这个文件进行备份，或者发给其他电脑使用，真正的数据随身带。

**Q: 我可以用 Navicat 查看数据吗？**
A: 当然可以！Navicat Premium 或 DBeaver 等工具均支持直接打开 `.db` 文件进行查看和编辑。

**Q: 为什么文件夹里多了 .wal 和 .shm 文件？**
A: 这是 SQLite 的 **WAL (Write-Ahead Logging)** 机制生成的临时文件，用于提升并发性能。**请勿手动删除**，程序关闭后它们会自动合并。

## 📧 联系方式

如果你对项目有任何疑问，或者想交流技术心得，欢迎通过邮件联系我：

📩 **yusongsong993@gmail.com**

## 🤝 贡献与支持

如果你觉得这个项目帮到了你，请给一个 ⭐️ **Star**！
欢迎提交 Issues 反馈 bug，或提交 PR 贡献代码。

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。
```