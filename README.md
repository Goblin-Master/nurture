# Nurture Backend Project

Welcome to the **Nurture** project! This is a high-performance backend service built with Go (Gin), designed with a clean layered architecture and integrated with modern development tools. This document serves as a comprehensive guide for developers to understand the project structure, development workflow, and coding standards.

## 📚 Table of Contents

- [Project Architecture](#-project-architecture)
- [Directory Structure](#-directory-structure)
- [Technology Stack](#-technology-stack)
- [Development Workflow](#-development-workflow)
  - [Prerequisites](#prerequisites)
  - [Setup & Run](#setup--run)
  - [API Development Guide](#api-development-guide)
  - [Database Operations (sqlc)](#database-operations-sqlc)
- [Error Handling](#-error-handling)
- [Git Commit Convention](#-git-commit-convention)
- [Deployment](#-deployment)

---

## 🏗 Project Architecture

This project follows a classic **Layered Architecture** to ensure separation of concerns, maintainability, and scalability.

### Layers Overview

1.  **Handler Layer (`internal/handler`)**
    *   **Role**: Entry point for HTTP requests.
    *   **Responsibility**:
        *   Parse request parameters (Query, Path, Body) using `middleware.GetBind`.
        *   Validate input data.
        *   Call the corresponding method in the **Logic Layer**.
        *   Format and send HTTP responses using `pkg/response`.
    *   **Rule**: No business logic should exist here; only request/response handling.

2.  **Logic Layer (`internal/logic`)**
    *   **Role**: Core business logic container.
    *   **Responsibility**:
        *   Implement business rules and workflows.
        *   Orchestrate calls to the **Repo Layer** and external services (Email, etc.).
        *   Handle business errors and convert them to user-friendly messages.
    *   **Rule**: Independent of HTTP context (Gin). Accepts DTOs, returns DTOs/Errors.

3.  **Repository Layer (`internal/repo`)**
    *   **Role**: Data Access Layer (DAL).
    *   **Responsibility**:
        *   Interact directly with the database (PostgreSQL) and Cache (Redis).
        *   Use `sqlc` generated code for type-safe database operations.
        *   Handle database-specific errors (e.g., converting `pgx.ErrNoRows` to domain errors).

4.  **Infrastructure/Package Layer (`internal/pkg`)**
    *   **Role**: Shared technical components.
    *   **Components**:
        *   `pgsqlx`: Database connection pool initialization.
        *   `redisx`: Redis client initialization.
        *   `jwtx`: JWT token generation and parsing.
        *   `emailx`: Email sending service.
        *   `zapx`: Logging configuration.

---

## 📂 Directory Structure

```text
├── deploy                  # Deployment configurations
│   ├── docker-compose.yaml # Container orchestration
│   └── schema              # Database schema definitions (SQL)
├── internal
│   ├── config              # Configuration loading and struct definitions
│   ├── constant            # Global constants
│   ├── dto                 # Data Transfer Objects (Request/Response structs)
│   ├── etc                 # Configuration files (local.yaml, template.yaml)
│   ├── global              # Global instances (DB, Redis, Logger)
│   ├── handler             # HTTP Handlers (Controller)
│   ├── logic               # Business Logic (Service)
│   │   ├── errors.go       # Logic-layer specific errors
│   │   └── ...
│   ├── manager             # Router & Middleware management
│   ├── middleware          # Gin Middlewares (Auth, CORS, Bind)
│   ├── pkg                 # Infrastructure packages (Email, JWT, DB, etc.)
│   ├── repo                # Data Access Layer
│   │   ├── sql             # SQL queries for sqlc
│   │   ├── user            # sqlc generated Go code
│   │   ├── sqlc.yaml       # sqlc configuration
│   │   └── ...
│   ├── router              # Router initialization
│   └── main.go             # Application entry point
├── go.mod                  # Dependency management
└── README.md               # Project documentation
```

---

## 🛠 Technology Stack

*   **Language**: Go 1.25+
*   **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
*   **Database**: PostgreSQL 16+ (with `pgvector`)
*   **Driver**: `pgx/v5`
*   **ORM/Codegen**: `sqlc` (Type-safe SQL to Go)
*   **Cache**: Redis
*   **Config**: Viper
*   **Logging**: Zap + Lumberjack
*   **Auth**: JWT

---

## 🚀 Development Workflow

### Prerequisites
*   Go 1.25+
*   Docker & Docker Compose
*   [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (CLI tool)

### Setup & Run

1.  **Start Infrastructure**:
    ```bash
    cd deploy
    docker-compose up -d
    ```
    This starts PostgreSQL (port 5432) and Redis (port 6379).

2.  **Configuration**:
    Copy `internal/etc/template.yaml` to `internal/etc/local.yaml` and adjust settings if necessary.

3.  **Run Application**:
    ```bash
    go mod tidy
    go run internal/main.go
    ```

### API Development Guide

To add a new API (e.g., `POST /api/user/profile`):

1.  **Define DTO (`internal/dto`)**:
    Create request and response structs with JSON tags.
    ```go
    type UpdateProfileReq struct {
        Nickname string `json:"nickname" binding:"required"`
    }
    ```

2.  **Define Route (`internal/router` & `internal/manager`)**:
    Register the route in `RouteManager`.

3.  **Implement Logic (`internal/logic`)**:
    Add method to `IUserLogic` interface and implement it.
    ```go
    func (l *UserLogic) UpdateProfile(ctx context.Context, req dto.UpdateProfileReq) error { ... }
    ```

4.  **Implement Handler (`internal/handler`)**:
    Use `middleware.GetBind` to get parsed request.
    ```go
    func (h *UserHandler) UpdateProfile(c *gin.Context) {
        req := middleware.GetBind[dto.UpdateProfileReq](c)
        err := h.userLogic.UpdateProfile(c.Request.Context(), req)
        response.Response(c, nil, err)
    }
    ```

### Database Operations (sqlc)

We use `sqlc` to generate Go code from SQL. **Do not write raw SQL in Go code.**

1.  **Modify Schema**: Edit `deploy/schema/user.sql` if table structure changes.
2.  **Write Query**: Create/Edit SQL files in `internal/repo/sql/`.
    ```sql
    -- name: GetUserByEmail :one
    SELECT * FROM "user" WHERE email = $1 LIMIT 1;
    ```
3.  **Generate Code**:
    ```bash
    # Execute in project root or internal/repo directory
    sqlc generate -f internal/repo/sqlc.yaml
    ```
4.  **Use in Repo**: Call the generated methods in `internal/repo/*.go`.

---

## 🚨 Error Handling

We follow a unified error handling strategy.

1.  **Repository Layer**:
    *   Catch `pgx` errors (e.g., `pgconn.PgError`, `pgx.ErrNoRows`).
    *   **Must** wrap or convert them into domain errors defined in `internal/repo/errors.go` or return specific defined errors (e.g., `ErrUserNotExist`).
    *   Log technical details (SQL errors) here using `global.Log.Error`.

2.  **Logic Layer**:
    *   Receive errors from Repo.
    *   Define business errors in `internal/logic/errors.go` (e.g., `ErrPasswordIncorrect`, `ErrEmailIsUsed`).
    *   Return these business errors to the Handler.

3.  **Handler Layer**:
    *   Pass the error directly to `response.Response(c, data, err)`.
    *   The `response` package will automatically map the error to a standardized JSON response with appropriate Code and Message.

**Example Response**:
```json
{
  "code": -1,
  "message": "User not found",
  "data": null
}
```

---

## 📝 Git Commit Convention

Please use **English** for all commit messages, following the [Conventional Commits](https://www.conventionalcommits.org/) specification.

**Format**: `<type>: <subject>`

**Types**:
*   `feat`: A new feature
*   `fix`: A bug fix
*   `docs`: Documentation only changes
*   `style`: Changes that do not affect the meaning of the code (white-space, formatting, etc)
*   `refactor`: A code change that neither fixes a bug nor adds a feature
*   `perf`: A code change that improves performance
*   `test`: Adding missing tests or correcting existing tests
*   `chore`: Changes to the build process or auxiliary tools and libraries

**Examples**:
*   `feat: add login with email functionality`
*   `fix: fix jwt token expiration time`
*   `refactor: migrate raw sql to sqlc`
*   `docs: update readme with project structure`

---

## 🤝 Git Collaboration Workflow

To ensure code stability and facilitate code review, **direct pushing to the `main` branch is prohibited**. Please follow the **Feature Branch Workflow**:

### 1. Create a Branch
Always create a new branch from `main` for your work.
*   **Branch Naming Convention**: `type/description`
    *   `feat/user-login`
    *   `fix/email-timeout`
    *   `refactor/project-structure`

```bash
# Update local main
git checkout main
git pull origin main

# Create and switch to a new branch
git checkout -b feat/add-user-profile
```

### 2. Commit Changes
Develop on your branch and commit changes using the [Commit Convention](#-git-commit-convention).

```bash
git add .
git commit -m "feat(user): implement update profile api"
```

### 3. Keep Up-to-Date
Before pushing, it is recommended to rebase your branch on the latest `main` to avoid conflicts and keep a clean history.

```bash
git fetch origin
git rebase origin/main
```
*If there are conflicts, resolve them, then `git rebase --continue`.*

### 4. Push & Pull Request
Push your branch to the remote repository.

```bash
git push origin feat/add-user-profile
# If you rebased, you might need to force push (be careful):
# git push -f origin feat/add-user-profile
```

Then, go to the repository page (GitHub/GitLab) and create a **Pull Request (PR)** or **Merge Request (MR)** to merge your branch into `main`.
*   Assign reviewers.
*   Ensure CI checks pass.
*   Wait for approval before merging.

---

## 🚢 Deployment

The project is containerized using Docker.

1.  **Build**:
    (Assuming a Dockerfile exists, otherwise standard Go build)
    ```bash
    go build -o nurture internal/main.go
    ```

2.  **Environment Variables**:
    Ensure `config/enter.go` can load configurations properly, typically via `local.yaml` or environment variables mapping.
