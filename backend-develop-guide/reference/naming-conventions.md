# 命名规范

## 1. 目录命名

| 类型 | 规范 | 示例 |
|------|------|------|
| 业务目录 | 小写单词 | `handler/`, `logic/`, `repo/` |
| 多单词目录 | 小写下划线或无分隔 | `pkg/emailx/`, `pkg/pgsqlx/` |
| 模块子目录 | 小写单词 | `repo/user/`, `repo/order/` |

## 2. 文件命名

| 类型 | 规范 | 示例 |
|------|------|------|
| 普通文件 | 小写单词 | `user.go`, `common.go` |
| 入口文件 | `enter.go` | `config/enter.go`, `global/enter.go` |
| 错误定义 | `errors.go` | `logic/errors.go`, `repo/errors.go` |
| SQL 文件 | 模块名.sql | `user.sql`, `order.sql` |
| 测试文件 | `xxx_test.go` | `email_test.go`, `minio_test.go` |
| 配置模板 | `template.yaml` | `etc/template.yaml` |
| 本地配置 | `local.yaml` | `etc/local.yaml` |

## 3. 包命名

| 规范 | 说明 | 示例 |
|------|------|------|
| 小写单词 | 不使用下划线或驼峰 | `handler`, `logic`, `repo` |
| 简短有意义 | 避免过长名称 | `dto`, `pkg`, `config` |
| 基础设施包后缀 | 使用 `x` 后缀 | `emailx`, `jwtx`, `pgsqlx`, `redisx`, `zapx` |
| 避免通用名 | 避免 `util`, `common`, `helper` | 使用具体功能命名 |

### 基础设施包命名示例

```
internal/pkg/
├── emailx/     # 邮件服务（email + x）
├── jwtx/       # JWT 处理（jwt + x）
├── pgsqlx/     # PostgreSQL（pgsql + x）
├── redisx/     # Redis（redis + x）
├── zapx/       # Zap 日志（zap + x）
├── miniox/     # MinIO 对象存储（minio + x）
├── response/   # 统一响应（稳定横切 API，无 x 后缀）
└── syncx/      # 同步工具（sync + x）
```

## 4. 结构体命名

| 类型 | 规范 | 示例 |
|------|------|------|
| Handler | `{模块}Handler` | `UserHandler`, `OrderHandler` |
| Logic | `{模块}Logic` | `UserLogic`, `OrderLogic` |
| Repo | `{模块}Repo` | `UserRepo`, `OrderRepo` |
| Config | `{配置项}` | `App`, `DB`, `Redis`, `Auth` |
| DTO Request | `{操作}Req` | `LoginReq`, `RegisterReq`, `UpdateProfileReq` |
| DTO Response | `{操作}Resp` | `LoginResp`, `RegisterResp`, `UpdateProfileResp` |
| 数据库模型 | 单数名词 | `User`, `Order`, `Product` |

## 5. 接口命名

| 规范 | 说明 | 示例 |
|------|------|------|
| `I{结构体名}` | 以 `I` 开头 | `IUserLogic`, `IUserRepo` |
| 方法签名 | 与结构体方法一致 | - |

```go
// 接口
type IUserLogic interface {
    Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error)
    Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error)
}

// 接口
type IUserRepo interface {
    LoginWithAccount(ctx context.Context, account, password string) (user.User, error)
    Register(ctx context.Context, userID, username, email, account, password string) error
}
```

## 6. 函数/方法命名

### 构造函数

| 规范 | 示例 |
|------|------|
| `New{结构体}` | `NewUserHandler()`, `NewUserLogic()`, `NewUserRepo()` |

### Handler 方法

| 规范 | 示例 |
|------|------|
| 动词 + 名词 | `Login`, `Register`, `GetUser`, `UpdateProfile` |
| CRUD 操作 | `Create{X}`, `Get{X}`, `Update{X}`, `Delete{X}` |
| 列表操作 | `List{X}s` | `ListUsers`, `ListOrders` |

### Logic 方法

与 Handler 方法命名一致，因为是一对一调用关系：

```go
// Handler
func (uh *UserHandler) Login(c *gin.Context) { ... }

// Logic
func (ul *UserLogic) Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error) { ... }
```

### Repo 方法

| 规范 | 示例 |
|------|------|
| 操作 + By + 条件 | `GetUserByID`, `GetUserByEmail` |
| 操作 + With + 方式 | `LoginWithAccount`, `LoginWithEmail` |
| 简单操作 | `Create`, `Update`, `Delete` |
| 批量操作 | `BatchCreate`, `BatchUpdate` |

```go
func (ur *UserRepo) LoginWithAccount(ctx context.Context, account, password string) (user.User, error)
func (ur *UserRepo) LoginWithEmail(ctx context.Context, email string) (user.User, error)
func (ur *UserRepo) GetUserByID(ctx context.Context, userID string) (user.User, error)
func (ur *UserRepo) UpdateProfile(ctx context.Context, userID, nickname, avatar string) error
func (ur *UserRepo) UpdatePasswordByEmail(ctx context.Context, email, newPassword string) error
```

### 初始化函数

| 规范 | 示例 |
|------|------|
| `Init{组件}` | `InitZap()`, `InitPgsql()`, `InitRedis()`, `InitMinio()` |

## 7. 变量命名

### 局部变量

| 规范 | 示例 |
|------|------|
| 驼峰命名 | `userName`, `loginReq`, `userID` |
| 简短有意义 | `req`, `resp`, `err`, `ctx` |
| 循环变量 | `i`, `j`, `k` 或有意义的名称 |

### 结构体字段

```go
type UserHandler struct {
    userLogic *logic.UserLogic  // 私有字段：驼峰命名
}

type LoginResp struct {
    Token    string `json:"token"`     // 公开字段：帕斯卡命名
    Username string `json:"username"`
    Avatar   string `json:"avatar"`
}
```

### 全局变量

```go
// global/enter.go
var (
    Log *zap.SugaredLogger  // 帕斯卡命名（公开）
    DB  *pgxpool.Pool
    RDB redis.Cmdable
    MIO *minio.Client
)
```

### 配置变量

```go
// config/enter.go
var Conf = new(Config)  // 帕斯卡命名，表示公开的配置实例
```

## 8. 常量命名

| 规范 | 示例 |
|------|------|
| 全大写下划线 | `TOKEN_USER_ID`, `LOGIN_WITH_ACCOUNT` |
| 相关常量分组 | 使用 `const ( ... )` 分组 |

```go
package constant

const (
    // Context Keys
    TOKEN_USER_ID = "UserID"
    TOKEN_ROLE    = "Role"
    
    // 登录方式
    LOGIN_WITH_ACCOUNT = "account"
    LOGIN_WITH_EMAIL   = "email"
    
    // 业务常量
    DEFAULT_NODE_ID = 1
    FILE_MAX_SIZE   = 1024 * 1024 * 10
    
    // Redis Key 模板
    LOGIN_CODE_KEY     = "login_code:%s"
    REGISTER_CODE_KEY  = "register_code:%s"
    RESET_PWD_CODE_KEY = "reset_pwd_code:%s"
)
```

## 9. 错误变量命名

| 规范 | 示例 |
|------|------|
| `Err{描述}` | `ErrUserNotExist`, `ErrEmailIsUsed`, `ErrDefault` |
| 按类型分组 | 通用错误、业务错误分开定义 |

```go
// repo/errors.go
var (
    ErrDefault       = errors.New("默认错误")
    ErrUserNotExist  = errors.New("用户不存在")
    ErrEmailIsUsed   = errors.New("邮箱已经被使用")
    ErrAccountIsUsed = errors.New("账号已经被使用")
)

// logic/errors.go
var (
    ErrParamsType        = errors.New("参数格式错误")
    ErrDefault           = errors.New("默认错误")
    ErrLoginWithFailedWay = errors.New("暂不支持这种登录方式")
    ErrAccountOrPassword  = errors.New("账号或密码错误")
    ErrCodeVerify         = errors.New("验证码错误")
)
```

## 10. JSON 字段命名

| 规范 | 示例 |
|------|------|
| snake_case | `user_id`, `login_type`, `access_token` |
| 与 Go 字段对应 | `UserID` -> `user_id` |

```go
type LoginReq struct {
    Account   string `json:"account"`
    Password  string `json:"password"`
    Email     string `json:"email"`
    Code      string `json:"code"`
    LoginType string `json:"login_type"`
}
```

## 11. 数据库字段命名

| 规范 | 示例 |
|------|------|
| snake_case | `user_id`, `created_at`, `is_deleted` |
| 时间字段 | `ctime`（创建时间）, `utime`（更新时间） |
| 主键 | `id`（自增）, `user_id`（UUID） |

```sql
CREATE TABLE "user" (
    id        BIGSERIAL PRIMARY KEY,
    user_id   UUID UNIQUE NOT NULL,
    ctime     BIGINT NOT NULL,
    utime     BIGINT NOT NULL,
    account   VARCHAR(20) UNIQUE NOT NULL,
    password  VARCHAR(20) NOT NULL,
    email     VARCHAR(20) UNIQUE NOT NULL,
    username  VARCHAR(20) NOT NULL,
    avatar    VARCHAR(255) NOT NULL,
    role      SMALLINT NOT NULL DEFAULT 1
);
```

## 12. SQL 查询命名（sqlc）

| 规范 | 示例 |
|------|------|
| `{动作}{目标}` | `CreateUser`, `GetUserByEmail` |
| `{动作}{目标}By{条件}` | `GetUserByAccountAndPassword` |
| `Update{字段}By{条件}` | `UpdatePasswordByEmail`, `UpdateAvatarByUserID` |

```sql
-- name: GetUserByAccountAndPassword :one
SELECT * FROM "user"
WHERE account = $1 AND password = $2 LIMIT 1;

-- name: CreateUser :exec
INSERT INTO "user" (...) VALUES (...);

-- name: UpdatePasswordByEmail :execrows
UPDATE "user" SET password = $2 WHERE email = $1;
```

## 13. 路由路径命名

| 规范 | 示例 |
|------|------|
| 小写 + 短横线 | `/api/user/login`, `/api/common/file/upload` |
| RESTful 风格 | `GET /users`, `POST /users`, `PUT /users/:id` |
| 按模块分组 | `/api/user/*`, `/api/common/*`, `/api/order/*` |

```go
// 路由分组
CommonRoutes: router.Group("/api/common")
UserRoutes:   router.Group("/api/user")

// 具体路由
rg.POST("/login", ...)
rg.POST("/register", ...)
rg.POST("/code/login", ...)
rg.POST("/code/register", ...)
rg.POST("/resetPassword", ...)
```
