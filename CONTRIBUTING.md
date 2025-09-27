# Tenantly Project Conventions & Guidelines

This document outlines the coding, structure, and process conventions for the Tenantly project, covering Go (backend), .NET (notification service), and Angular (frontend) based on the latest best practices (as of 2025).

---

## General Principles
- Prioritize readability, maintainability, and security.
- Use clear, descriptive names for files, variables, and functions.
- Write modular, testable code with clear separation of concerns.
- Document public APIs, modules, and complex logic.
- Use version control (Git) with meaningful commit messages.

---

## Go Backend (Gin API)
- **Project Structure:**
  - Use a layered structure: `cmd/`, `internal/config/`, `internal/database/`, `internal/handlers/`, `internal/middleware/`, `internal/models/`, `internal/repositories/`, `internal/services/`, `internal/server/`.
  - Place main entry point in `cmd/server/main.go`.
  - Keep domain logic in `services/`, data access in `repositories/`, and HTTP logic in `handlers/`.
- **Code Style:**
  - Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
  - Use `gofmt` for formatting and `golint` for linting.
  - Use dependency injection via interfaces for services and repositories.
  - Prefer context-aware functions (`ctx context.Context`).
  - Use environment variables for configuration (see `internal/config`).
- **Testing:**
  - Place tests in the same package with `_test.go` suffix.
  - Use table-driven tests and mocks for external dependencies.
- **Security:**
  - Never hardcode secrets; use environment variables or secret managers.
  - Validate and sanitize all user input.
  - Use parameterized queries to prevent SQL injection.

---

## .NET Notification Service (C#)
- **Project Structure:**
  - Organize by domain: `Services/`, `Models/`, `Configuration/`.
  - Use dependency injection for all services.
- **Code Style:**
  - Follow [Microsoft C# Coding Conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions).
  - Use `async`/`await` for all I/O operations.
  - Use nullable reference types and guard against nulls.
  - Use `ILogger<T>` for logging.
- **Testing:**
  - Use xUnit or NUnit for unit tests.
  - Place tests in a separate test project.
- **Security:**
  - Store secrets in `appsettings.json` (never in code) and use user secrets or environment variables for local dev.
  - Validate all external input.

---

## Angular Frontend (Angular 20+)
- **Project Structure:**
  - Use feature-based folders under `src/app/features/`.
  - Use standalone components (`[name].ts`, `[name].html`, `[name].scss`).
  - Place shared services in `src/app/core/services/`.
- **Code Style:**
  - Follow [Angular Style Guide](https://angular.io/guide/styleguide).
  - Use TypeScript strict mode.
  - Prefer signals and RxJS for state management.
  - Use Angular Material for UI components and theming.
  - Use external templates and styles (no inline HTML/CSS).
  - Remove `.component` suffix from filenames (e.g., `login.ts`).
- **Testing:**
  - Use Jasmine/Karma for unit tests.
  - Place tests alongside components with `.spec.ts` suffix.
- **Security:**
  - Never commit secrets or API keys.
  - Use Angular's built-in sanitization for user input.
  - Use environment files for configuration.

---


## Git & Collaboration
- Use feature branches for all new work.
- Open pull requests for all changes; require code review before merging.
- Run all tests and linters before submitting code.
- Do not commit secrets, credentials, or sensitive data.

### Branch Naming Strategy
- Use short, descriptive, and lowercase branch names separated by hyphens.
- Prefix branches by type:
  - `feat/` for new features (e.g., `feat/user-authentication`)
  - `fix/` for bug fixes (e.g., `fix/payment-validation`)
  - `chore/` for maintenance (e.g., `chore/update-deps`)
  - `docs/` for documentation (e.g., `docs/api-docs`)
  - `test/` for testing (e.g., `test/unit-payment-service`)
- Use issue or ticket numbers if available (e.g., `feat/123-user-login`).

### Commit Message Guidelines
- Use the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format:
  - `type(scope): short description`
  - Example: `feat(auth): add JWT authentication middleware`
- Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `style`, `perf`.
- Use the imperative mood: "add", "fix", "update", not "added" or "fixed".
- Keep the first line under 72 characters.
- Add a blank line before more detailed body text if needed.
- Reference issues or PRs when relevant (e.g., `fix: correct payment logic (#42)`).

---

## Naming Conventions
- **Go:** `CamelCase` for types, `snake_case` for file names, short receiver names.
- **C#:** `PascalCase` for types/methods, `camelCase` for parameters/fields.
- **Angular:** `kebab-case` for files/folders, `PascalCase` for classes.

---

## Security & Secrets
- Use [Gitleaks](https://github.com/gitleaks/gitleaks) or similar tools to scan for secrets before pushing.
- Add `.env`, `appsettings.*.json`, and other secret files to `.gitignore`.

---

## Documentation
- Update the root `README.md` and feature/module-level `README.md` as needed.
- Document all public APIs and endpoints.

---

## Further Reading
- [Go Best Practices](https://github.com/golang/go/wiki/BestPractices)
- [Microsoft .NET Docs](https://learn.microsoft.com/en-us/dotnet/)
- [Angular Docs](https://angular.io/docs)

---

_Keep code clean, secure, and maintainable!_
