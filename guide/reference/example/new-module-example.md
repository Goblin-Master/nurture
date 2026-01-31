# 新增业务模块完整示例

本示例展示如何从零开始添加一个新的业务模块。

## 场景

添加「文章」模块，包含以下功能：

- 创建文章
- 获取文章详情
- 获取文章列表

## 步骤 1：创建数据库表

**文件**：`deploy/schema/article.sql`

```sql
-- 文章表
CREATE TABLE IF NOT EXISTS "article" (
    id           BIGSERIAL PRIMARY KEY,
    article_id   UUID UNIQUE NOT NULL,
    user_id      UUID NOT NULL,
    title        VARCHAR(100) NOT NULL,
    content      TEXT NOT NULL,
    status       SMALLINT NOT NULL DEFAULT 1,  -- 1=草稿, 2=已发布
    view_count   BIGINT NOT NULL DEFAULT 0,
    ctime        BIGINT NOT NULL,
    utime        BIGINT NOT NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_article_user_id ON "article"(user_id);
CREATE INDEX IF NOT EXISTS idx_article_status ON "article"(status);
CREATE INDEX IF NOT EXISTS idx_article_ctime ON "article"(ctime DESC);

-- 注释
COMMENT ON TABLE "article" IS '文章表';
COMMENT ON COLUMN "article".id IS '主键ID';
COMMENT ON COLUMN "article".article_id IS '文章ID';
COMMENT ON COLUMN "article".user_id IS '作者ID';
COMMENT ON COLUMN "article".title IS '标题';
COMMENT ON COLUMN "article".content IS '内容';
COMMENT ON COLUMN "article".status IS '状态：1=草稿, 2=已发布';
COMMENT ON COLUMN "article".view_count IS '浏览量';
COMMENT ON COLUMN "article".ctime IS '创建时间';
COMMENT ON COLUMN "article".utime IS '更新时间';
```

执行 SQL 创建表：

```bash
# 通过 Docker
docker exec -i nurture-pg psql -U nurture -d nurture < deploy/schema/article.sql

# 或直接连接数据库执行
```

## 步骤 2：定义 DTO

**文件**：`internal/dto/article.go`（新建）

```go
package dto

// 创建文章
type (
    CreateArticleReq struct {
        Title   string `json:"title" binding:"required,max=100"`
        Content string `json:"content" binding:"required"`
        Status  int16  `json:"status" binding:"oneof=1 2"`  // 1=草稿, 2=发布
    }
    CreateArticleResp struct {
        ArticleID string `json:"article_id"`
        Message   string `json:"message"`
    }
)

// 获取文章详情
type (
    GetArticleReq struct {
        ArticleID string `uri:"article_id" binding:"required,uuid"`
    }
    GetArticleResp struct {
        ArticleID string `json:"article_id"`
        Title     string `json:"title"`
        Content   string `json:"content"`
        Status    int16  `json:"status"`
        ViewCount int64  `json:"view_count"`
        AuthorID  string `json:"author_id"`
        Ctime     int64  `json:"ctime"`
    }
)

// 获取文章列表
type (
    ListArticlesReq struct {
        Page     int `form:"page" binding:"required,min=1"`
        PageSize int `form:"page_size" binding:"required,min=1,max=50"`
    }
    ListArticlesResp struct {
        Total    int64         `json:"total"`
        Articles []ArticleItem `json:"articles"`
    }
    ArticleItem struct {
        ArticleID string `json:"article_id"`
        Title     string `json:"title"`
        Status    int16  `json:"status"`
        ViewCount int64  `json:"view_count"`
        Ctime     int64  `json:"ctime"`
    }
)
```

## 步骤 3：编写 SQL 查询

**文件**：`internal/repo/sql/article.sql`（新建）

```sql
-- name: CreateArticle :exec
INSERT INTO "article" (
    article_id, user_id, title, content, status, ctime, utime
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetArticleByID :one
SELECT * FROM "article"
WHERE article_id = $1;

-- name: ListArticles :many
SELECT * FROM "article"
WHERE status = 2
ORDER BY ctime DESC
LIMIT $1 OFFSET $2;

-- name: CountArticles :one
SELECT COUNT(*) FROM "article"
WHERE status = 2;

-- name: IncrementViewCount :exec
UPDATE "article"
SET view_count = view_count + 1
WHERE article_id = $1;

-- name: UpdateArticle :execrows
UPDATE "article"
SET title = $2, content = $3, status = $4, utime = $5
WHERE article_id = $1;

-- name: DeleteArticle :execrows
DELETE FROM "article"
WHERE article_id = $1;
```

## 步骤 4：更新 sqlc 配置

**文件**：`internal/repo/sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "sql/user.sql"
    schema: "../../deploy/schema/user.sql"
    gen:
      go:
        package: "user"
        out: "user"
        sql_package: "pgx/v5"
  
  # 新增 article 模块配置
  - engine: "postgresql"
    queries: "sql/article.sql"
    schema: "../../deploy/schema/article.sql"
    gen:
      go:
        package: "article"
        out: "article"
        sql_package: "pgx/v5"
```

## 步骤 5：生成 sqlc 代码

```bash
sqlc generate -f internal/repo/sqlc.yaml
```

生成以下文件：
- `internal/repo/article/db.go`
- `internal/repo/article/models.go`
- `internal/repo/article/article.sql.go`

## 步骤 6：添加 Repo 层错误

**文件**：`internal/repo/errors.go`

```go
package repo

import "errors"

var (
    ErrDefault = errors.New("默认错误")
)

var (
    // User 相关错误
    ErrUserNotExist  = errors.New("用户不存在")
    ErrEmailIsUsed   = errors.New("邮箱已经被使用")
    ErrAccountIsUsed = errors.New("账号已经被使用")
)

// 新增 Article 相关错误
var (
    ErrArticleNotExist = errors.New("文章不存在")
)
```

## 步骤 7：实现 Repo 层

**文件**：`internal/repo/article.go`（新建）

```go
package repo

import (
    "context"
    "errors"
    "nurture/internal/global"
    "nurture/internal/repo/article"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
)

type IArticleRepo interface {
    CreateArticle(ctx context.Context, articleID, userID, title, content string, status int16) error
    GetArticleByID(ctx context.Context, articleID string) (article.Article, error)
    ListArticles(ctx context.Context, limit, offset int) ([]article.Article, error)
    CountArticles(ctx context.Context) (int64, error)
    IncrementViewCount(ctx context.Context, articleID string) error
}

type ArticleRepo struct {
    articleDao *article.Queries
}

func NewArticleRepo() *ArticleRepo {
    return &ArticleRepo{
        articleDao: article.New(global.DB),
    }
}

var _ IArticleRepo = (*ArticleRepo)(nil)

func (ar *ArticleRepo) CreateArticle(ctx context.Context, articleID, userID, title, content string, status int16) error {
    var articleUUID, userUUID pgtype.UUID
    if err := articleUUID.Scan(articleID); err != nil {
        return err
    }
    if err := userUUID.Scan(userID); err != nil {
        return err
    }

    now := time.Now().UnixMilli()
    err := ar.articleDao.CreateArticle(ctx, article.CreateArticleParams{
        ArticleID: articleUUID,
        UserID:    userUUID,
        Title:     title,
        Content:   content,
        Status:    status,
        Ctime:     now,
        Utime:     now,
    })
    if err != nil {
        global.Log.Error(err)
        return ErrDefault
    }
    return nil
}

func (ar *ArticleRepo) GetArticleByID(ctx context.Context, articleID string) (article.Article, error) {
    var articleUUID pgtype.UUID
    if err := articleUUID.Scan(articleID); err != nil {
        return article.Article{}, err
    }

    a, err := ar.articleDao.GetArticleByID(ctx, articleUUID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return article.Article{}, ErrArticleNotExist
        }
        global.Log.Error(err)
        return article.Article{}, ErrDefault
    }
    return a, nil
}

func (ar *ArticleRepo) ListArticles(ctx context.Context, limit, offset int) ([]article.Article, error) {
    articles, err := ar.articleDao.ListArticles(ctx, article.ListArticlesParams{
        Limit:  int32(limit),
        Offset: int32(offset),
    })
    if err != nil {
        global.Log.Error(err)
        return nil, ErrDefault
    }
    return articles, nil
}

func (ar *ArticleRepo) CountArticles(ctx context.Context) (int64, error) {
    count, err := ar.articleDao.CountArticles(ctx)
    if err != nil {
        global.Log.Error(err)
        return 0, ErrDefault
    }
    return count, nil
}

func (ar *ArticleRepo) IncrementViewCount(ctx context.Context, articleID string) error {
    var articleUUID pgtype.UUID
    if err := articleUUID.Scan(articleID); err != nil {
        return err
    }
    return ar.articleDao.IncrementViewCount(ctx, articleUUID)
}
```

## 步骤 8：添加 Logic 层错误

**文件**：`internal/logic/errors.go`

```go
package logic

// 新增 Article 相关错误
var (
    ErrArticleNotExist = errors.New("文章不存在")
    ErrArticleNoPermission = errors.New("无权限操作此文章")
)
```

## 步骤 9：实现 Logic 层

**文件**：`internal/logic/article.go`（新建）

```go
package logic

import (
    "context"
    "errors"
    "nurture/internal/dto"
    "nurture/internal/global"
    "nurture/internal/repo"

    "github.com/google/uuid"
)

type IArticleLogic interface {
    CreateArticle(ctx context.Context, userID string, req dto.CreateArticleReq) (dto.CreateArticleResp, error)
    GetArticle(ctx context.Context, req dto.GetArticleReq) (dto.GetArticleResp, error)
    ListArticles(ctx context.Context, req dto.ListArticlesReq) (dto.ListArticlesResp, error)
}

type ArticleLogic struct {
    articleRepo *repo.ArticleRepo
}

func NewArticleLogic() *ArticleLogic {
    return &ArticleLogic{
        articleRepo: repo.NewArticleRepo(),
    }
}

var _ IArticleLogic = (*ArticleLogic)(nil)

func (al *ArticleLogic) CreateArticle(ctx context.Context, userID string, req dto.CreateArticleReq) (dto.CreateArticleResp, error) {
    var resp dto.CreateArticleResp

    articleID := uuid.NewString()
    err := al.articleRepo.CreateArticle(ctx, articleID, userID, req.Title, req.Content, req.Status)
    if err != nil {
        global.Log.Error(err)
        return resp, ErrDefault
    }

    resp.ArticleID = articleID
    resp.Message = "文章创建成功"
    return resp, nil
}

func (al *ArticleLogic) GetArticle(ctx context.Context, req dto.GetArticleReq) (dto.GetArticleResp, error) {
    var resp dto.GetArticleResp

    article, err := al.articleRepo.GetArticleByID(ctx, req.ArticleID)
    if err != nil {
        if errors.Is(err, repo.ErrArticleNotExist) {
            return resp, ErrArticleNotExist
        }
        global.Log.Error(err)
        return resp, ErrDefault
    }

    // 增加浏览量（异步，不阻塞响应）
    go func() {
        _ = al.articleRepo.IncrementViewCount(context.Background(), req.ArticleID)
    }()

    resp.ArticleID = article.ArticleID.String()
    resp.Title = article.Title
    resp.Content = article.Content
    resp.Status = article.Status
    resp.ViewCount = article.ViewCount
    resp.AuthorID = article.UserID.String()
    resp.Ctime = article.Ctime

    return resp, nil
}

func (al *ArticleLogic) ListArticles(ctx context.Context, req dto.ListArticlesReq) (dto.ListArticlesResp, error) {
    var resp dto.ListArticlesResp

    offset := (req.Page - 1) * req.PageSize

    // 获取总数
    total, err := al.articleRepo.CountArticles(ctx)
    if err != nil {
        global.Log.Error(err)
        return resp, ErrDefault
    }

    // 获取列表
    articles, err := al.articleRepo.ListArticles(ctx, req.PageSize, offset)
    if err != nil {
        global.Log.Error(err)
        return resp, ErrDefault
    }

    resp.Total = total
    resp.Articles = make([]dto.ArticleItem, len(articles))
    for i, a := range articles {
        resp.Articles[i] = dto.ArticleItem{
            ArticleID: a.ArticleID.String(),
            Title:     a.Title,
            Status:    a.Status,
            ViewCount: a.ViewCount,
            Ctime:     a.Ctime,
        }
    }

    return resp, nil
}
```

## 步骤 10：实现 Handler 层

**文件**：`internal/handler/article.go`（新建）

```go
package handler

import (
    "nurture/internal/dto"
    "nurture/internal/logic"
    "nurture/internal/middleware"
    "nurture/internal/pkg/jwtx"
    "nurture/internal/pkg/response"

    "github.com/gin-gonic/gin"
)

type ArticleHandler struct {
    articleLogic *logic.ArticleLogic
}

func NewArticleHandler() *ArticleHandler {
    return &ArticleHandler{
        articleLogic: logic.NewArticleLogic(),
    }
}

func (ah *ArticleHandler) CreateArticle(c *gin.Context) {
    req := middleware.GetBind[dto.CreateArticleReq](c)
    userID := jwtx.GetUserID(c)
    
    resp, err := ah.articleLogic.CreateArticle(c.Request.Context(), userID, req)
    response.Response(c, resp, err)
}

func (ah *ArticleHandler) GetArticle(c *gin.Context) {
    req := middleware.GetBind[dto.GetArticleReq](c)
    
    resp, err := ah.articleLogic.GetArticle(c.Request.Context(), req)
    response.Response(c, resp, err)
}

func (ah *ArticleHandler) ListArticles(c *gin.Context) {
    req := middleware.GetBind[dto.ListArticlesReq](c)
    
    resp, err := ah.articleLogic.ListArticles(c.Request.Context(), req)
    response.Response(c, resp, err)
}
```

## 步骤 11：更新路由管理器

**文件**：`internal/manger/enter.go`

```go
package manager

type RouteManager struct {
    CommonRoutes  *gin.RouterGroup
    UserRoutes    *gin.RouterGroup
    ArticleRoutes *gin.RouterGroup  // 新增
}

func NewRouteManager(router *gin.Engine) *RouteManager {
    return &RouteManager{
        CommonRoutes:  router.Group("/api/common"),
        UserRoutes:    router.Group("/api/user"),
        ArticleRoutes: router.Group("/api/article"),  // 新增
    }
}

// 新增方法
func (rm *RouteManager) RegisterArticleRoutes(handler PathHandler) {
    handler(rm.ArticleRoutes)
}
```

## 步骤 12：注册路由

**文件**：`internal/router/enter.go`

```go
package router

func registerRoutes(routeManager *manager.RouteManager) {
    // ... 已有路由

    // 新增 Article 路由
    routeManager.RegisterArticleRoutes(func(rg *gin.RouterGroup) {
        articleHandler := handler.NewArticleHandler()
        
        // 创建文章（需要登录）
        rg.POST("", 
            middleware.Authentication(jwtx.COMMON_USER),
            middleware.BindJsonMiddleware[dto.CreateArticleReq],
            articleHandler.CreateArticle,
        )
        
        // 获取文章详情（公开）
        rg.GET("/:article_id", 
            middleware.BindUriMiddleware[dto.GetArticleReq],
            articleHandler.GetArticle,
        )
        
        // 获取文章列表（公开）
        rg.GET("", 
            middleware.BindQueryMiddleware[dto.ListArticlesReq],
            articleHandler.ListArticles,
        )
    })
}
```

## 完整文件清单

### 新建文件

| 文件 | 说明 |
|------|------|
| `deploy/schema/article.sql` | 数据库表定义 |
| `internal/dto/article.go` | 请求/响应 DTO |
| `internal/repo/sql/article.sql` | SQL 查询 |
| `internal/repo/article/` | sqlc 生成（自动） |
| `internal/repo/article.go` | Repo 实现 |
| `internal/logic/article.go` | Logic 实现 |
| `internal/handler/article.go` | Handler 实现 |

### 修改文件

| 文件 | 修改内容 |
|------|----------|
| `internal/repo/sqlc.yaml` | 添加 article 模块配置 |
| `internal/repo/errors.go` | 添加 `ErrArticleNotExist` |
| `internal/logic/errors.go` | 添加文章相关错误 |
| `internal/manger/enter.go` | 添加 `ArticleRoutes` 和注册方法 |
| `internal/router/enter.go` | 注册 article 路由 |

## 测试

```bash
# 创建文章
curl -X POST http://localhost:8080/api/article \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title": "第一篇文章", "content": "这是内容", "status": 2}'

# 获取文章
curl http://localhost:8080/api/article/<article_id>

# 获取列表
curl "http://localhost:8080/api/article?page=1&page_size=10"
```