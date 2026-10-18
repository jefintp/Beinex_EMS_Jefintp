# Employee Management – Backend (Go + Gin + PostgreSQL)

## Structure

```
backend/
├── cmd/server/main.go            # entrypoint, wiring, graceful shutdown
├── internal/
│   ├── config/                   # env-based configuration
│   ├── model/                    # Employee, input/filter types, domain errors
│   ├── repository/               # PostgreSQL data access (database/sql + pgx)
│   ├── service/                  # validation + business rules
│   └── handler/                  # Gin HTTP handlers, error -> status mapping
├── migrations/001_create_employees.sql
├── docker-compose.yml            # optional PostgreSQL for local dev
└── .env.example
```

## Run

1. Start PostgreSQL (either option):
   - Docker: `docker compose up -d` (creates `employee_db` and applies the migration automatically)
   - Local PostgreSQL:
     ```bash
     createdb -U postgres employee_db
     psql -U postgres -d employee_db -f migrations/001_create_employees.sql
     ```
2. `cp .env.example .env` (Windows: `copy .env.example .env`) and adjust DB credentials.
3. `go mod tidy`
4. `go run ./cmd/server`

API base URL: `http://localhost:8080/api/v1`

Hosted databases usually require TLS: set `DB_SSLMODE=require` in `.env`.

## Endpoints

| Method | Path                     | Description                | Success |
|--------|--------------------------|----------------------------|---------|
| POST   | /employees               | Create employee            | 201     |
| GET    | /employees               | List (search/filter)       | 200     |
| GET    | /employees/:id           | Get by ID                  | 200     |
| PUT    | /employees/:id           | Update employee            | 200     |
| DELETE | /employees/:id           | Delete employee            | 204     |
| GET    | /departments             | Distinct departments       | 200     |
| GET    | /health (no prefix)      | Health check               | 200     |

List query params: `search` (name or email, case-insensitive), `department`, `status` (`active`/`inactive`).

### Employee payload

```json
{
  "name": "Asha Nair",
  "email": "asha.nair@example.com",
  "department": "Engineering",
  "designation": "Software Engineer",
  "joining_date": "2024-06-15",
  "status": "active"
}
```

`status` defaults to `active` when omitted. Emails are stored lowercase.

### Error responses

```json
{ "error": "validation failed", "fields": { "email": "email is not a valid email address" } }
```

| Status | When                                              |
|--------|---------------------------------------------------|
| 400    | Invalid JSON, invalid ID, or validation failure   |
| 404    | Employee not found                                |
| 409    | Email already exists (unique constraint)          |
| 500    | Unexpected server error                           |

## Try it

```bash
curl -X POST localhost:8080/api/v1/employees -H 'Content-Type: application/json' \
  -d '{"name":"Asha Nair","email":"asha.nair@example.com","department":"Engineering","designation":"Software Engineer","joining_date":"2024-06-15"}'

curl 'localhost:8080/api/v1/employees?search=asha&status=active'
```
