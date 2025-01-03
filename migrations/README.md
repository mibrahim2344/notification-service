# Database Migrations

This directory contains database migrations for the notification service.

## Migration Files

Migration files follow the naming convention:
```
{timestamp}_{description}.{up|down}.sql
```

For example:
- `20250103180000_create_templates_table.up.sql`
- `20250103180000_create_templates_table.down.sql`

## Commands

### Create a new migration
```bash
make migrate-create
```
This will prompt for a migration name and create two files:
- `{timestamp}_{name}.up.sql`: Contains the forward migration
- `{timestamp}_{name}.down.sql`: Contains the rollback migration

### Run migrations
```bash
# Run all pending migrations
make migrate-up

# Rollback all migrations
make migrate-down

# Run specific number of migrations
make migrate-steps steps=1

# Force set migration version
make migrate-force version=1

# Check current migration version
make migrate-version
```

## Environment Variables

The migration tool uses the following environment variables:
- `DB_HOST`: PostgreSQL host (default: "localhost")
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: PostgreSQL user (default: "postgres")
- `DB_PASSWORD`: PostgreSQL password (default: "postgres")
- `DB_NAME`: Database name (default: "notification_service")
- `DB_SSLMODE`: SSL mode (default: "disable")
