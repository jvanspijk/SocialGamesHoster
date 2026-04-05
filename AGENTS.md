# AGENTS.md

Guidance for agentic coding assistants working in `SocialGamesHoster`.

## Scope

- Applies to the whole repository rooted at `C:\Users\Jason\Documents\GitHub\SocialGamesHoster`.
- Tech stack: .NET 10 backend (`API*` projects) + SvelteKit frontend (`Web`).
- Architecture emphasis: vertical slices by feature, minimal API endpoints, EF Core + PostgreSQL.

## Rule Sources To Honor

- Copilot rules present: `/.github/copilot-instructions.md`.
- Additional project instructions present: `/copilot-instructions.md`.
- Frontend visual language source: `/Web/DESIGN.MD`.
- Cursor rules: no `.cursorrules` and no `.cursor/rules/` directory were found at time of writing.

## Repository Layout

- `API/`: ASP.NET Core minimal API host, endpoint definitions, filters, auth, SignalR hubs.
- `API.DataAccess/`: EF Core `APIDatabaseContext`, generic `Repository<T>`, migrations, seeders.
- `API.Domain/`: entities, shared `Result` types, domain validation/error primitives.
- `API.LogViewer/`: auxiliary web project for viewing debug logs.
- `Web/`: SvelteKit app (TypeScript, ESLint, Prettier, Tailwind CSS).
- Root scripts: `install.bat` and `run.bat` for local dependency setup and startup.

## Setup Commands

- Full local dependency bootstrap (Windows):
  - `install.bat`
- Start API + web together (Windows):
  - `run.bat`
- Manual frontend dependency install:
  - `npm install` (run in `Web/`)

## Build, Lint, Check, Run Commands

### Backend (.NET)

- Restore solution:
  - `dotnet restore SocialGamesHoster.sln`
- Build all projects:
  - `dotnet build SocialGamesHoster.sln`
- Build API only:
  - `dotnet build API/API.csproj`
- Run API:
  - `dotnet run --project API/API.csproj`
- Optional format pass:
  - `dotnet format SocialGamesHoster.sln`

### Frontend (SvelteKit)

- Run dev server:
  - `npm run dev` (in `Web/`)
- Production build:
  - `npm run build` (in `Web/`)
- Type and Svelte checks:
  - `npm run check` (in `Web/`)
- Lint:
  - `npm run lint` (in `Web/`)
- Format:
  - `npm run format` (in `Web/`)
- SDK generation:
  - `npm run client:generate` (in `Web/`)

## Test Commands (Current State + Single-Test Guidance)

Current state:

- No backend test project (`*Test*.csproj` / `*Tests*.csproj`) detected.
- No frontend test runner config (Vitest/Jest/Playwright) detected.

If/when backend tests are added:

- Run all tests:
  - `dotnet test SocialGamesHoster.sln`
- Run one test by fully qualified name:
  - `dotnet test --filter "FullyQualifiedName~Namespace.ClassName.MethodName"`
- Run by display name substring:
  - `dotnet test --filter "DisplayName~partial name"`

If/when frontend tests are added (examples):

- Vitest single test file:
  - `npx vitest run src/path/to/file.test.ts`
- Vitest by test name:
  - `npx vitest run -t "test name"`

## Coding Style Guidelines

### Cross-Cutting

- Prefer small, composable functions over inheritance-heavy designs.
- Use guard clauses to keep control flow flat and reduce nested conditionals.
- Keep feature changes localized to their vertical slice when possible.
- Avoid introducing heavy dependencies unless there is a clear need.
- Do not use boxing; project rule explicitly disallows boxing.

### C# / Backend

- Use modern C#/.NET style used in repo:
  - file-scoped namespaces,
  - primary constructors where practical,
  - target-typed `new()` where clear.
- Type usage:
  - avoid `var` for primitive types when explicit type improves clarity,
  - prefer explicit DTO/request/response records for endpoint contracts,
  - prefer immutable patterns (`readonly` fields, `init` where appropriate).
- Naming conventions:
  - private fields start with `_`,
  - async methods must end with `Async`,
  - endpoint handlers follow `HandleAsync`/`Handle` naming seen in `Features/*/Endpoints`.
- Imports/usings:
  - keep usings at file top,
  - remove unused usings.
- Data access and EF Core:
  - use repository abstractions already present (`IRepository<T>`),
  - use `AsNoTracking()` for read-only queries,
  - use projection via `IProjectable` DTOs for query shaping,
  - avoid bypassing repository pattern without clear reason.
- API behavior:
  - never return HTTP 200 for failed operations,
  - return typed results and proper status codes,
  - keep endpoint metadata (`Produces`, tags, names) consistent with existing style.

### Error Handling (Backend)

- Prefer railway-oriented flow using `Result` / `Result<T>` patterns.
- Do not use exceptions for normal control flow.
- Return explicit, actionable error messages for failures.
- Throw only for exceptional, unrecoverable conditions.

### TypeScript / Svelte Frontend

- Use TypeScript strict mode expectations (`Web/tsconfig.json` has `strict: true`).
- Keep API interactions typed (`ApiResponse<T>`, endpoint wrappers).
- Prefer explicit domain types over `any`.
- Follow existing formatting tools instead of manual style drift:
  - Prettier config: tabs, single quotes, no trailing commas, print width 100.
  - ESLint config: JS + TS + Svelte recommended configs with Prettier integration.
- For chat admin viewers, use `readerId = 0` to indicate no real player identity.

## Architecture-Specific Practices

- Keep API features organized by slice:
  - `Features/<FeatureName>/Endpoints/*`
  - optional `Common/DTO.cs`, `Hubs/*`, supporting services.
- Use shared endpoint registration patterns in `API/Endpoints.cs`.
- Reuse filters for cross-cutting concerns (error formatting, cache invalidation, debug behavior).
- For generated client code under `Web/src/lib/client`, prefer regeneration over manual drift when endpoint contracts change.

## Agent Workflow Checklist

- Before editing, inspect neighboring files in the same feature slice for local conventions.
- After backend changes, run `dotnet build SocialGamesHoster.sln`.
- After frontend changes, run `npm run check` and `npm run lint` (and `npm run build` for production-impacting changes).
- Keep docs in sync when changing commands, architecture, or rules.

## Operational Notes

- During development/debug, project guidance prefers running without Docker/nginx.
- Build-time OpenAPI generation concerns exist; avoid forcing PostgreSQL-dependent migration work in doc-generation contexts.
- Prefer PostgreSQL over in-memory database for normal local dev unless a specific build-time/doc-generation path requires otherwise.
