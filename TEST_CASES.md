# Employee Management System – Test Cases

How to run each layer:

| Layer | Command (run in `backend/`) | Needs |
|---|---|---|
| Automated unit + HTTP tests | `go test ./...` | Go only (no database) |
| Live API test | `node tests/api-smoke.mjs` | Backend + PostgreSQL running, Node 18+ |
| UI checklist | Manual, below | Backend + frontend running |

---

## 1. Automated backend tests (`go test ./...`)

Expected: every package prints `ok`. Add `-v` to see each case.

| Area | What is checked |
|---|---|
| Validation | Required fields, trimming, email lowercased, status defaults to `active`, status case-insensitive |
| Email format | Rejects `abc`, `abc@`, `@example.com`, `a@b`, `a@b.`, spaces, `Bob <bob@example.com>`, `a@@example.com` |
| Dates | Rejects `2024-13-45`, `15-06-2024`, `2024/06/15`, `2024-02-30` |
| Duplicate email | Create with an existing email (any letter case) is rejected; updating to another employee's email is rejected; keeping your own email is fine |
| CRUD | Create, get, update, delete; unknown id returns not found; deleting twice returns not found |
| Search and filters | Name/email search (case-insensitive), department, status, combined, no match |
| HTTP status codes | 201 create, 200 get/list/update, 204 delete, 400 bad JSON / bad id / validation, 404 not found, 409 duplicate, 500 without leaking internal error text |
| Response shape | List returns `{"data":[...],"total":n}`; empty list is `[]`, not `null` |

## 2. Live API test (`node tests/api-smoke.mjs`)

Runs ~35 checks against the real database and cleans up after itself. Expected last line: `N passed, 0 failed`.
Covers everything in section 1 end-to-end, plus: data persisting after update, the departments endpoint, and a deleted email being reusable.

If it prints "Cannot reach the backend", start the backend first.

## 3. Manual UI checklist (browser at http://localhost:4200)

Tick each box. "Expected" is what you should see.

### Listing
| # | Steps | Expected | OK |
|---|---|---|---|
| L1 | Open the app with an empty database | Empty-state message, no errors | ☐ |
| L2 | Open the app with employees | Table shows name, email, department, designation, joining date, status | ☐ |
| L3 | Refresh the page | A loading indicator shows briefly, then the list | ☐ |

### Add employee
| # | Steps | Expected | OK |
|---|---|---|---|
| A1 | Click Add employee, fill every field correctly, save | Returns to the list; new employee is shown | ☐ |
| A2 | Submit the empty form | Required-field messages; nothing is sent | ☐ |
| A3 | Enter email `abc` | "valid email" message under the field | ☐ |
| A4 | Enter a name of 1 character | Length message | ☐ |
| A5 | Add an employee with an email already in use | Error under the Email field (duplicate); you stay on the form with your data intact | ☐ |
| A6 | Click Save and watch the button | Button is disabled and shows a saving state while the request runs | ☐ |
| A8 | Open the Joining date picker | Dates after today are greyed out; typing a future date shows "cannot be in the future" | ☐ |
| A7 | Click Cancel | Returns to the list; nothing saved | ☐ |

### Edit employee
| # | Steps | Expected | OK |
|---|---|---|---|
| E1 | Click Edit on a row | Form opens pre-filled with that employee | ☐ |
| E2 | Change designation and save | List shows the new value | ☐ |
| E3 | Change status to Inactive and save | Badge changes to inactive | ☐ |
| E4 | Change email to another employee's email | Duplicate error under Email | ☐ |
| E5 | Save without changing anything | Saves with no error (own email is not a duplicate) | ☐ |
| E6 | Open `/employees/99999/edit` directly | Friendly "not found" message with a way back | ☐ |

### Delete
| # | Steps | Expected | OK |
|---|---|---|---|
| D1 | Click Delete on a row | Confirmation dialog names the employee | ☐ |
| D2 | Click Cancel | Dialog closes; employee still listed | ☐ |
| D3 | Click Delete, then confirm | Employee disappears from the list | ☐ |
| D4 | Delete the last remaining employee | Empty-state message appears | ☐ |

### Search and filters
| # | Steps | Expected | OK |
|---|---|---|---|
| S1 | Type part of a name | List narrows after a short pause | ☐ |
| S2 | Type part of an email | List narrows | ☐ |
| S3 | Type in UPPER case | Same results (case-insensitive) | ☐ |
| S4 | Search for text that matches nobody | Empty state saying nothing matched; a way to clear filters | ☐ |
| S5 | Pick a department | Only that department is shown | ☐ |
| S6 | Pick a status | Only that status is shown | ☐ |
| S7 | Combine search + department + status | Results satisfy all three | ☐ |
| S8 | Clear filters | Full list returns | ☐ |

### Error handling
| # | Steps | Expected | OK |
|---|---|---|---|
| X1 | Stop the backend (close its window), refresh the list | Clear "cannot reach the server" message with a Retry button | ☐ |
| X2 | Start the backend again, click Retry | List loads | ☐ |
| X3 | With the backend stopped, try to save a new employee | Error message shown; form data kept | ☐ |

### Known limits (not bugs)
- The department dropdown is a fixed list in the frontend. Departments created only through the API still appear in the table but not in the dropdown.
- Searching matches name and email only (not department or designation).
