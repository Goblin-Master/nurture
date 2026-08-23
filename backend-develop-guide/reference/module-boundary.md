# 模块边界与可拆卸架构

本文档说明什么时候继续使用共享三层目录，什么时候拆成 `internal/<domain>` 可拆卸模块。

## 核心原则

- 默认从共享三层开始：`internal/handler`、`internal/logic`、`internal/repo`、`internal/dto`。
- 当业务边界稳定且复杂度开始外溢时，再拆成 `internal/<domain>`。
- 模块化的目标是让模块可以独立理解、独立测试、低成本删除，不是为了提前套目录。
- `internal/pkg` 只放基础设施和横切能力，不放业务语义。

## 继续使用共享三层的情况

满足以下情况时，不要拆模块：

- 只有少量 CRUD 或简单查询。
- 业务没有自己的会话、worker、队列、缓存策略、复杂权限或状态机。
- 代码主要复用现有全局 DTO、handler、logic、repo。
- SQL 和 repo 逻辑很薄，不会形成独立数据边界。
- 未来拆除这个功能时，仍然要保留大量公共代码。

推荐结构：

```text
internal/
  dto/<domain>.go
  handler/<domain>.go
  logic/<domain>.go
  repo/<domain>.go
  repo/sql/<domain>.sql
```

## 应该模块化的信号

出现多个信号时，优先考虑 `internal/<domain>`：

- 业务有清晰名词，可以一句话描述边界，例如 `chat`、`payment`、`notification`。
- 业务拥有独立入口：HTTP routes、WS routes、worker、consumer、scheduler 或 outbox。
- 业务拥有独立持久化模型，SQL、dao、repo 与其它业务交叉少。
- 业务有自己的状态流，例如消息投递、订单生命周期、任务重试。
- handler、logic、repo 文件过长，继续放共享目录会导致命名和职责混乱。
- 测试需要独立 fakes、fixtures、session、worker 或 repo cache。
- 未来可能整体替换、下线或抽成服务。

## 模块目标结构

模块目录按实际需要增减，不为空建目录。

```text
internal/<domain>/
  enter.go          # Module/Deps/NewModule
  route.go          # RegisterRoutes
  constant/         # 固定业务策略和常量
  dto/              # 模块请求/响应/消息 DTO
  handler/          # HTTP/WS handler
  logic/            # 业务规则和编排
  repo/             # repo facade and domain errors
    dao/            # sqlc 生成代码
    cache/          # 模块缓存
  worker/           # consumer/scheduler/outbox worker
  session/          # 连接、订阅、在线会话
  doc/              # 模块链路和设计说明
  test/             # 模块测试，package 使用 test
```

`repo` 下的包直接叫 `dao` 和 `cache`，不要再包一层无意义目录。

## 模块入口

模块入口负责收敛依赖，模块内部禁止直接依赖 `global`。

```go
type Deps struct {
    DB  *pgxpool.Pool
    RDB redis.Cmdable
    Log *zap.SugaredLogger
}

type Module struct {
    handler *handler.Handler
    worker  *worker.Worker
}

func NewModule(deps Deps) *Module {
    repo := repo.NewRepo(deps.DB, deps.RDB, deps.Log)
    logic := logic.NewLogic(repo)
    return &Module{
        handler: handler.NewHandler(logic),
        worker:  worker.NewWorker(repo, deps.Log),
    }
}
```

`main` 或 `router` 做顶层组装：

```go
api := r.Group("/api")
ws := r.Group("/ws")

chat.NewModule(chat.Deps{
    DB:  global.DB,
    RDB: global.RDB,
    Log: global.Log,
}).RegisterRoutes(api.Group("/chat"), ws)
```

## 依赖规则

- `internal/<domain>` 可以依赖 `internal/pkg`、`internal/middleware`、`internal/config` 的类型和值。
- `internal/<domain>` 不依赖其它业务模块。需要交互时，通过上层编排或明确接口注入。
- `internal/pkg` 不依赖 `internal/<domain>`、`internal/logic`、`internal/repo`、`internal/handler`。
- 能注入的基础设施优先注入：DB、Redis、RabbitMQ、logger、token parser、middleware。
- 不要为了所有东西都可注入而包一层无意义接口；接口定义在使用方，且只定义真实需要的方法。

## 配置与常量

- 环境相关、部署相关、密钥相关内容放 `internal/config` 和配置文件。
- 固定业务策略放模块 `constant`，例如消息类型、默认分页上限、限流 key。
- 模块只有在确实有环境差异配置时才新增 `internal/<domain>/config`。
- DTO 不能承载配置结构。

## SQLC 迁移规则

模块拥有独立 SQL 边界时，sqlc 跟随模块移动。

```text
internal/<domain>/repo/dao/
  sql/<domain>.sql
  sqlc.yaml
  db.go
  models.go
  <domain>.sql.go
```

模块 `sqlc.yaml` 示例：

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "sql/<domain>.sql"
    schema:
      - "../../../../deploy/schema/<domain>.sql"
      - "../../../../deploy/schema/user.sql"
    gen:
      go:
        package: "dao"
        out: "."
        sql_package: "pgx/v5"
```

迁移时需要：

- 从共享 `internal/repo/sqlc.yaml` 移除该业务 SQL 配置。
- 删除旧生成目录，重新运行共享和模块 sqlc。
- 保持 deploy schema 仍在 `deploy/schema`，不要复制 schema 到模块。
- Repo 只暴露领域方法，不把 sqlc 生成类型泄漏到 handler 或 logic。

## 错误流

- repo 错误放 `internal/<domain>/repo/errors.go`。
- logic 错误放 `internal/<domain>/logic/errors.go`。
- repo 记录技术错误并返回 repo error。
- logic 用 `errors.Is` 映射 repo error，返回用户可读业务错误。
- handler 不直接判断 repo error，只调用 logic 并统一响应。
- worker/session 等基础流程只处理自己的职责，不判断业务权限，权限判断放 logic。

## 路由规则

- 模块暴露 `RegisterRoutes`，顶层 router 只负责分组和注入中间件。
- HTTP API 使用业务名词，例如 `/api/chat/groups`。
- WS 路径也使用业务名词，不使用协议名当业务名，例如 `/ws/chat`、`/ws/group`。
- 同一个概念路径单复数要统一。资源集合用复数，单一连接或操作入口用单数。

## 测试规则

- 模块测试放 `internal/<domain>/test`，package 使用 `test`。
- pkg 测试放各自包下，例如 `internal/pkg/ratelimitx/ratelimitx_test.go`。
- `internal/test` 只放跨业务层联调测试。
- 模块测试可使用 fakes 验证 logic、session、worker 行为，不强依赖真实 DB/MQ。
- SQLC 和 repo 行为变化需要覆盖 SQL 生成、repo error 映射或集成测试。

## 迁移清单

1. 确认模块化信号，不为简单 CRUD 提前拆模块。
2. 建立 `internal/<domain>` 入口和路由注册。
3. 迁移 dto、handler、logic、repo、constant、test。
4. 如果有独立 SQL，迁移 sqlc 配置和生成代码。
5. 通过 `Deps` 注入 DB、Redis、logger、middleware、外部 client。
6. 删除共享目录里的旧业务文件和旧路由注册。
7. 运行 `sqlc generate`、`go test ./...`，并检查路径、错误流、测试位置。
8. 补充模块 `doc/`，记录关键链路和拆除边界。

## 反例

- 把业务代码放进 `internal/pkg`，例如 `pkg/chatx`、`pkg/wsx`。
- 为只有一个方法的简单 CRUD 提前建完整模块目录。
- 模块内部直接读写 `global.DB`、`global.RDB`、`global.Log`。
- handler 判断 repo 错误或直接操作数据库。
- sqlc 生成代码留在共享 `internal/repo`，业务 repo 却放进模块。
- 为了模块化复制公共中间件、响应函数或基础设施 client。
