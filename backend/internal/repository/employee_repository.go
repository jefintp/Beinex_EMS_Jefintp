package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"employee-management/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL SQLSTATE for unique_violation.
const pgUniqueViolation = "23505"

// EmployeeRepository defines persistence operations for employees.
type EmployeeRepository interface {
	Create(ctx context.Context, e *model.Employee) error
	GetByID(ctx context.Context, id int64) (*model.Employee, error)
	List(ctx context.Context, f model.ListFilter) ([]model.Employee, error)
	Update(ctx context.Context, e *model.Employee) error
	Delete(ctx context.Context, id int64) error
	ListDepartments(ctx context.Context) ([]string, error)
}

type postgresEmployeeRepository struct {
	db *sql.DB
}

// NewEmployeeRepository returns a PostgreSQL-backed repository.
func NewEmployeeRepository(db *sql.DB) EmployeeRepository {
	return &postgresEmployeeRepository{db: db}
}

const selectColumns = `id, name, email, department, designation,
	to_char(joining_date, 'YYYY-MM-DD'), status, created_at, updated_at`

func scanEmployee(row interface{ Scan(dest ...any) error }) (*model.Employee, error) {
	var e model.Employee
	err := row.Scan(&e.ID, &e.Name, &e.Email, &e.Department, &e.Designation,
		&e.JoiningDate, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func mapError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return model.ErrDuplicateEmail
	}
	return err
}

func (r *postgresEmployeeRepository) Create(ctx context.Context, e *model.Employee) error {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO employees (name, email, department, designation, joining_date, status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+selectColumns,
		e.Name, e.Email, e.Department, e.Designation, e.JoiningDate, e.Status)
	created, err := scanEmployee(row)
	if err != nil {
		return mapError(err)
	}
	*e = *created
	return nil
}

func (r *postgresEmployeeRepository) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+selectColumns+` FROM employees WHERE id = $1`, id)
	e, err := scanEmployee(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	return e, err
}

func (r *postgresEmployeeRepository) List(ctx context.Context, f model.ListFilter) ([]model.Employee, error) {
	query := `SELECT ` + selectColumns + ` FROM employees`
	var (
		conds []string
		args  []any
	)
	if s := strings.TrimSpace(f.Search); s != "" {
		// name/email: partial, case-insensitive. id: exact match on the text form.
		args = append(args, "%"+s+"%", s)
		like := fmt.Sprintf("$%d", len(args)-1)
		exact := fmt.Sprintf("$%d", len(args))
		conds = append(conds, fmt.Sprintf(
			"(name ILIKE %s OR email ILIKE %s OR CAST(id AS TEXT) = %s)", like, like, exact))
	}
	if f.Department != "" {
		args = append(args, f.Department)
		conds = append(conds, fmt.Sprintf("department = $%d", len(args)))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := []model.Employee{}
	for rows.Next() {
		e, err := scanEmployee(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, *e)
	}
	return employees, rows.Err()
}

func (r *postgresEmployeeRepository) Update(ctx context.Context, e *model.Employee) error {
	row := r.db.QueryRowContext(ctx,
		`UPDATE employees
		 SET name = $1, email = $2, department = $3, designation = $4,
		     joining_date = $5, status = $6, updated_at = NOW()
		 WHERE id = $7
		 RETURNING `+selectColumns,
		e.Name, e.Email, e.Department, e.Designation, e.JoiningDate, e.Status, e.ID)
	updated, err := scanEmployee(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrNotFound
	}
	if err != nil {
		return mapError(err)
	}
	*e = *updated
	return nil
}

func (r *postgresEmployeeRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM employees WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *postgresEmployeeRepository) ListDepartments(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT department FROM employees ORDER BY department`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	departments := []string{}
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return departments, rows.Err()
}
