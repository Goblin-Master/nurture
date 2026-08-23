# 数据库操作指南

## 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| 数据库 | PostgreSQL 16+ | 支持 pgvector 扩展 |
| 驱动 | pgx/v5 | 高性能 PostgreSQL 驱动 |
| 代码生成 | sqlc | 类型安全的 SQL 到 Go 代码生成 |
| 连接池 | pgxpool | pgx 内置连接池 |

## sqlc 工作流程

```
┌─────────────────────────────────────────────────────────────────┐
│  1. 定义 Schema                                                  │
│  deploy/schema/*.sql                                            │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  2. 编写 SQL 查询                                                │
│  internal/repo/sql/*.sql                                        │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  3. 运行 sqlc generate                                          │
│  sqlc generate -f internal/repo/sqlc.yaml                       │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  4. 生成 Go 代码                                                 │
│  internal/repo/{module}/                                        │
│  ├── db.go        # Queries 结构体                               │
│  ├── models.go    # 数据模型                                     │
│  └── {query}.sql.go  # 查询方法                                  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  5. 在 Repo 层使用                                               │
│  internal/repo/{module}.go                                      │
└─────────────────────────────────────────────────────────────────┘
```

## 目录结构

```
internal/repo/
├── errors.go           # Repo 层错误定义
├── user.go             # UserRepo 实现
├── order.go            # OrderRepo 实现（示例）
├── sql/                # SQL 查询文件
│   ├── user.sql
│   └── order.sql
├── sqlc.yaml           # sqlc 配置
├── user/               # sqlc 生成的 user 模块代码
│   ├── db.go
│   ├── models.go
│   └── user.sql.go
└── order/              # sqlc 生成的 order 模块代码
    ├── db.go
    ├── models.go
    └── order.sql.go

deploy/schema/
├── user.sql            # 用户表 DDL
└── order.sql           # 订单表 DDL
```

## 1. Schema 定义

### 位置

`deploy/schema/{module}.sql`

### 规范

1. 使用 `CREATE TABLE IF NOT EXISTS` 确保幂等
2. 添加表和字段注释
3. 使用合适的数据类型
4. 定义约束和索引

### 示例

```sql
-- deploy/schema/user.sql

-- 1. 扩展插件支持
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. 用户表
CREATE TABLE IF NOT EXISTS "user" (
    id        BIGSERIAL PRIMARY KEY,           -- 主键ID（自增）
    user_id   UUID UNIQUE NOT NULL,            -- 用户ID（UUID）
    ctime     BIGINT NOT NULL,                 -- 创建时间（毫秒时间戳）
    utime     BIGINT NOT NULL,                 -- 更新时间（毫秒时间戳）
    account   VARCHAR(20) UNIQUE NOT NULL,     -- 账号
    password  VARCHAR(20) NOT NULL,            -- 密码
    email     VARCHAR(50) UNIQUE NOT NULL,     -- 邮箱
    username  VARCHAR(20) NOT NULL,            -- 用户名
    avatar    VARCHAR(255) NOT NULL,           -- 头像URL
    role      SMALLINT NOT NULL DEFAULT 1      -- 角色（1=普通用户）
);

-- 表注释
COMMENT ON TABLE "user" IS '用户表';

-- 字段注释
COMMENT ON COLUMN "user".id IS '主键ID';
COMMENT ON COLUMN "user".user_id IS '用户ID';
COMMENT ON COLUMN "user".ctime IS '创建时间';
COMMENT ON COLUMN "user".utime IS '更新时间';
COMMENT ON COLUMN "user".account IS '账号';
COMMENT ON COLUMN "user".password IS '密码';
COMMENT ON COLUMN "user".email IS '邮箱';
COMMENT ON COLUMN "user".username IS '用户名';
COMMENT ON COLUMN "user".avatar IS '头像';
COMMENT ON COLUMN "user".role IS '角色';
```

### 字段类型选择

| 场景 | PostgreSQL 类型 | Go 类型 |
|------|-----------------|---------|
| 主键 | `BIGSERIAL` | `int64` |
| UUID | `UUID` | `pgtype.UUID` |
| 时间戳 | `BIGINT` | `int64` |
| 短文本 | `VARCHAR(n)` | `string` |
| 长文本 | `TEXT` | `string` |
| 整数 | `INTEGER` / `BIGINT` | `int32` / `int64` |
| 小整数 | `SMALLINT` | `int16` |
| 布尔 | `BOOLEAN` | `bool` |
| JSON | `JSONB` | `[]byte` 或自定义 |

## 2. SQL 查询编写

### 位置

`internal/repo/sql/{module}.sql`

### sqlc 注释规范

```sql
-- name: 方法名 :返回类型
```

### 返回类型

| 类型 | 说明 | 返回值 |
|------|------|--------|
| `:one` | 返回单行 | `(Model, error)` |
| `:many` | 返回多行 | `([]Model, error)` |
| `:exec` | 不返回结果 | `error` |
| `:execrows` | 返回影响行数 | `(int64, error)` |
| `:execresult` | 返回完整结果 | `(pgconn.CommandTag, error)` |

### 示例

```sql
-- internal/repo/sql/user.sql

-- 查询单行
-- name: GetUserByAccountAndPassword :one
SELECT * FROM "user"
WHERE account = $1 AND password = $2 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM "user"
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM "user"
WHERE user_id = $1 LIMIT 1;

-- 插入（不返回结果）
-- name: CreateUser :exec
INSERT INTO "user" (
    user_id, ctime, utime, account, password, email, username, avatar, role
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- 更新（返回影响行数）
-- name: UpdatePasswordByEmail :execrows
UPDATE "user"
SET password = $2
WHERE email = $1;

-- name: UpdateAvatarByUserID :execrows
UPDATE "user"
SET avatar = $2
WHERE user_id = $1;

-- name: UpdateProfile :execrows
UPDATE "user"
SET username = $2, avatar = $3, utime = $4
WHERE user_id = $1;

-- 查询多行
-- name: ListUsers :many
SELECT * FROM "user"
ORDER BY ctime DESC
LIMIT $1 OFFSET $2;

-- 删除
-- name: DeleteUserByID :execrows
DELETE FROM "user"
WHERE user_id = $1;
```

## 3. sqlc 配置

### 位置

`internal/repo/sqlc.yaml`

### 配置示例

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
```

### 添加新模块

当添加新业务模块时，在 `sql` 数组中添加新配置：

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
  
  - engine: "postgresql"
    queries: "sql/order.sql"
    schema: "../../deploy/schema/order.sql"
    gen:
      go:
        package: "order"
        out: "order"
        sql_package: "pgx/v5"
```

## 4. 代码生成

### 安装 sqlc

```bash
# macOS
brew install sqlc

# Go install
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 生成命令

```bash
# 在项目根目录执行
sqlc generate -f internal/repo/sqlc.yaml

# 或在 internal/repo 目录执行
cd internal/repo
sqlc generate
```

### 生成文件说明

生成后会在 `internal/repo/{module}/` 目录下创建以下文件：

**db.go**：
```go
// Code generated by sqlc. DO NOT EDIT.

package user

type DBTX interface {
    Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
    Query(context.Context, string, ...interface{}) (pgx.Rows, error)
    QueryRow(context.Context, string, ...interface{}) pgx.Row
}

func New(db DBTX) *Queries {
    return &Queries{db: db}
}

type Queries struct {
    db DBTX
}

func (q *Queries) WithTx(tx pgx.Tx) *Queries {
    return &Queries{db: tx}
}
```

**models.go**：
```go
// Code generated by sqlc. DO NOT EDIT.

package user

type User struct {
    ID       int64
    UserID   pgtype.UUID
    Ctime    int64
    Utime    int64
    Account  string
    Password string
    Email    string
    Username string
    Avatar   string
    Role     int16
}
```

**user.sql.go**：
```go
// Code generated by sqlc. DO NOT EDIT.

package user

const createUser = `-- name: CreateUser :exec
INSERT INTO "user" (...) VALUES (...)
`

type CreateUserParams struct {
    UserID   pgtype.UUID
    Ctime    int64
    // ...
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
    _, err := q.db.Exec(ctx, createUser, ...)
    return err
}
```

## 5. Repo 层使用

### Repo 结构

```go
// internal/repo/user.go
package repo

import (
    "nurture/internal/global"
    "nurture/internal/repo/user"
)

type IUserRepo interface {
    LoginWithAccount(ctx context.Context, account, password string) (user.User, error)
    Register(ctx context.Context, userID, username, email, account, password string) error
}

type UserRepo struct {
    userDao *user.Queries
}

func NewUserRepo() *UserRepo {
    return &UserRepo{
        userDao: user.New(global.DB),
    }
}

var _ IUserRepo = (*UserRepo)(nil)
```

### 方法实现

```go
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

func (ur *UserRepo) Register(ctx context.Context, userID, username, email, account, password string) error {
    var userUUID pgtype.UUID
    if err := userUUID.Scan(userID); err != nil {
        return err
    }

    err := ur.userDao.CreateUser(ctx, user.CreateUserParams{
        UserID:   userUUID,
        Username: username,
        Email:    email,
        Account:  account,
        Password: password,
        Ctime:    time.Now().UnixMilli(),
        Utime:    time.Now().UnixMilli(),
        Avatar:   "",
        Role:     1,
    })

    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
```

## 6. UUID 处理

### pgtype.UUID 使用

```go
import "github.com/jackc/pgx/v5/pgtype"

// 字符串转 pgtype.UUID
var userUUID pgtype.UUID
if err := userUUID.Scan(userIDString); err != nil {
    return err
}

// pgtype.UUID 转字符串
userIDString := userUUID.String()
```

## 7. 事务处理

### 使用 WithTx

```go
func (ur *UserRepo) TransferBalance(ctx context.Context, fromID, toID string, amount int64) error {
    tx, err := global.DB.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // 使用事务创建新的 Queries
    qtx := ur.userDao.WithTx(tx)

    // 执行事务内的操作
    err = qtx.DeductBalance(ctx, user.DeductBalanceParams{
        UserID: fromID,
        Amount: amount,
    })
    if err != nil {
        return err
    }

    err = qtx.AddBalance(ctx, user.AddBalanceParams{
        UserID: toID,
        Amount: amount,
    })
    if err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

## 8. 数据库连接初始化

### 位置

`internal/pkg/pgsqlx/enter.go`

### 实现

```go
package pgsqlx

import (
    "context"
    "fmt"
    "nurture/internal/config"

    "github.com/jackc/pgx/v5/pgxpool"
)

func InitPgsql() *pgxpool.Pool {
    dsn := config.Conf.DB.DSN()

    poolConfig, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        panic(fmt.Sprintf("parse pgsql config error: %v", err))
    }

    pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
    if err != nil {
        panic(fmt.Sprintf("connect pgsql error: %v", err))
    }

    if err := pool.Ping(context.Background()); err != nil {
        panic(fmt.Sprintf("ping pgsql error: %v", err))
    }

    return pool
}
```

## 9. 新增模块完整流程

以添加「订单」模块为例：

本节适用于共享三层目录下的新业务。如果业务应拆成 `internal/<domain>` 可拆卸模块，先按 `module-boundary.md` 判断边界，并把 SQL 和 sqlc 配置放到模块自己的 `repo/dao` 下。

### Step 1: 创建 Schema

```sql
-- deploy/schema/order.sql
CREATE TABLE IF NOT EXISTS "order" (
    id         BIGSERIAL PRIMARY KEY,
    order_id   UUID UNIQUE NOT NULL,
    user_id    UUID NOT NULL REFERENCES "user"(user_id),
    amount     BIGINT NOT NULL,
    status     SMALLINT NOT NULL DEFAULT 1,
    ctime      BIGINT NOT NULL,
    utime      BIGINT NOT NULL
);

COMMENT ON TABLE "order" IS '订单表';
```

### Step 2: 编写 SQL 查询

```sql
-- internal/repo/sql/order.sql

-- name: CreateOrder :exec
INSERT INTO "order" (order_id, user_id, amount, status, ctime, utime)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetOrderByID :one
SELECT * FROM "order" WHERE order_id = $1;

-- name: ListOrdersByUserID :many
SELECT * FROM "order" WHERE user_id = $1 ORDER BY ctime DESC;

-- name: UpdateOrderStatus :execrows
UPDATE "order" SET status = $2, utime = $3 WHERE order_id = $1;
```

### Step 3: 更新 sqlc.yaml

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
  
  - engine: "postgresql"
    queries: "sql/order.sql"
    schema: "../../deploy/schema/order.sql"
    gen:
      go:
        package: "order"
        out: "order"
        sql_package: "pgx/v5"
```

### Step 4: 生成代码

```bash
sqlc generate -f internal/repo/sqlc.yaml
```

### Step 5: 实现 Repo

```go
// internal/repo/order.go
package repo

type IOrderRepo interface {
    CreateOrder(ctx context.Context, order Order) error
    GetOrderByID(ctx context.Context, orderID string) (order.Order, error)
}

type OrderRepo struct {
    orderDao *order.Queries
}

func NewOrderRepo() *OrderRepo {
    return &OrderRepo{
        orderDao: order.New(global.DB),
    }
}

var _ IOrderRepo = (*OrderRepo)(nil)

// 实现方法...
```
