# Backend Develop Guide Skill

这是一个 Agent Skill，用于指导 Go 后端开发工作。基于 Gin + PostgreSQL + Redis 的三层架构项目规范。

## 技能说明

当 AI 助手检测到以下任务时，会自动激活此技能：

- 创建新 API 接口
- 添加新业务模块
- 代码审查
- 数据库表设计和 SQL 操作
- 模块边界判断
- 项目初始化或架构设计

## 目录结构

```
backend-develop-guide/
├── SKILL.md                      # 核心技能文件（AI 读取）
├── README.md                     # 本文件（人类阅读）
└── reference/                    # 详细参考文档
    ├── architecture.md           # 分层架构详解
    ├── module-boundary.md        # 模块边界与可拆卸架构
    ├── code-patterns.md          # 代码模式规范
    ├── naming-conventions.md     # 命名规范
    ├── error-handling.md         # 错误处理规范
    ├── database-guide.md         # 数据库操作指南（sqlc）
    └── example/                  # 代码示例
        ├── new-api-example.md    # 新增 API 完整示例
        └── new-module-example.md # 新增业务模块完整示例
```

## 核心规范概览

### 三层架构

```
Handler (internal/handler/)  →  接收请求、参数绑定、调用 Logic、统一响应
    ↓
Logic (internal/logic/)      →  业务逻辑、编排 Repo 和 pkg 服务
    ↓
Repo (internal/repo/)        →  数据访问、与数据库交互

pkg (internal/pkg/)          →  基础设施包（横切关注点）
```

### 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.24+ |
| Web 框架 | Gin |
| 数据库 | PostgreSQL (pgx/v5) |
| ORM/Codegen | sqlc |
| 缓存 | Redis |
| 配置 | Viper + YAML |
| 日志 | Zap + Lumberjack |
| 认证 | JWT |

### 关键代码模式

1. **构造函数模式**：`NewXxxHandler()`, `NewXxxLogic()`, `NewXxxRepo()`
2. **接口定义**：`IXxxLogic`, `IXxxRepo`
3. **接口验证**：`var _ IXxxLogic = (*XxxLogic)(nil)`
4. **DTO 成对定义**：`XxxReq`, `XxxResp`
5. **统一响应**：`response.Response(c, resp, err)`
6. **泛型中间件**：`middleware.BindJsonMiddleware[T]`

## 使用方式

### 对于 AI 助手

此 Skill 会在以下情况自动激活：
- 用户请求创建 API
- 用户请求添加新功能模块
- 用户进行代码审查
- 用户询问架构相关问题

### 对于开发者

1. **新增 API**：参考 `reference/example/new-api-example.md`
2. **模块边界判断**：参考 `reference/module-boundary.md`
3. **新增模块**：参考 `reference/example/new-module-example.md`
4. **命名规范**：参考 `reference/naming-conventions.md`
5. **错误处理**：参考 `reference/error-handling.md`
6. **数据库操作**：参考 `reference/database-guide.md`

## 参考文档

| 文档 | 说明 |
|------|------|
| [架构详解](reference/architecture.md) | 三层架构、各层职责、依赖方向 |
| [模块边界](reference/module-boundary.md) | 共享三层与可拆卸模块的判断标准、目录和迁移清单 |
| [代码模式](reference/code-patterns.md) | 构造函数、接口、DTO、中间件等模式 |
| [命名规范](reference/naming-conventions.md) | 文件、包、函数、变量命名规范 |
| [错误处理](reference/error-handling.md) | 各层错误处理规范和最佳实践 |
| [数据库指南](reference/database-guide.md) | sqlc 工作流、SQL 编写规范 |
| [新增 API 示例](reference/example/new-api-example.md) | 在现有模块添加 API 的完整流程 |
| [新增模块示例](reference/example/new-module-example.md) | 从零创建新业务模块的完整流程 |

## 版本信息

- **技能版本**：1.0.0
- **基于项目**：nurture
- **最后更新**：2026-08-23
