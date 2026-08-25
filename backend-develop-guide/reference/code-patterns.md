# 代码模式规范

## 1. 构造函数模式

每层使用 `NewXxx()` 构造函数创建实例，而不是直接初始化结构体。

### Handler 层

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
```

### Logic 层

```go
package logic

type UserLogic struct {
    userRepo *repo.UserRepo
    email    *emailx.EmailX
}

func NewUserLogic() *UserLogic {
    return &UserLogic{
        userRepo: repo.NewUserRepo(),
        email:    emailx.NewEmailX(),
    }
}
```

### Repo 层

```go
package repo

type UserRepo struct {
    userDao *user.Queries
}

func NewUserRepo() *UserRepo {
    return &UserRepo{
        userDao: user.New(global.DB),
    }
}
```

### pkg 层

```go
package emailx

type EmailX struct {
    config config.Email
    ttl    time.Duration
    rdb    redis.Cmdable
}

func NewEmailX() *EmailX {
    return &EmailX{
        config: config.Conf.Email,
        ttl:    10 * time.Minute,
        rdb:    global.RDB,
    }
}
```

### pkg 脚本资源

pkg 自己使用的 Lua、SQL 片段、模板等脚本资源必须放在当前 pkg 的 `scripts/` 目录下，并通过 `go:embed` 嵌入代码。不要把多行脚本直接内联在 `enter.go` 或业务方法里。

```text
internal/pkg/emailx/
├── enter.go
└── scripts/
    └── verify.lua
```

```go
package emailx

import _ "embed"

//go:embed scripts/verify.lua
var verifyScript string
```

脚本归属哪个 pkg，就放在哪个 pkg 自己的 `scripts/` 下；不要放进 `internal/test`、业务模块或共享杂物目录。

## 2. 接口定义模式

Logic 层和 Repo 层必须定义接口，便于测试和解耦。

### 接口命名规范

- 接口名：`I{结构体名}` （如 `IUserLogic`、`IUserRepo`）
- 放在结构体定义之前

### 接口定义示例

```go
package logic

// 接口定义（放在文件开头）
type IUserLogic interface {
    Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error)
    Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error)
    GetLoginCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
}

// 结构体定义
type UserLogic struct {
    userRepo *repo.UserRepo
    email    *emailx.EmailX
}

// 接口验证（编译时检查）
var _ IUserLogic = (*UserLogic)(nil)
```

### 接口验证

使用 `var _ Interface = (*Struct)(nil)` 在编译时验证接口实现：

```go
// 如果 UserLogic 没有实现 IUserLogic 的所有方法，编译会报错
var _ IUserLogic = (*UserLogic)(nil)
var _ IUserRepo = (*UserRepo)(nil)
```

### 避免过度接口包装

- Logic/Repo 的接口用于表达业务层契约和便于测试。
- 对已有稳定横切能力，例如 `middleware.Authentication`、`middleware.GetBind`、`jwtx.GetUserID`、`response.Response`，优先直接复用原包。
- 不要为了“看起来可注入”而额外定义 `AuthUserFunc`、`RespondFunc`、`GetUserIDFunc` 这类只有一处调用且没有替换需求的包装类型。
- 只有当业务层确实需要 fake、替换实现或隔离外部系统时，才在使用方定义小接口。

## 3. DTO 定义模式

DTO（Data Transfer Object）用于定义请求和响应结构。

### 命名规范

- 请求：`{操作}Req`（如 `LoginReq`、`RegisterReq`）
- 响应：`{操作}Resp`（如 `LoginResp`、`RegisterResp`）

### 组织方式

使用类型分组语法，将相关的 Req/Resp 放在一起：

```go
package dto

// 登录相关
type (
    LoginReq struct {
        Account   string `json:"account"`
        Password  string `json:"password"`
        Email     string `json:"email"`
        Code      string `json:"code"`
        LoginType string `json:"login_type"`
    }
    LoginResp struct {
        Token    string `json:"token"`
        Username string `json:"username"`
        Avatar   string `json:"avatar"`
    }
)

// 注册相关
type (
    RegisterReq struct {
        Account  string `json:"account"`
        Password string `json:"password"`
        Username string `json:"username"`
        Email    string `json:"email"`
        Code     string `json:"code"`
    }
    RegisterResp struct {
        Message string `json:"message"`
    }
)
```

### JSON Tag 规范

- 使用 `json:"field_name"` 标签
- 字段名使用 snake_case
- 必填字段添加 `binding:"required"`

```go
type UpdateProfileReq struct {
    Nickname string `json:"nickname" binding:"required"`
    Avatar   string `json:"avatar"`
    Bio      string `json:"bio" binding:"max=200"`
}
```

## 4. 中间件使用模式

### 泛型参数绑定中间件

使用泛型中间件自动绑定请求参数：

```go
// middleware/bind.go
func BindJsonMiddleware[T any](c *gin.Context) {
    var req T
    err := c.ShouldBindJSON(&req)
    if err != nil {
        response.Response(c, nil, err)
        c.Abort()
        return
    }
    c.Set("request", req)
}

func GetBind[T any](c *gin.Context) T {
    return c.MustGet("request").(T)
}
```

### 使用方式

```go
// 路由注册
rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)
rg.GET("/users", middleware.BindQueryMiddleware[dto.ListUsersReq], userHandler.ListUsers)
rg.GET("/users/:id", middleware.BindUriMiddleware[dto.GetUserReq], userHandler.GetUser)

// Handler 中获取
func (uh *UserHandler) Login(c *gin.Context) {
    req := middleware.GetBind[dto.LoginReq](c)
    // 使用 req...
}
```

### 可用的绑定中间件

| 中间件 | 用途 | 对应 Gin 方法 |
|--------|------|---------------|
| `BindJsonMiddleware[T]` | 绑定 JSON Body | `ShouldBindJSON` |
| `BindQueryMiddleware[T]` | 绑定 Query 参数 | `ShouldBindQuery` |
| `BindUriMiddleware[T]` | 绑定 URI 参数 | `ShouldBindUri` |

## 5. 认证中间件模式

### JWT 认证中间件

```go
// middleware/jwt.go
func Authentication(role jwtx.Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        UserID, Role, err := jwtx.ParseToken(c)
        if err != nil {
            c.JSON(401, response.Body{
                Code:    -1,
                Message: err.Error(),
                Data:    nil,
            })
            c.Abort()
            return
        }
        if Role < role {
            c.JSON(403, response.Body{
                Code:    -1,
                Message: jwtx.ErrPermissionDenied.Error(),
                Data:    nil,
            })
            c.Abort()
            return
        }
        // 将用户信息存入 Context
        c.Set(jwtx.ContextUserIDKey, UserID)
        c.Set(jwtx.ContextRoleKey, Role)
        c.Next()
    }
}
```

### 使用方式

```go
// 需要普通用户权限
rg.POST("/profile", middleware.Authentication(jwtx.COMMON_USER), handler.UpdateProfile)

// 需要管理员权限
rg.DELETE("/users/:id", middleware.Authentication(jwtx.ADMIN), handler.DeleteUser)
```

### 获取用户信息

```go
// 必须在使用了 Authentication 中间件的路由中使用
func (uh *UserHandler) UpdateProfile(c *gin.Context) {
    userID := jwtx.GetUserID(c)
    role := jwtx.GetRole(c)
    // ...
}
```

## 6. 统一响应模式

### 响应结构

```go
package response

type Body struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

### 响应函数

```go
func Response(c *gin.Context, resp interface{}, err error) {
    var body Body
    if err != nil {
        body.Code = -1
        body.Message = err.Error()
        body.Data = nil
    } else {
        body.Code = 0
        body.Message = "OK"
        body.Data = resp
    }
    c.JSON(200, body)
}
```

### 使用方式

```go
// 成功响应
response.Response(c, dto.LoginResp{Token: "xxx"}, nil)
// 输出: {"code": 0, "message": "OK", "data": {"token": "xxx"}}

// 错误响应
response.Response(c, nil, errors.New("用户不存在"))
// 输出: {"code": -1, "message": "用户不存在", "data": null}
```

## 7. 配置结构模式

### 配置结构定义

```go
// config/enter.go
var Conf = new(Config)

type Config struct {
    App   App   `mapstructure:"app"`
    DB    DB    `mapstructure:"db"`
    Redis Redis `mapstructure:"redis"`
    Auth  Auth  `mapstructure:"auth"`
}

type App struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    Env  string `mapstructure:"env"`
    Log  string `mapstructure:"log"`
}

// 辅助方法
func (app *App) Link() string {
    return fmt.Sprintf("%s:%d", app.Host, app.Port)
}

type DB struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"dbname"`
}

func (db *DB) DSN() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        db.Username, db.Password, db.Host, db.Port, db.DBName)
}
```

### 配置加载

```go
// config/read.go
func LoadConfig() {
    viper.SetConfigName("local")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("internal/etc")
    viper.AddConfigPath("etc")
    viper.AddConfigPath(".")
    
    if err := viper.ReadInConfig(); err != nil {
        panic(fmt.Errorf("配置读取失败: %w", err))
    }
    if err := viper.Unmarshal(&Conf); err != nil {
        panic(fmt.Errorf("配置解析失败: %w", err))
    }
}
```

## 8. 路由注册模式

```go
// router/enter.go
func registerRoutes(r *gin.Engine) {
    registerHealthRoutes(r)
    api := r.Group("/api")
    registerFileRoutes(api.Group("/file"))
    registerUserRoutes(api.Group("/user"))
}

func requestGlobalMiddleware(r *gin.Engine) {
    r.Use(middleware.Cors())
}
```

模块路由按文件拆分在 `internal/router/` 下：

```go
// router/health.go
func registerHealthRoutes(r *gin.Engine) {
    r.GET("/healthz", func(c *gin.Context) {
        response.Response(c, "ok", nil)
    })
}

// router/file.go
func registerFileRoutes(rg *gin.RouterGroup) {
    fileHandler := handler.NewFileHandler()
    rg.POST("/upload", middleware.Authentication(jwtx.COMMON_USER), fileHandler.Upload)
}

// router/user.go
func registerUserRoutes(rg *gin.RouterGroup) {
    userHandler := handler.NewUserHandler()
    rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)
    rg.POST("/register", middleware.BindJsonMiddleware[dto.RegisterReq], userHandler.Register)
}
```

## 9. 常量定义模式

根级 `internal/constant` 容易退化成杂物包。常量应优先归属业务模块或基础设施包自身，例如 `internal/user/constant`、`internal/chat/constant`、`internal/pkg/jwtx`。

```go
// user/constant/enter.go
package constant

const (
    // 登录方式
    LOGIN_WITH_ACCOUNT = "account"
    LOGIN_WITH_EMAIL   = "email"

    // Redis Key 模板
    LOGIN_CODE_KEY     = "login_code:%s"
    REGISTER_CODE_KEY  = "register_code:%s"
    RESET_PWD_CODE_KEY = "reset_pwd_code:%s"
)
```

## 10. 角色权限模式

```go
// pkg/jwtx/enter.go
type Role int

const (
    COMMON_USER   Role = iota + 1  // 普通用户
    INTERNAL_USER                  // 内部用户
    ADMIN                          // 管理员
)
```

权限校验使用数值比较，数值越大权限越高：

```go
if Role < role {
    // 权限不足
}
```
