# Repository Layer (internal/repository)

The **Repository** layer handles direct data access and persistence (PostgreSQL, Redis, etc.).

## Go Best Practices:
1. **Return Concrete Structs**: Repositories implement methods that satisfy the interfaces required by services without needing to declare them explicitly.
2. **Modern Tooling Recommendations**:
   - Driver: `github.com/jackc/pgx/v5` (high-performance native PostgreSQL driver with connection pooling).
   - Typed SQL: `sqlc` (generates type-safe Go code from pure `.sql` query files).
3. **Always Accept `context.Context`**: Every database operation must accept `ctx` as its first parameter to honor query timeouts and HTTP client aborts:
   ```go
   func (r *PostgresAppointmentRepo) Create(ctx context.Context, appt domain.Appointment) error
   ```
