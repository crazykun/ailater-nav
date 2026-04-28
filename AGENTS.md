# AGENTS.md

## Commands

```bash
# First run: create config (config.yaml is gitignored)
cp config.demo.yaml config.yaml

# Create database (requires MySQL running)
go run scripts/create_db/main.go

# Migrate data from JSON to MySQL
go run scripts/migrate/main.go

# Dev server (port 8080, configurable in config.yaml)
go run main.go

# Build binary
go build -o ai-later-nav

# Restart (stop old process, rebuild, start)
./scripts/restart.sh

# Stop running process
./scripts/stop.sh

# CI pipeline runs these in order:
go build -v ./...
go test -v ./...
```

No linter, formatter, or typecheck tooling configured.

## Architecture

Single-binary Gin web app with MySQL backend. Module name: `ai-later-nav` (Go 1.24+).

**Data flow**: Sites stored in MySQL (`ai_later` database). Admin CRUD: service layer → repository layer → database. Migrations run automatically on startup.

**Project structure**:
```
main.go                        Entry point, routes, go:embed templates/static
internal/
  config/config.go             Config struct, YAML loading, env overrides
  database/
    db.go                      MySQL connection management
    migrate.go                 SQL migration runner (runs on startup)
    migrations/                SQL schema files
    repository/
      site_repo.go             Site CRUD operations
      user_repo.go             User and favorites operations
  handlers/
    page_handlers.go           SSR page rendering
    api_handlers.go            HTMX API endpoints
    admin_handlers.go          Admin CRUD handlers
  models/
    site.go, user.go           Data models
  services/
    site_service.go            Site business logic
    user_service.go            User auth and favorites
  middleware/
    globalmiddleware.go        JWT auth, global context injection
  web/
    templates.go               Template registration (BuildPageTemplates, BuildSharedTemplates)
scripts/
  create_db/main.go            Create MySQL database
  migrate/main.go              JSON → MySQL migration tool
  restart.sh                   Stop, rebuild, start
  stop.sh                      Stop running process
templates/                     Go HTML templates (go:embed)
  *.html                       Page templates
  partials/                    HTMX partial templates
  admin/                       Admin templates
static/                        CSS, JS, images (go:embed)
test/
  templ_test.go                Template rendering tests
```

## Key gotchas

- **Template registration is two-step**: Templates are embedded via `//go:embed templates/*` in `main.go`, but new templates must also be registered in `internal/web/templates.go` (`BuildPageTemplates` for public pages, `BuildSharedTemplates` for admin pages). Forgetting this causes runtime nil-pointer panics.
- **Migrations run on startup**: `database.RunMigrations()` in `main.go` auto-applies SQL files from `internal/database/migrations/`. No separate migrate command needed for normal dev.
- **MySQL must be running**: App requires MySQL. Connection config in `config.yaml` under `mysql:` section.
- **config.yaml is gitignored**: Never commit it. `config.demo.yaml` is the committed template.
- **Site lookup uses database ID**: Admin routes use numeric `:id` parameter (database primary key), not site name.
- **Working directory matters**: go:embed paths are relative to source, but `config.yaml` is read from the working directory.
- **First-run setup**: If no admin user exists, the app redirects to `/setup` to create one.

## Database

MySQL 8.0+ with tables: `sites`, `tags`, `site_tags`, `users`, `favorites`, `visits`, `settings`, `schema_migrations`.

Env vars for MySQL config: `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`.

## Authentication

JWT with httpOnly cookies. Tokens expire after 7 days (configurable in `jwt.expire_days`). Admin access requires `role = 'admin'` in `users` table.
