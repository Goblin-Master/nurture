# 分层架构详解

## 架构概览

本项目采用经典的三层架构（Three-Tier Architecture），确保关注点分离、可维护性和可扩展性。

默认使用共享三层目录。业务出现独立入口、独立数据边界、复杂状态流或可拆除诉求时，先阅读 `module-boundary.md` 判断是否拆成 `internal/<domain>` 可拆卸模块。

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Request                             │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Router (路由层)                              │
│  internal/router/*.go                                           │
│  职责：路由注册、中间件配置、路由分组                               │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Handler (控制器层)                           │
│  internal/handler/*.go                                          │
│  职责：接收请求、参数绑定、调用 Logic、统一响应                      │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Logic (业务逻辑层)                           │
│  internal/logic/*.go                                            │
│  职责：业务规则实现、编排 Repo 和 pkg 服务                          │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Repo (数据访问层)                            │
│  internal/repo/*.go + internal/repo/{module}/                   │
│  职责：数据库操作、缓存操作、错误转换                               │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Database / Cache                            │
│  PostgreSQL + Redis                                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                  pkg (基础设施包 - 横切关注点)                     │
│  internal/pkg/{emailx, jwtx, pgsqlx, redisx, zapx, response}    │
│  职责：日志、认证、邮件、数据库连接、统一响应等                       │
└─────────────────────────────────────────────────────────────────┘
```

## 各层详细说明

### 1. Handler 层（控制器层）

**位置**：`internal/handler/`

**职责**：
- 接收 HTTP 请求
- 使用中间件绑定和验证请求参数
- 调用 Logic 层的业务方法
- 使用统一响应格式返回结果

**规则**：
- 不包含任何业务逻辑
- 只负责请求/响应处理
- 每个 Handler 结构体持有对应的 Logic 实例

**示例结构**：

```go
package handler

type UserHandler struct {
    userLogic *logic.UserLogic
}

func NewUserHandler() *UserHandler {
    return &UserHandler{
        userLogic: logic.NewUserLogic(),
    }
}

func (uh *UserHandler) Login(c *gin.Context) {
    // 1. 获取绑定的请求参数
    req := middleware.GetBind[dto.LoginReq](c)
    
    // 2. 记录日志
    global.Log.Info(req)
    
    // 3. 调用 Logic 层
    resp, err := uh.userLogic.Login(c.Request.Context(), req)
    
    // 4. 统一响应
    response.Response(c, resp, err)
}
```

### 2. Logic 层（业务逻辑层）

**位置**：`internal/logic/`

**职责**：
- 实现核心业务逻辑
- 编排多个 Repo 和 pkg 服务
- 处理业务规则和流程
- 定义和返回业务错误

**规则**：
- 独立于 HTTP 上下文（不直接使用 `*gin.Context`，使用 `context.Context`）
- 接收 DTO，返回 DTO 或 error
- 必须定义接口 `IXxxLogic`
- 必须验证接口实现 `var _ IXxxLogic = (*XxxLogic)(nil)`

**示例结构**：

```go
package logic

// 接口定义
type IUserLogic interface {
    Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error)
    Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error)
}

// 结构体
type UserLogic struct {
    userRepo *repo.UserRepo
    email    *emailx.EmailX
}

// 构造函数
func NewUserLogic() *UserLogic {
    return &UserLogic{
        userRepo: repo.NewUserRepo(),
        email:    emailx.NewEmailX(),
    }
}

// 接口验证
var _ IUserLogic = (*UserLogic)(nil)

// 方法实现
func (ul *UserLogic) Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error) {
    var resp dto.LoginResp
    
    // 业务逻辑处理
    data, err := ul.userRepo.LoginWithAccount(ctx, req.Account, req.Password)
    if err != nil {
        return resp, ErrAccountOrPassword
    }
    
    // 生成 Token
    token, err := jwtx.GenToken(jwtx.Claims{
        UserID: data.UserID.String(),
        Role:   jwtx.Role(data.Role),
    })
    if err != nil {
        global.Log.Error(err)
        return resp, ErrDefault
    }
    
    resp.Username = data.Username
    resp.Token = token
    return resp, nil
}
```

### 3. Repo 层（数据访问层）

**位置**：`internal/repo/`

**职责**：
- 与数据库交互（通过 sqlc 生成的代码）
- 与缓存交互（Redis）
- 捕获数据库错误并转换为领域错误
- 记录技术日志

**规则**：
- 不包含业务逻辑
- 使用 sqlc 生成的 DAO 进行数据库操作
- 必须定义接口 `IXxxRepo`
- 错误必须转换为领域错误

**目录结构**：

```
internal/repo/
├── errors.go           # Repo 层错误定义
├── user.go             # UserRepo 实现
├── sql/                # SQL 查询文件
│   └── user.sql
├── sqlc.yaml           # sqlc 配置
└── user/               # sqlc 生成的代码
    ├── db.go
    ├── models.go
    └── user.sql.go
```

**示例结构**：

```go
package repo

// 接口定义
type IUserRepo interface {
    LoginWithAccount(ctx context.Context, account, password string) (user.User, error)
    Register(ctx context.Context, userID, username, email, account, password string) error
}

// 结构体
type UserRepo struct {
    userDao *user.Queries
}

// 构造函数
func NewUserRepo() *UserRepo {
    return &UserRepo{
        userDao: user.New(global.DB),
    }
}

// 接口验证
var _ IUserRepo = (*UserRepo)(nil)

// 方法实现
func (ur *UserRepo) LoginWithAccount(ctx context.Context, account, password string) (user.User, error) {
    u, err := ur.userDao.GetUserByAccountAndPassword(ctx, user.GetUserByAccountAndPasswordParams{
        Account:  account,
        Password: password,
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return user.User{}, ErrUserNotExist
        }
        global.Log.Error(err)
        return user.User{}, ErrDefault
    }
    return u, nil
}
```

### 4. pkg 基础设施包（横切关注点）

**位置**：`internal/pkg/`

**职责**：
- 提供与业务无关的通用功能
- 封装第三方库
- 初始化基础设施连接

**规则**：
- 不依赖业务层（Handler、Logic、Repo）
- 只依赖 config 和 global
- 每个包独立，职责单一

**包列表**：

| 包名 | 职责 |
|------|------|
| `pgsqlx` | PostgreSQL 连接池初始化 |
| `redisx` | Redis 客户端初始化 |
| `jwtx` | JWT Token 生成和解析 |
| `emailx` | 邮件发送服务 |
| `zapx` | Zap 日志配置 |
| `miniox` | MinIO 客户端初始化 |
| `response` | 统一响应格式 |
| `syncx` | 并发工具（泛型 sync.Map） |

## 依赖方向

```
Handler → Logic → Repo → sqlc DAO
    ↓        ↓       ↓
   pkg     pkg     pkg
    ↓        ↓       ↓
  global  global  global
    ↓        ↓       ↓
  config  config  config
```

**关键规则**：
1. 上层可以依赖下层，下层不能依赖上层
2. Handler 不能直接依赖 Repo
3. Logic 不能直接操作数据库（必须通过 Repo）
4. pkg 不能依赖业务层
5. 可拆卸模块内部仍遵守 Handler → Logic → Repo 方向，模块边界参考 `module-boundary.md`

## 全局变量管理

**位置**：`internal/global/enter.go`

全局变量用于存储基础设施实例，在应用启动时初始化：

```go
package global

var (
    Log *zap.SugaredLogger  // 日志
    DB  *pgxpool.Pool       // 数据库连接池
    RDB redis.Cmdable       // Redis 客户端
    MIO *minio.Client       // MinIO 客户端
)

func Init() {
    Log = zapx.InitZap()
    DB = pgsqlx.InitPgsql()
    RDB = redisx.InitRedis()
    MIO = miniox.InitMinio()
}
```

## 应用启动流程

**位置**：`internal/main.go`

```go
func main() {
    config.LoadConfig() // 1. 加载配置
    global.Init()       // 2. 初始化全局中间件
    router.RunServer()  // 3. 启动服务端
}
```

## 路由管理

**位置**：`internal/router/`

路由在 `router` 包内按业务模块拆分注册，`enter.go` 只负责 Gin 初始化、全局中间件和顶层路由组：

```go
func registerRoutes(r *gin.Engine) {
    registerHealthRoutes(r)
    api := r.Group("/api")

    registerUserRoutes(api.Group("/user"))
    registerFileRoutes(api.Group("/file"))
}

func registerUserRoutes(rg *gin.RouterGroup) {
    userHandler := handler.NewUserHandler()
    rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)
}

func registerFileRoutes(rg *gin.RouterGroup) {
    fileHandler := handler.NewFileHandler()
    rg.POST("/upload", middleware.Authentication(jwtx.COMMON_USER), fileHandler.Upload)
}

func registerHealthRoutes(r *gin.Engine) {
    r.GET("/healthz", func(c *gin.Context) {
        response.Response(c, "ok", nil)
    })
}
```
