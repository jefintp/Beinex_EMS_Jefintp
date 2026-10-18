package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"

	"employee-management/internal/model"
)

// ---- in-memory fake repository ----------------------------------------

type fakeRepo struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]model.Employee
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{nextID: 1, items: map[int64]model.Employee{}}
}

func (r *fakeRepo) emailTaken(email string, exceptID int64) bool {
	for id, e := range r.items {
		if id != exceptID && e.Email == email {
			return true
		}
	}
	return false
}

func (r *fakeRepo) Create(_ context.Context, e *model.Employee) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.emailTaken(e.Email, 0) {
		return model.ErrDuplicateEmail
	}
	e.ID = r.nextID
	r.nextID++
	r.items[e.ID] = *e
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id int64) (*model.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return &e, nil
}

func (r *fakeRepo) List(_ context.Context, f model.ListFilter) ([]model.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []model.Employee{}
	for _, e := range r.items {
		if f.Search != "" {
			s := strings.ToLower(f.Search)
			byID := strconv.FormatInt(e.ID, 10) == f.Search
			if !byID && !strings.Contains(strings.ToLower(e.Name), s) && !strings.Contains(strings.ToLower(e.Email), s) {
				continue
			}
		}
		if f.Department != "" && e.Department != f.Department {
			continue
		}
		if f.Status != "" && e.Status != f.Status {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *fakeRepo) Update(_ context.Context, e *model.Employee) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[e.ID]; !ok {
		return model.ErrNotFound
	}
	if r.emailTaken(e.Email, e.ID) {
		return model.ErrDuplicateEmail
	}
	r.items[e.ID] = *e
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return model.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *fakeRepo) ListDepartments(_ context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := map[string]bool{}
	out := []string{}
	for _, e := range r.items {
		if !seen[e.Department] {
			seen[e.Department] = true
			out = append(out, e.Department)
		}
	}
	return out, nil
}

// ---- helpers ------------------------------------------------------------

func validInput() model.EmployeeInput {
	return model.EmployeeInput{
		Name:        "Asha Nair",
		Email:       "asha.nair@example.com",
		Department:  "Engineering",
		Designation: "Software Engineer",
		JoiningDate: "2024-06-15",
		Status:      "active",
	}
}

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	var ve *model.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}
	return ve.Fields
}

// ---- validation ----------------------------------------------------------

func TestValidate_NormalizesInput(t *testing.T) {
	in := validInput()
	in.Name = "  Asha Nair  "
	in.Email = "  Asha.Nair@Example.COM "
	in.Status = ""

	e, err := validate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Name != "Asha Nair" {
		t.Errorf("name not trimmed: %q", e.Name)
	}
	if e.Email != "asha.nair@example.com" {
		t.Errorf("email not lowercased/trimmed: %q", e.Email)
	}
	if e.Status != model.StatusActive {
		t.Errorf("status should default to active, got %q", e.Status)
	}
}

func TestValidate_RequiredFields(t *testing.T) {
	for _, field := range []string{"name", "email", "department", "designation", "joining_date"} {
		t.Run(field, func(t *testing.T) {
			in := validInput()
			switch field {
			case "name":
				in.Name = "   "
			case "email":
				in.Email = ""
			case "department":
				in.Department = ""
			case "designation":
				in.Designation = ""
			case "joining_date":
				in.JoiningDate = ""
			}
			_, err := validate(in)
			if _, ok := fieldsOf(t, err)[field]; !ok {
				t.Errorf("expected an error for %q", field)
			}
		})
	}
}

func TestValidate_EmailFormat(t *testing.T) {
	bad := []string{
		"abc",
		"abc@",
		"@example.com",
		"a@b",
		"a@b.",
		"a b@example.com",
		"Bob <bob@example.com>",
		"a@@example.com",
	}
	for _, email := range bad {
		t.Run("bad/"+email, func(t *testing.T) {
			in := validInput()
			in.Email = email
			_, err := validate(in)
			if _, ok := fieldsOf(t, err)["email"]; !ok {
				t.Errorf("%q should be rejected", email)
			}
		})
	}

	good := []string{"a@b.co", "first.last@example.com", "user+tag@sub.example.org"}
	for _, email := range good {
		t.Run("good/"+email, func(t *testing.T) {
			in := validInput()
			in.Email = email
			if _, err := validate(in); err != nil {
				t.Errorf("%q should be accepted, got %v", email, err)
			}
		})
	}
}

func TestValidate_JoiningDate(t *testing.T) {
	for _, d := range []string{"2024-13-45", "15-06-2024", "2024/06/15", "yesterday", "2024-02-30"} {
		t.Run("bad/"+d, func(t *testing.T) {
			in := validInput()
			in.JoiningDate = d
			_, err := validate(in)
			if _, ok := fieldsOf(t, err)["joining_date"]; !ok {
				t.Errorf("%q should be rejected", d)
			}
		})
	}
}

func TestValidate_FutureJoiningDate(t *testing.T) {
	original := today
	today = func() string { return "2026-10-18" }
	defer func() { today = original }()

	cases := []struct {
		date      string
		wantError bool
	}{
		{"2026-10-19", true},  // tomorrow
		{"2030-01-01", true},  // far future
		{"2026-10-18", false}, // today is allowed
		{"2026-10-17", false}, // yesterday
		{"2020-02-29", false}, // valid leap day in the past
	}
	for _, c := range cases {
		t.Run(c.date, func(t *testing.T) {
			in := validInput()
			in.JoiningDate = c.date
			_, err := validate(in)
			if c.wantError {
				if _, ok := fieldsOf(t, err)["joining_date"]; !ok {
					t.Errorf("%s should be rejected as a future date", c.date)
				}
			} else if err != nil {
				t.Errorf("%s should be accepted, got %v", c.date, err)
			}
		})
	}
}

func TestValidate_StatusAndLengths(t *testing.T) {
	t.Run("invalid status", func(t *testing.T) {
		in := validInput()
		in.Status = "pending"
		_, err := validate(in)
		if _, ok := fieldsOf(t, err)["status"]; !ok {
			t.Error("status 'pending' should be rejected")
		}
	})
	t.Run("status is case-insensitive", func(t *testing.T) {
		in := validInput()
		in.Status = "INACTIVE"
		e, err := validate(in)
		if err != nil || e.Status != model.StatusInactive {
			t.Errorf("expected inactive, got %v / %v", e, err)
		}
	})
	t.Run("name too short", func(t *testing.T) {
		in := validInput()
		in.Name = "A"
		_, err := validate(in)
		if _, ok := fieldsOf(t, err)["name"]; !ok {
			t.Error("1-character name should be rejected")
		}
	})
	t.Run("name too long", func(t *testing.T) {
		in := validInput()
		in.Name = strings.Repeat("a", 101)
		_, err := validate(in)
		if _, ok := fieldsOf(t, err)["name"]; !ok {
			t.Error("101-character name should be rejected")
		}
	})
}

func TestValidate_ReportsAllFieldsAtOnce(t *testing.T) {
	_, err := validate(model.EmployeeInput{})
	fields := fieldsOf(t, err)
	for _, f := range []string{"name", "email", "department", "designation", "joining_date"} {
		if _, ok := fields[f]; !ok {
			t.Errorf("expected error for %q in %v", f, fields)
		}
	}
}

// ---- CRUD behaviour through the service ----------------------------------

func TestService_CreateAndGet(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	ctx := context.Background()

	created, err := svc.Create(ctx, validInput())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Error("created employee should have an ID")
	}
	got, err := svc.Get(ctx, created.ID)
	if err != nil || got.Email != "asha.nair@example.com" {
		t.Errorf("get returned %v / %v", got, err)
	}
}

func TestService_DuplicateEmail(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	ctx := context.Background()

	if _, err := svc.Create(ctx, validInput()); err != nil {
		t.Fatal(err)
	}
	dup := validInput()
	dup.Name = "Someone Else"
	dup.Email = "ASHA.NAIR@example.com" // same address, different case
	if _, err := svc.Create(ctx, dup); !errors.Is(err, model.ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestService_UpdateRules(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	ctx := context.Background()

	a, _ := svc.Create(ctx, validInput())
	other := validInput()
	other.Email = "other@example.com"
	b, _ := svc.Create(ctx, other)

	t.Run("update own record keeping same email is allowed", func(t *testing.T) {
		in := validInput()
		in.Designation = "Senior Engineer"
		got, err := svc.Update(ctx, a.ID, in)
		if err != nil || got.Designation != "Senior Engineer" {
			t.Errorf("got %v / %v", got, err)
		}
	})
	t.Run("update to another employee's email is rejected", func(t *testing.T) {
		in := validInput()
		in.Email = "other@example.com"
		if _, err := svc.Update(ctx, a.ID, in); !errors.Is(err, model.ErrDuplicateEmail) {
			t.Errorf("expected ErrDuplicateEmail, got %v", err)
		}
	})
	t.Run("update unknown id", func(t *testing.T) {
		if _, err := svc.Update(ctx, 9999, validInput()); !errors.Is(err, model.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
	t.Run("invalid data is rejected before hitting the repo", func(t *testing.T) {
		in := validInput()
		in.Email = "nope"
		if _, err := svc.Update(ctx, b.ID, in); err == nil {
			t.Error("expected validation error")
		}
	})
}

func TestService_Delete(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	ctx := context.Background()

	e, _ := svc.Create(ctx, validInput())
	if err := svc.Delete(ctx, e.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.Get(ctx, e.ID); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
	if err := svc.Delete(ctx, e.ID); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("deleting twice should return ErrNotFound, got %v", err)
	}
}

func TestService_ListSearchAndFilters(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())
	ctx := context.Background()

	mk := func(name, email, dept, status string) {
		in := validInput()
		in.Name, in.Email, in.Department, in.Status = name, email, dept, status
		if _, err := svc.Create(ctx, in); err != nil {
			t.Fatal(err)
		}
	}
	mk("Asha Nair", "asha@example.com", "Engineering", "active")
	mk("Ravi Kumar", "ravi@example.com", "Finance", "active")
	mk("Meera Iyer", "meera@corp.com", "Engineering", "inactive")

	cases := []struct {
		name   string
		filter model.ListFilter
		want   int
	}{
		{"no filter", model.ListFilter{}, 3},
		{"search by name (case-insensitive)", model.ListFilter{Search: "ASHA"}, 1},
		{"search by email domain", model.ListFilter{Search: "corp.com"}, 1},
		{"search by exact id", model.ListFilter{Search: "2"}, 1},
		{"filter by department", model.ListFilter{Department: "Engineering"}, 2},
		{"filter by status", model.ListFilter{Status: "inactive"}, 1},
		{"department + status", model.ListFilter{Department: "Engineering", Status: "active"}, 1},
		{"no match", model.ListFilter{Search: "zzz"}, 0},
		{"status is normalised", model.ListFilter{Status: " ACTIVE "}, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := svc.List(ctx, c.filter)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != c.want {
				t.Errorf("want %d results, got %d", c.want, len(got))
			}
		})
	}
}
