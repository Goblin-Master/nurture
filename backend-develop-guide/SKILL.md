---
name: backend-develop-guide
description: 当用户需要开发 Go 后端 API、创建新业务模块、或进行代码审查时使用。适用于基于 Gin + PostgreSQL + Redis 的三层架构项目。
---

# Go 后端开发指南

本技能提供基于 Gin 框架的 Go 后端开发规范，遵循三层架构设计模式。

## 触发场景

当用户执行以下任务时激活此技能：

- 创建新 API 接口
- 添加新业务模块（如：用户模块、订单模块）
- 进行代码审查
- 数据库表设计和 SQL 操作
- 项目初始化或架构设计

## 核心架构：三层架构

```
HTTP Request
     ↓
┌─────────────────────────────────────┐
│  Handler 层 (internal/handler/)     │  ← 接收请求、参数绑定、调用 Logic、统一响应
└─────────────────────────────────────┘
     ↓
┌─────────────────────────────────────┐
│  Logic 层 (internal/logic/)         │  ← 业务逻辑、编排 Repo 和 pkg 服务
└─────────────────────────────────────┘
     ↓
┌─────────────────────────────────────┐
│  Repo 层 (internal/repo/)           │  ← 数据访问、与数据库交互
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│  pkg 基础设施包 (internal/pkg/)      │  ← 横切关注点：jwtx、emailx、pgsqlx 等
└─────────────────────────────────────┘
```

### 各层职责

| 层级 | 目录 | 职责 | 禁止事项 |
|------|------|------|----------|
| Handler | `internal/handler/` | 参数绑定、调用 Logic、统一响应 | 不写业务逻辑 |
| Logic | `internal/logic/` | 业务规则、编排 Repo 和 pkg | 不直接操作数据库 |
| Repo | `internal/repo/` | 数据库操作、缓存操作 | 不写业务逻辑 |
| pkg | `internal/pkg/` | 基础设施（日志、JWT、邮件等） | 不依赖业务层 |

## 标准目录结构

```
internal/
├── config/          # 配置结构定义和加载
│   ├── enter.go     # Config 结构体定义
│   └── read.go      # 配置加载逻辑
├── constant/        # 常量定义
│   └── enter.go
├── dto/             # 数据传输对象（Request/Response）
│   └── user.go
├── etc/             # 配置文件
│   ├── local.yaml
│   └── template.yaml
├── global/          # 全局变量（DB、Redis、Log）
│   └── enter.go
├── handler/         # HTTP 处理器
│   └── user.go
├── logic/           # 业务逻辑
│   ├── errors.go    # Logic 层错误定义
│   └── user.go
├── middleware/      # 中间件
│   ├── bind.go      # 参数绑定中间件
│   ├── cors.go      # CORS 中间件
│   └── jwt.go       # JWT 认证中间件
├── pkg/             # 基础设施包
│   ├── emailx/
│   ├── jwtx/
│   ├── pgsqlx/
│   ├── redisx/
│   ├── response/
│   └── zapx/
├── repo/            # 数据访问层
│   ├── errors.go    # Repo 层错误定义
│   ├── sql/         # SQL 查询文件（sqlc 使用）
│   ├── sqlc.yaml    # sqlc 配置
│   └── user/        # sqlc 生成的代码
├── router/          # 路由初始化
│   ├── enter.go     # Gin 初始化、全局中间件、顶层路由组
│   └── user.go      # 模块路由注册
└── main.go          # 应用入口
```

## 新增 API 开发流程

### 步骤 1：定义 DTO（`internal/dto/`）

```go
// internal/dto/user.go
type (
    UpdateProfileReq struct {
        Nickname string `json:"nickname" binding:"required"`
        Avatar   string `json:"avatar"`
    }
    UpdateProfileResp struct {
        Message string `json:"message"`
    }
)
```

### 步骤 2：定义 SQL 查询（`internal/repo/sql/`）

```sql
-- internal/repo/sql/user.sql
-- name: UpdateUserProfile :execrows
UPDATE "user"
SET username = $2, avatar = $3, utime = $4
WHERE user_id = $1;
```

### 步骤 3：生成 sqlc 代码

```bash
sqlc generate -f internal/repo/sqlc.yaml
```

### 步骤 4：实现 Repo 层（`internal/repo/`）

```go
// internal/repo/user.go
func (ur *UserRepo) UpdateProfile(ctx context.Context, userID, nickname, avatar string) error {
    var userUUID pgtype.UUID
    if err := userUUID.Scan(userID); err != nil {
        return err
    }
    count, err := ur.userDao.UpdateUserProfile(ctx, user.UpdateUserProfileParams{
        UserID:   userUUID,
        Username: nickname,
        Avatar:   avatar,
        Utime:    time.Now().UnixMilli(),
    })
    if err != nil {
        global.Log.Error(err)
        return ErrDefault
    }
    if count == 0 {
        return ErrUserNotExist
    }
    return nil
}
```

### 步骤 5：实现 Logic 层（`internal/logic/`）

```go
// internal/logic/user.go
func (ul *UserLogic) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileReq) (dto.UpdateProfileResp, error) {
    var resp dto.UpdateProfileResp
    err := ul.userRepo.UpdateProfile(ctx, userID, req.Nickname, req.Avatar)
    if err != nil {
        if errors.Is(err, repo.ErrUserNotExist) {
            return resp, ErrUserNotExist
        }
        global.Log.Error(err)
        return resp, ErrDefault
    }
    resp.Message = "更新成功"
    return resp, nil
}
```

### 步骤 6：实现 Handler 层（`internal/handler/`）

```go
// internal/handler/user.go
func (uh *UserHandler) UpdateProfile(c *gin.Context) {
    req := middleware.GetBind[dto.UpdateProfileReq](c)
    userID := jwtx.GetUserID(c)
    resp, err := uh.userLogic.UpdateProfile(c.Request.Context(), userID, req)
    response.Response(c, resp, err)
}
```

### 步骤 7：注册路由（`internal/router/`）

```go
// internal/router/user.go
func registerUserRoutes(rg *gin.RouterGroup) {
    userHandler := handler.NewUserHandler()
    rg.PUT("/profile", middleware.Authentication(jwtx.COMMON_USER), 
        middleware.BindJsonMiddleware[dto.UpdateProfileReq], userHandler.UpdateProfile)
}
```

## 关键代码模式

### 1. 构造函数模式

每层使用 `NewXxx()` 构造函数创建实例：

```go
// Handler
func NewUserHandler() *UserHandler {
    return &UserHandler{
        userLogic: logic.NewUserLogic(),
    }
}

// Logic
func NewUserLogic() *UserLogic {
    return &UserLogic{
        userRepo: repo.NewUserRepo(),
        email:    emailx.NewEmailX(),
    }
}

// Repo
func NewUserRepo() *UserRepo {
    return &UserRepo{
        userDao: user.New(global.DB),
    }
}
```

### 2. 接口定义与验证

Logic 和 Repo 层必须定义接口，并验证实现：

```go
// 接口定义
type IUserLogic interface {
    Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error)
    Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error)
}

// 接口验证（编译时检查）
var _ IUserLogic = (*UserLogic)(nil)
```

### 3. 泛型参数绑定中间件

使用泛型中间件绑定请求参数：

```go
// 路由注册
rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)

// Handler 中获取
func (uh *UserHandler) Login(c *gin.Context) {
    req := middleware.GetBind[dto.LoginReq](c)
    // ...
}
```

### 4. 统一响应格式

所有 Handler 使用统一响应：

```go
response.Response(c, resp, err)

// 响应结构
// 成功: { "code": 0, "message": "OK", "data": {...} }
// 失败: { "code": -1, "message": "错误信息", "data": null }
```

## 错误处理规范

### Repo 层

- 捕获数据库错误（`pgx.ErrNoRows`、`pgconn.PgError`）
- 转换为领域错误（`ErrUserNotExist`、`ErrEmailIsUsed`）
- 记录技术日志：`global.Log.Error(err)`

### Logic 层

- 定义业务错误（`internal/logic/errors.go`）
- 接收 Repo 错误并转换为用户友好消息
- 返回业务错误给 Handler

### Handler 层

- 直接传递错误给 `response.Response()`
- **自定义错误返回**：如果需要提前返回错误，应定义并使用自定义错误，而不是直接返回 HTTP 状态码，保持响应格式统一。
- **鉴权数据获取**：如果接口需要鉴权，`UserID` 必须从请求头（Token）中获取（使用 `jwtx.GetUserID(c)`），禁止从请求参数中获取（除非特殊场景，如管理员查询用户信息）。

## 测试规范

- **pkg 测试位置**：`internal/pkg/<pkg>` 的单元测试和基础设施集成测试必须放在各自包下，例如 `internal/pkg/jwtx/jwtx_test.go`、`internal/pkg/ratelimitx/ratelimitx_test.go`。pkg 测试不得放业务测试。
- **模块测试位置**：可拆卸业务模块允许自带测试包，例如 `internal/chat/test`。模块测试只覆盖该模块自己的 handler、logic、repo、session、worker 等边界。
- **业务联调测试位置**：`internal/test` 只放跨业务层联调测试，例如需要串联 logic、repo、global、pkg 或真实外部服务的业务链路。
- **依赖初始化**：pkg 集成测试只初始化当前包必要依赖，避免使用 `global.Init()` 拉起无关 DB、RabbitMQ、AI 等服务。
- **Token 生成**：测试时如需 Token，使用 `jwtx.GenTestToken()` 生成，禁止使用硬编码 Token。
- **配置加载**：测试代码中涉及密钥或配置时，必须通过 `config.LoadConfig()` 加载，**禁止在代码中明文写死密钥**。

## 安全规范

- **密钥管理**：所有敏感信息（API Key、Secret Key、密码等）必须通过配置文件加载，**严禁在代码中硬编码**。
- **鉴权**：业务接口默认开启 JWT 鉴权，`UserID` 必须从 Token 解析。

## AI 模块开发规范

- **文档参考**：AI 系统详细设计参考 `docs/ai-system-srs.md`。
- **接口位置**：AI 相关接口统一放在 `/api/common/ai/` 路由组下。
- **鉴权**：所有 AI 接口强制鉴权，`UserID` 从 Token 获取，`SessionID` 由前端生成并传递。
- **配置**：AI 模型密钥（API Key）必须从 `config` 读取，禁止硬编码。

## 技术栈

| 组件 | 技术选型 |
|------|----------|
| 语言 | Go 1.24+ |
| Web 框架 | Gin |
| 数据库 | PostgreSQL (pgx/v5) |
| ORM/Codegen | sqlc |
| 缓存 | Redis |
| 配置 | Viper + YAML |
| 日志 | Zap + Lumberjack |
| 认证 | JWT |
| 对象存储 | MinIO |

## 详细参考文档

- 分层架构详解：`reference/architecture.md`
- 代码模式规范：`reference/code-patterns.md`
- 命名规范：`reference/naming-conventions.md`
- 错误处理规范：`reference/error-handling.md`
- 数据库操作指南：`reference/database-guide.md`
- 代码示例：`reference/examples/`
