# Nurture Agent Instructions

This repository has a project-specific backend guide. Before changing Go backend code, API routes, database/sqlc files, tests, or architecture, read:

- `backend-develop-guide/SKILL.md`

Then read the focused reference that matches the task:

- Module boundaries or detachable domains: `backend-develop-guide/reference/module-boundary.md`
- Database schema, SQL, or sqlc: `backend-develop-guide/reference/database-guide.md`
- Error flow or error mapping: `backend-develop-guide/reference/error-handling.md`
- Naming, package, or file placement: `backend-develop-guide/reference/naming-conventions.md`
- Code patterns, DTO, handlers, middleware, or constructors: `backend-develop-guide/reference/code-patterns.md`

Repository rules:

- Keep `internal/pkg` business-free. Put business code in shared layers or an `internal/<domain>` module.
- Keep package-owned tests beside the package. Use `internal/test` only for cross-business integration tests.
- Use English commit messages in `action: content` format.
- Do not use unrelated external skills or conventions when this guide covers the task.
- Verify changes before reporting completion.
