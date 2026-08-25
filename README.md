# Nurture Backend

Nurture is a Go backend for infant care, family collaboration, community posts, chat, file upload, and AI-assisted growth analysis.

The codebase is now organized around detachable domain modules under `internal/<domain>`. Each module owns its DTO, handler, logic, repo, route registration, sqlc files, cache helpers, tests, and design docs when those pieces are needed. Shared infrastructure stays in `internal/pkg` and must remain business-free.

## Architecture

```text
HTTP / WS request
  -> internal/router          # top-level groups, middleware, dependency wiring
  -> internal/<domain>/route  # module route registration
  -> handler                  # bind inputs, read auth identity, call logic, respond
  -> logic                    # business rules and orchestration
  -> repo                     # database/cache persistence and domain repo errors
  -> repo/dao                 # sqlc generated PostgreSQL access
  -> repo/cache               # module-owned Redis key helpers
```

`internal/global` initializes infrastructure from `internal/config`; modules receive dependencies through their `Deps` structs. Cross-module reads use small injected interfaces instead of importing another module's internals.

## Directory Layout

```text
backend-develop-guide/       # project-specific backend development guide
deploy/
  docker-compose.yaml        # app and local infrastructure
  schema/                    # PostgreSQL schema files
internal/
  ai/                        # AI chat, RAG, growth analysis/report module
  baby/                      # baby profile, growth, vaccine, daily record module
  chat/                      # direct/group chat, session, worker, outbox module
  file/                      # file upload module
  post/                      # community post, tag, comment, recommendation module
  user/                      # account, auth, partner, follow module
  config/                    # config structs and loader
  etc/                       # local.yaml and template.yaml
  global/                    # process-wide infrastructure initialization
  middleware/                # Gin middleware
  pkg/                       # business-free infrastructure packages
  router/                    # top-level service composition and health route
```

Module design docs live with the module:

- `internal/ai/doc/README.md`
- `internal/baby/doc/README.md`
- `internal/chat/doc/README.md`
- `internal/file/doc/README.md`
- `internal/post/doc/README.md`
- `internal/user/doc/README.md`

## Technology Stack

- Go `1.24.4`
- Gin
- PostgreSQL 16+ with pgvector
- pgx/v5
- sqlc
- Redis
- RabbitMQ
- MinIO
- Viper + YAML
- Zap + Lumberjack
- JWT
- langchain-go

## Routes

- `GET /healthz`
- `/api/user/*`
- `/api/baby/*`
- `/api/post/*`
- `/api/chat/*`
- `/api/file/*`
- `/api/ai/*`
- `/api/admin/*`
- `GET /ws/chat`
- `GET /ws/group`

## Local Development

Read the project guide before changing backend code or architecture:

```bash
sed -n '1,220p' backend-develop-guide/SKILL.md
```

Create local config:

```bash
cp internal/etc/template.yaml internal/etc/local.yaml
```

For a full Docker run, start the app and infrastructure together:

```bash
docker compose -f deploy/docker-compose.yaml up -d --build
curl http://127.0.0.1:8080/healthz
```

For a host-run process, make sure enabled dependencies in `internal/etc/local.yaml` are reachable from the host, or disable the services you do not need for a quick health check:

```bash
go run ./internal
curl http://127.0.0.1:8080/healthz
```

## SQLC

Each SQL-backed module owns its sqlc config:

```bash
sqlc generate -f internal/user/repo/dao/sqlc.yaml
sqlc generate -f internal/baby/repo/dao/sqlc.yaml
sqlc generate -f internal/post/repo/dao/sqlc.yaml
sqlc generate -f internal/chat/repo/dao/sqlc.yaml
```

Generate all module DAOs:

```bash
for f in internal/*/repo/dao/sqlc.yaml; do sqlc generate -f "$f"; done
```

## Tests

Run all tests:

```bash
go test -count=1 ./...
```

Run one package or one test:

```bash
go test -count=1 -v ./internal/chat/test
go test -count=1 -v ./internal/chat/test -run TestName
```

Package-owned tests stay beside their package. `internal/test` is reserved for cross-business integration tests.

## Development Rules

- Keep `internal/pkg` business-free.
- Put detachable business domains under `internal/<domain>`.
- Prefer module-owned `constant`, `dto`, `handler`, `logic`, `repo`, `doc`, and `test` packages when a domain has its own boundary.
- Keep root docs minimal; domain design belongs in `internal/<domain>/doc`.
- Use English commit messages in `action: content` format, for example `docs: update module documentation`.
