# Database migration operations

The API and cron process run the versioned migration runner. Each migration is
recorded in `schema_migrations`, and PostgreSQL uses a transaction-scoped
advisory lock so only one process changes the schema at a time. GORM
`AutoMigrate` is intentionally contained inside this reviewed, versioned
runner for the current release.

Before production rollout:

1. Take and verify a PostgreSQL backup.
2. Restore the backup into a staging database and run the API/cron startup
   checks against it.
3. Exercise the changed endpoints and inspect the migration table.
4. Roll forward with a corrected migration for failures; do not edit an
   already-applied migration or manually delete migration rows.

Schema rollback is a data-operation decision, not an automatic application
rollback. Keep the previous application available until the migration and
backward-compatibility checks pass. Data migrations must be additive and
restart-safe whenever possible.
