# 错误处理规范

## 错误处理总体原则

1. **分层定义错误**：每层定义自己的错误，不跨层使用
2. **错误转换**：下层错误必须转换为上层可理解的错误
3. **日志记录**：技术错误在发生层记录，业务错误不重复记录
4. **用户友好**：最终返回给用户的错误信息必须友好可读

## 错误流转图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Handler 层                               │
│  职责：透传错误给 response.Response()                              │
│  一般不做额外处理，有需要定义：internal/handler/errors.go            │
└─────────────────────────────────────────────────────────────────┘
                              ↑ 业务错误
┌─────────────────────────────────────────────────────────────────┐
│                        Logic 层                                  │
│  职责：接收 Repo 错误，转换为业务错误                              │
│  定义：internal/logic/errors.go                                  │
└─────────────────────────────────────────────────────────────────┘
                              ↑ 领域错误
┌─────────────────────────────────────────────────────────────────┐
│                        Repo 层                                   │
│  职责：捕获数据库错误，转换为领域错误，记录技术日志                   │
│  定义：internal/repo/errors.go                                   │
└─────────────────────────────────────────────────────────────────┘
                              ↑ 数据库错误
┌─────────────────────────────────────────────────────────────────┐
│                      Database (pgx)                              │
│  原始错误：pgx.ErrNoRows, pgconn.PgError 等                       │
└─────────────────────────────────────────────────────────────────┘
```

## 1. Repo 层错误处理

### 错误定义位置

`internal/repo/errors.go`

### 错误定义示例

```go
package repo

import "errors"

var (
    // 通用错误
    ErrDefault = errors.New("默认错误")
)

var (
    // 用户相关错误
    ErrUserNotExist  = errors.New("用户不存在")
    ErrEmailIsUsed   = errors.New("邮箱已经被使用")
    ErrAccountIsUsed = errors.New("账号已经被使用")
)

var (
    // 订单相关错误（示例）
    ErrOrderNotExist = errors.New("订单不存在")
    ErrOrderClosed   = errors.New("订单已关闭")
)
```

### 错误处理规则

1. **捕获 pgx 错误**：
   - `pgx.ErrNoRows` → 转换为 `ErrXxxNotExist`
   - `pgconn.PgError` → 根据错误码转换

2. **记录技术日志**：使用 `global.Log.Error(err)` 记录原始错误

3. **返回领域错误**：返回定义好的领域错误，不暴露数据库细节

### 代码示例

```go
package repo

import (
    "errors"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

// 查询操作 - 处理 ErrNoRows
func (ur *UserRepo) LoginWithAccount(ctx context.Context, account, password string) (user.User, error) {
    u, err := ur.userDao.GetUserByAccountAndPassword(ctx, user.GetUserByAccountAndPasswordParams{
        Account:  account,
        Password: password,
    })
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            // 转换为领域错误，不记录日志（这是业务正常情况）
            return user.User{}, ErrUserNotExist
        }
        // 记录技术错误
        global.Log.Error(err)
        return user.User{}, ErrDefault
    }
    return u, nil
}

// 插入操作 - 处理唯一约束冲突
func (ur *UserRepo) Register(ctx context.Context, userID, username, email, account, password string) error {
    err := ur.userDao.CreateUser(ctx, user.CreateUserParams{
        UserID:   userUUID,
        Username: username,
        Email:    email,
        Account:  account,
        Password: password,
        Ctime:    time.Now().UnixMilli(),
        Utime:    time.Now().UnixMilli(),
    })

    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
            // 根据约束名判断是哪个字段冲突
            switch pgErr.ConstraintName {
            case "user_account_key":
                return ErrAccountIsUsed
            case "user_email_key":
                return ErrEmailIsUsed
            }
        }
        global.Log.Error(err)
        return ErrDefault
    }
    return nil
}

// 更新操作 - 处理影响行数为 0
func (ur *UserRepo) ResetPassword(ctx context.Context, email, newPassword string) error {
    count, err := ur.userDao.UpdatePasswordByEmail(ctx, user.UpdatePasswordByEmailParams{
        Email:    email,
        Password: newPassword,
    })
    if err != nil {
        global.Log.Error(err)
        return ErrDefault
    }
    if count == 0 {
        // 没有更新任何行，说明用户不存在
        return ErrUserNotExist
    }
    return nil
}
```

### PostgreSQL 常见错误码

| 错误码 | 含义 | 处理方式 |
|--------|------|----------|
| `23505` | unique_violation | 检查 `ConstraintName` 判断哪个字段冲突 |
| `23503` | foreign_key_violation | 外键约束失败 |
| `23502` | not_null_violation | 非空约束失败 |
| `23514` | check_violation | 检查约束失败 |

## 2. Logic 层错误处理

### 错误定义位置

`internal/logic/errors.go`

### 错误定义示例

```go
package logic

import (
    "errors"
    "fmt"
    "nurture/internal/constant"
)

// 通用错误
var (
    ErrParamsType   = errors.New("参数格式错误")
    ErrDefault      = errors.New("默认错误")
    ErrFileOverSize = fmt.Errorf("文件大小不能超过%dMB", constant.FILE_MAX_SIZE/1024/1024)
    ErrFileRead     = errors.New("文件读取失败")
    ErrFileUpload   = errors.New("文件上传失败")
)

// 用户相关错误
var (
    ErrLoginWithFailedWay = errors.New("暂不支持这种登录方式")
    ErrAccountOrPassword  = errors.New("账号或密码错误")
    ErrEmail              = errors.New("邮箱错误")
    ErrCodeGet            = errors.New("code获取失败")
    ErrCodeVerify         = errors.New("验证码错误")
    ErrEmailIsUsed        = errors.New("邮箱已经被使用")
    ErrAccountIsUsed      = errors.New("账号已经被使用")
    ErrUserNotExist       = errors.New("用户不存在")
)
```

### 错误处理规则

1. **接收 Repo 错误**：使用 `errors.Is()` 判断错误类型
2. **转换为业务错误**：将领域错误转换为用户友好的业务错误
3. **记录意外错误**：对于非预期错误，记录日志并返回默认错误

### 代码示例

```go
package logic

import (
    "context"
    "errors"
)

func (ul *UserLogic) Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error) {
    var resp dto.LoginResp
    
    switch req.LoginType {
    case constant.LOGIN_WITH_ACCOUNT:
        data, err := ul.userRepo.LoginWithAccount(ctx, req.Account, req.Password)
        if err != nil {
            // Repo 层已经返回了 ErrUserNotExist，这里转换为更友好的错误
            return resp, ErrAccountOrPassword
        }
        
        token, err := jwtx.GenToken(jwtx.Claims{
            UserID: data.UserID.String(),
            Role:   jwtx.Role(data.Role),
        })
        if err != nil {
            // 记录技术错误
            global.Log.Error(err)
            return resp, ErrDefault
        }
        
        resp.Token = token
        resp.Username = data.Username
        return resp, nil
        
    default:
        global.Log.Warnf("错误的登录方式:%s", req.LoginType)
        return resp, ErrLoginWithFailedWay
    }
}

func (ul *UserLogic) Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error) {
    var resp dto.RegisterResp
    
    // 验证验证码
    ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(constant.REGISTER_CODE_KEY, req.Email), req.Code)
    if err != nil {
        global.Log.Error(err)
        return resp, ErrCodeVerify
    }
    if !ok {
        return resp, ErrCodeVerify
    }
    
    // 注册用户
    err = ul.userRepo.Register(ctx, uuid.NewString(), req.Username, req.Email, req.Account, req.Password)
    if err != nil {
        // 使用 errors.Is 判断具体错误类型
        if errors.Is(err, repo.ErrEmailIsUsed) {
            return resp, ErrEmailIsUsed
        } else if errors.Is(err, repo.ErrAccountIsUsed) {
            return resp, ErrAccountIsUsed
        } else {
            global.Log.Error(err)
            return resp, ErrDefault
        }
    }
    
    resp.Message = "用户注册成功！"
    return resp, nil
}
```

## 3. Handler 层错误处理

### 错误处理规则

1. **统一返回**：使用 `response.Response(c, data, err)` 统一返回，`err` 会被自动解析并返回错误码和消息。
2. **不做额外处理**：Handler 层一般不进行错误判断，除非需要处理 HTTP 层面的逻辑，提前返回错误等。
3. **自定义错误返回**：如果需要提前中断并返回错误，应定义并使用自定义错误，而不是直接返回 HTTP 状态码。

### 代码示例

```go
// internal/handler/user.go
func (uh *UserHandler) Login(c *gin.Context) {
    req := middleware.GetBind[dto.LoginReq](c)
    
    // 调用 Logic
    resp, err := uh.userLogic.Login(c.Request.Context(), req)
    
    // 如果 Logic 返回错误，Response 会自动处理
    response.Response(c, resp, err)
}

// 提前返回示例
func (h *CommonHandler) UploadFile(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        // 使用自定义错误提前返回
        response.Response(c, nil, ErrFileRead)
        return
    }
    // ...
}
```

## 4. 统一响应格式

### 响应结构

```go
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

### 响应示例

**成功响应**：
```json
{
    "code": 0,
    "message": "OK",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "username": "张三",
        "avatar": "https://..."
    }
}
```

**错误响应**：
```json
{
    "code": -1,
    "message": "账号或密码错误",
    "data": null
}
```

## 5. 中间件错误处理

### 参数绑定中间件

```go
func BindJsonMiddleware[T any](c *gin.Context) {
    var req T
    err := c.ShouldBindJSON(&req)
    if err != nil {
        response.Response(c, nil, err)
        c.Abort()  // 终止请求
        return
    }
    c.Set("request", req)
}
```

### 认证中间件

```go
func Authentication(role jwtx.Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        UserID, Role, err := jwtx.ParseToken(c)
        if err != nil {
            // 401 Unauthorized
            c.JSON(401, response.Body{
                Code:    -1,
                Message: err.Error(),
                Data:    nil,
            })
            c.Abort()
            return
        }
        if Role < role {
            // 403 Forbidden
            c.JSON(403, response.Body{
                Code:    -1,
                Message: jwtx.ErrPermissionDenied.Error(),
                Data:    nil,
            })
            c.Abort()
            return
        }
        c.Set(constant.TOKEN_USER_ID, UserID)
        c.Set(constant.TOKEN_ROLE, Role)
        c.Next()
    }
}
```

## 6. pkg 层错误定义

基础设施包也可以定义自己的错误：

```go
// pkg/jwtx/enter.go
var (
    ErrDefault          = errors.New("jwt default error")
    ErrTokenEmpty       = errors.New("token is empty")
    ErrTokenExpired     = errors.New("token has expired")
    ErrTokenInvalid     = errors.New("token is invalid")
    ErrPermissionDenied = errors.New("permission denied")
)

// pkg/emailx/enter.go
var (
    ErrSendOverTime = errors.New("邮件发送超时")
)
```

## 7. 错误处理最佳实践

### Do（推荐做法）

1. ✅ 使用 `errors.Is()` 判断错误类型
2. ✅ 在 Repo 层记录技术错误日志
3. ✅ 定义清晰的错误变量名：`ErrXxx`
4. ✅ 错误信息使用中文，便于直接返回给用户
5. ✅ 使用 `errors.As()` 提取嵌套错误信息

### Don't（避免做法）

1. ❌ 将数据库错误信息暴露给用户
2. ❌ 在多层重复记录同一个错误
3. ❌ 使用 `err.Error() == "xxx"` 进行错误判断
4. ❌ 跨层使用错误变量（如 Handler 层直接使用 Logic 层错误）