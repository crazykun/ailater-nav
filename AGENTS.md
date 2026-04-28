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

# CI pipeline runs these in order:
go build -v ./...
go test -v ./...
```

No tests exist yet. `go test` passes vacuously. No linter, formatter, or typecheck tooling configured.

## Architecture

Single-binary Gin web app with MySQL backend. Module name: `ai-later-nav` (Go 1.24+).

**Data flow**: Sites are stored in MySQL (`ai_later` database). Admin CRUD operations go through the service layer → repository layer → database. Data is no longer loaded from JSON files at runtime.

**Database**: MySQL 5.6+ with connection pooling. Schema managed via SQL migrations in `internal/database/migrations/`. Connection config in `config.yaml` under `mysql:` section.

**Project structure** (follows Go standard layout):
```
main.go                        Entry point, routes, go:embed templates/static
internal/                      Private packages (not importable externally)
  config/config.go             Config struct, YAML loading, env overrides
  database/
    db.go                      MySQL connection management
    migrate.go                 SQL migration runner
    migrations/                SQL schema files
    repository/
      site_repo.go             Site CRUD operations
      user_repo.go             User and favorites operations
  handlers/
    page_handlers.go           SSR page rendering
    api_handlers.go            HTMX API endpoints
    admin_handlers.go          Admin CRUD handlers
    helpers.go                 JWT token generation helper
  models/
    site.go                    Site and SiteDisplay structs
    user.go                    User and UserClaims structs
  services/
    site_service.go            Site business logic
    user_service.go            User auth and favorites
  middleware/
    globalmiddleware.go        JWT auth, global context injection
  utils/
    color_helper.go            Deterministic color+initials from site name
scripts/
  create_db/main.go            Create MySQL database
  migrate/main.go              JSON → MySQL migration tool
templates/                     Go HTML templates (go:embed)
  *.html                       Page templates
  partials/                    HTMX partial templates
  admin/                       Admin templates
static/                        CSS, JS, images (go:embed)
data/
  ai.json                      Legacy data (used only for migration)
```

## Key gotchas

- **Template loading uses go:embed**: Templates are embedded in the binary via `//go:embed templates/*`. No need to manually register new templates, but they must be in the `templates/` directory.
- **MySQL must be running**: The app requires a MySQL instance. Connection config in `config.yaml` under `mysql:` section.
- **Run migrations before first use**: `go run scripts/create_db/main.go` creates the database, then `go run scripts/migrate/main.go` creates tables and imports data.
- **Site lookup uses database ID**: Admin routes use numeric `:id` parameter (database primary key), not site name.
- **config.yaml is gitignored**: Never commit it. `config.demo.yaml` is the committed template.
- **Working directory matters**: Template and static paths are embedded via go:embed, but MySQL config is read from config.yaml in working directory.

## Database

**MySQL 5.6+** with the following tables:
- `sites` - Site information with soft delete support
- `tags` - Tag names
- `site_tags` - Many-to-many relationship between sites and tags
- `users` - User accounts with bcrypt password hashing
- `favorites` - User's favorite sites
- `visits` - Visit tracking
- `schema_migrations` - Migration version tracking

**Env vars for MySQL config**:
- `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`

## Authentication

JWT-based authentication with httpOnly cookies. Tokens expire after 7 days (configurable in `jwt.expire_days`).

**Admin access**: Set user role to `admin` in the `users` table. First registered user should be manually promoted.

## HTMX Integration

The frontend uses HTMX for dynamic interactions:
- `hx-get="/api/search"` - Live search with debounce
- `hx-get="/api/sites/:id"` - Site detail modal
- `hx-post="/api/auth/login"` - Login form submission
- `hx-post="/api/auth/register"` - Register form submission
- `hx-post="/api/favorites/:id"` - Toggle favorite (requires auth)

HTMX partials are in `templates/partials/`.
