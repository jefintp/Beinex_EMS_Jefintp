# Employee Management – Angular frontend

Angular 20 (standalone components, signals, reactive forms). Talks to the Go REST API.

## Run

```bash
npm install
npm start            # http://localhost:4200
```

`npm start` uses `proxy.conf.json`, which forwards `/api/*` to `http://localhost:8080` (the Go backend), so no CORS setup is needed in development. Change the target there if your backend runs elsewhere.

Production build: `npm run build` (output in `dist/ems-frontend/browser`).

## Features

| Requirement | Where |
| --- | --- |
| Employee listing | `employees/employee-list` |
| Add / Edit form (same component, mode chosen by route) | `employees/employee-form` |
| Delete confirmation | `shared/confirm-dialog`, used by the list |
| Form validation (required, length, email format, whitespace-only) | `employee-form.component.ts`, `shared/validators.ts` |
| Search by name/email, filter by department/status | list filters (300 ms debounce, in-flight requests cancelled) |
| API errors | `core/api-error.ts` → banner, per-field messages, toast |
| Loading states | list overlay, form "loading employee" card, button spinners while saving/deleting |

## API contract the frontend expects

Base path `/api/v1/employees`, JSON with snake_case fields.

| Action | Request | Success |
| --- | --- | --- |
| List | `GET /api/v1/employees?search=&department=&status=` | `200` `{ "data": [Employee], "total": n }` (`search` matches name **or** email, case-insensitive; empty filters are omitted) |
| Get | `GET /api/v1/employees/{id}` | `200` `Employee` |
| Create | `POST /api/v1/employees` | `201` `Employee` |
| Update | `PUT /api/v1/employees/{id}` | `200` `Employee` |
| Delete | `DELETE /api/v1/employees/{id}` | `204` |

```json
{
  "id": 1,
  "name": "Asha Nair",
  "email": "asha@company.com",
  "department": "Engineering",
  "designation": "Software Engineer",
  "joining_date": "2025-06-02",
  "status": "active"
}
```

`status` is `"active"` or `"inactive"`. `joining_date` is `YYYY-MM-DD` (a full ISO timestamp is also accepted).

### Error format

```json
{ "error": "validation failed", "fields": { "email": "email already exists" } }
```

- `400` / `422` – validation errors; `fields` entries show under the matching input.
- `409` – duplicate email; shown under the Email field even if `fields` is omitted.
- `404` – employee not found; `5xx` and network failures show a generic message with a Retry button on the list.

## Structure

```
src/app/
  core/        api-error.ts, notification.service.ts
  employees/   employee.model.ts, employee.service.ts, employee-list/, employee-form/
  shared/      confirm-dialog/, validators.ts
```
