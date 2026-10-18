package service

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"employee-management/internal/model"
	"employee-management/internal/repository"
)

// EmployeeService contains business rules and validation.
type EmployeeService interface {
	Create(ctx context.Context, in model.EmployeeInput) (*model.Employee, error)
	List(ctx context.Context, f model.ListFilter) ([]model.Employee, error)
	Get(ctx context.Context, id int64) (*model.Employee, error)
	Update(ctx context.Context, id int64, in model.EmployeeInput) (*model.Employee, error)
	Delete(ctx context.Context, id int64) error
	Departments(ctx context.Context) ([]string, error)
}

type employeeService struct {
	repo repository.EmployeeRepository
}

// NewEmployeeService wires the service to a repository.
func NewEmployeeService(repo repository.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) Create(ctx context.Context, in model.EmployeeInput) (*model.Employee, error) {
	e, err := validate(in)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *employeeService) List(ctx context.Context, f model.ListFilter) ([]model.Employee, error) {
	f.Search = strings.TrimSpace(f.Search)
	f.Department = strings.TrimSpace(f.Department)
	f.Status = strings.ToLower(strings.TrimSpace(f.Status))
	return s.repo.List(ctx, f)
}

func (s *employeeService) Get(ctx context.Context, id int64) (*model.Employee, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *employeeService) Update(ctx context.Context, id int64, in model.EmployeeInput) (*model.Employee, error) {
	e, err := validate(in)
	if err != nil {
		return nil, err
	}
	e.ID = id
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *employeeService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *employeeService) Departments(ctx context.Context) ([]string, error) {
	return s.repo.ListDepartments(ctx)
}

// today returns the current date (server local time) as YYYY-MM-DD.
// It is a variable so tests can pin the date.
var today = func() string { return time.Now().Format("2006-01-02") }

// validate trims and checks input, returning a normalized Employee.
func validate(in model.EmployeeInput) (*model.Employee, error) {
	fields := map[string]string{}

	name := strings.TrimSpace(in.Name)
	switch {
	case name == "":
		fields["name"] = "name is required"
	case len(name) < 2 || len(name) > 100:
		fields["name"] = "name must be between 2 and 100 characters"
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	switch {
	case email == "":
		fields["email"] = "email is required"
	case len(email) > 255 || !validEmail(email):
		fields["email"] = "email is not a valid email address"
	}

	department := strings.TrimSpace(in.Department)
	switch {
	case department == "":
		fields["department"] = "department is required"
	case len(department) > 100:
		fields["department"] = "department must be at most 100 characters"
	}

	designation := strings.TrimSpace(in.Designation)
	switch {
	case designation == "":
		fields["designation"] = "designation is required"
	case len(designation) > 100:
		fields["designation"] = "designation must be at most 100 characters"
	}

	joiningDate := strings.TrimSpace(in.JoiningDate)
	if joiningDate == "" {
		fields["joining_date"] = "joining_date is required"
	} else if _, err := time.Parse("2006-01-02", joiningDate); err != nil {
		fields["joining_date"] = "joining_date must be a valid date in YYYY-MM-DD format"
	} else if joiningDate > today() {
		// Both are zero-padded YYYY-MM-DD, so string comparison is date comparison.
		fields["joining_date"] = "joining_date cannot be in the future"
	}

	status := strings.ToLower(strings.TrimSpace(in.Status))
	if status == "" {
		status = model.StatusActive
	}
	if status != model.StatusActive && status != model.StatusInactive {
		fields["status"] = "status must be either 'active' or 'inactive'"
	}

	if len(fields) > 0 {
		return nil, &model.ValidationError{Fields: fields}
	}

	return &model.Employee{
		Name:        name,
		Email:       email,
		Department:  department,
		Designation: designation,
		JoiningDate: joiningDate,
		Status:      status,
	}, nil
}

// validEmail accepts plain addresses like user@example.com only
// (rejects display-name forms such as "Bob <bob@example.com>").
func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	at := strings.LastIndex(email, "@")
	domain := email[at+1:]
	return strings.Contains(domain, ".") && !strings.HasSuffix(domain, ".")
}
