# Database Migrations

This directory contains SQL migration files to set up and update the database schema.

## Migration Files

- `001_initial_schema.sql` - Creates the initial database schema with decks and games tables

## Running Migrations

### Using Make (Recommended)
```bash
# Run all pending migrations
make migrate

# Show migration help
make migrate-help
```

### Using Go directly
```bash
# Run migrations
go run cmd/migrate/main.go

# Show help
go run cmd/migrate/main.go -h
```

### Using compiled binary
```bash
# Build the migration tool
go build -o bin/migrate cmd/migrate/main.go

# Run migrations
./bin/migrate
```

## Environment Configuration

The migration tool uses the same environment variables as the main application:

- `DATABASE_URL` - PostgreSQL connection string
- If no environment variable is set, it will try to load from `.env` file

Example `.env`:
```
DATABASE_URL=postgres://quards:quards@localhost/quards_db?sslmode=disable
```

## Production Deployment

For production, set the DATABASE_URL environment variable and run:

```bash
# Set production database URL
export DATABASE_URL="postgres://user:password@host:5432/database?sslmode=require"

# Run migrations
go run cmd/migrate/main.go
```

## Migration Tracking

The system automatically creates a `schema_migrations` table to track which migrations have been applied. Each migration is run only once.

## Creating New Migrations

1. Create a new file in this directory with format: `XXX_description.sql`
   - Use sequential numbers (002, 003, etc.)
   - Use descriptive names

2. Include proper SQL statements with error handling:
   ```sql
   CREATE TABLE IF NOT EXISTS ...
   CREATE INDEX IF NOT EXISTS ...
   ```

3. The migration runner will automatically detect and apply new migrations.

## Safety Features

- Migrations run in transactions
- Each migration is applied only once
- Migrations are applied in alphabetical order
- Failed migrations will roll back automatically