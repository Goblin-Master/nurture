# 新增 API 完整示例

本示例展示如何在现有模块中添加一个新的 API 接口。

## 场景

在用户模块中添加「更新用户资料」接口：

- **路径**：`PUT /api/user/profile`
- **认证**：需要登录
- **功能**：更新当前用户的昵称和头像

## 步骤 1：定义 DTO

**文件**：`internal/dto/user.go`

```go
package dto

// 在文件中添加新的 DTO 定义

type (
    UpdateProfileReq struct {
        Nickname string `json:"nickname" binding:"required,max=20"`
        Avatar   string `json:"avatar" binding:"omitempty,url"`
    }
    UpdateProfileResp struct {
        Message string `json:"message"`
    }
)
```

## 步骤 2：编写 SQL 查询

**文件**：`internal/repo/sql/user.sql`

```sql
-- 在文件末尾添加新的查询

-- name: UpdateProfile :execrows
UPDATE "user"
SET username = $2, avatar = $3, utime = $4
WHERE user_id = $1;
```

## 步骤 3：生成 sqlc 代码

```bash
sqlc generate -f internal/repo/sqlc.yaml
```

生成后会在 `internal/repo/user/user.sql.go` 中自动添加：

```go
const updateProfile = `-- name: UpdateProfile :execrows
UPDATE "user"
SET username = $2, avatar = $3, utime = $4
WHERE user_id = $1
`

type UpdateProfileParams struct {
    UserID   pgtype.UUID
    Username string
    Avatar   string
    Utime    int64
}

func (q *Queries) UpdateProfile(ctx context.Context, arg UpdateProfileParams) (int64, error) {
    result, err := q.db.Exec(ctx, updateProfile, arg.UserID, arg.Username, arg.Avatar, arg.Utime)
    if err != nil {
        return 0, err
    }
    return result.RowsAffected(), nil
}
```

## 步骤 4：更新 Repo 接口和实现

**文件**：`internal/repo/user.go`

```go
package repo

// 1. 更新接口定义
type IUserRepo interface {
    // ... 已有方法
    UpdateProfile(ctx context.Context, userID, nickname, avatar string) error  // 新增
}

// 2. 添加方法实现
func (ur *UserRepo) UpdateProfile(ctx context.Context, userID, nickname, avatar string) error {
    var userUUID pgtype.UUID
    if err := userUUID.Scan(userID); err != nil {
        global.Log.Error(err)
        return ErrDefault
    }
    
    count, err := ur.userDao.UpdateProfile(ctx, user.UpdateProfileParams{
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

## 步骤 5：更新 Logic 接口和实现

**文件**：`internal/logic/user.go`

```go
package logic

// 1. 更新接口定义
type IUserLogic interface {
    // ... 已有方法
    UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileReq) (dto.UpdateProfileResp, error)  // 新增
}

// 2. 添加方法实现
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
    
    resp.Message = "资料更新成功"
    return resp, nil
}
```

## 步骤 6：添加 Handler 方法

**文件**：`internal/handler/user.go`

```go
package handler

// 添加新方法
func (uh *UserHandler) UpdateProfile(c *gin.Context) {
    // 1. 获取绑定的请求参数
    req := middleware.GetBind[dto.UpdateProfileReq](c)
    
    // 2. 获取当前用户 ID（从 JWT 中间件设置的 Context 中获取）
    userID := jwtx.GetUserID(c)
    
    // 3. 调用 Logic 层
    resp, err := uh.userLogic.UpdateProfile(c.Request.Context(), userID, req)
    
    // 4. 统一响应
    response.Response(c, resp, err)
}
```

## 步骤 7：注册路由

**文件**：`internal/router/enter.go`

```go
package router

func registerRoutes(r *gin.Engine) {
    api := r.Group("/api")
    registerUserRoutes(api.Group("/user"))
}
```

**文件**：`internal/router/user.go`

```go
package router

func registerUserRoutes(rg *gin.RouterGroup) {
    userHandler := handler.NewUserHandler()

    // 已有路由
    rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)
    rg.POST("/register", middleware.BindJsonMiddleware[dto.RegisterReq], userHandler.Register)

    // 新增路由（需要认证）
    rg.PUT("/profile",
        middleware.Authentication(jwtx.COMMON_USER),           // 认证中间件
        middleware.BindJsonMiddleware[dto.UpdateProfileReq],   // 参数绑定中间件
        userHandler.UpdateProfile,                             // Handler
    )
}
```

## 完整文件变更清单

| 文件 | 操作 |
|------|------|
| `internal/dto/user.go` | 添加 `UpdateProfileReq`, `UpdateProfileResp` |
| `internal/repo/sql/user.sql` | 添加 `UpdateProfile` 查询 |
| `internal/repo/user/user.sql.go` | 自动生成（sqlc） |
| `internal/repo/user.go` | 更新接口，添加 `UpdateProfile` 方法 |
| `internal/logic/user.go` | 更新接口，添加 `UpdateProfile` 方法 |
| `internal/handler/user.go` | 添加 `UpdateProfile` 方法 |
| `internal/router/user.go` | 注册新路由 |

## 测试

```bash
# 启动服务
go run internal/main.go

# 测试请求
curl -X PUT http://localhost:8080/api/user/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{"nickname": "新昵称", "avatar": "https://example.com/avatar.jpg"}'
```

**成功响应**：
```json
{
    "code": 0,
    "message": "OK",
    "data": {
        "message": "资料更新成功"
    }
}
```

**未登录响应**：
```json
{
    "code": -1,
    "message": "token is empty",
    "data": null
}
```

**用户不存在响应**：
```json
{
    "code": -1,
    "message": "用户不存在",
    "data": null
}
```
