package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"employee-management/internal/model"

	"github.com/gin-gonic/gin"
)

// stubService lets each test decide what the service returns, so these tests
// check only the HTTP layer: routing, status codes and JSON shapes.
type stubService struct {
	createFn func(model.EmployeeInput) (*model.Employee, error)
	listFn   func(model.ListFilter) ([]model.Employee, error)
	getFn    func(int64) (*model.Employee, error)
	updateFn func(int64, model.EmployeeInput) (*model.Employee, error)
	deleteFn func(int64) error
}

func (s *stubService) Create(_ context.Context, in model.EmployeeInput) (*model.Employee, error) {
	return s.createFn(in)
}
func (s *stubService) List(_ context.Context, f model.ListFilter) ([]model.Employee, error) {
	return s.listFn(f)
}
func (s *stubService) Get(_ context.Context, id int64) (*model.Employee, error) {
	return s.getFn(id)
}
func (s *stubService) Update(_ context.Context, id int64, in model.EmployeeInput) (*model.Employee, error) {
	return s.updateFn(id, in)
}
func (s *stubService) Delete(_ context.Context, id int64) error { return s.deleteFn(id) }
func (s *stubService) Departments(_ context.Context) ([]string, error) {
	return []string{"Engineering"}, nil
}

func newRouter(svc *stubService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewEmployeeHandler(svc).RegisterRoutes(r.Group("/api/v1"))
	return r
}

func do(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var sample = &model.Employee{ID: 1, Name: "Asha Nair", Email: "asha@example.com"}

const validBody = `{"name":"Asha Nair","email":"asha@example.com","department":"Engineering","designation":"Dev","joining_date":"2024-06-15"}`

func TestCreate_Returns201(t *testing.T) {
	r := newRouter(&stubService{createFn: func(model.EmployeeInput) (*model.Employee, error) { return sample, nil }})
	w := do(r, "POST", "/api/v1/employees", validBody)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body)
	}
	var got model.Employee
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.ID != 1 {
		t.Errorf("unexpected body: %s", w.Body)
	}
}

func TestCreate_MalformedJSON_Returns400(t *testing.T) {
	r := newRouter(&stubService{})
	w := do(r, "POST", "/api/v1/employees", `{not json`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestCreate_ValidationError_Returns400WithFields(t *testing.T) {
	r := newRouter(&stubService{createFn: func(model.EmployeeInput) (*model.Employee, error) {
		return nil, &model.ValidationError{Fields: map[string]string{"email": "bad email"}}
	}})
	w := do(r, "POST", "/api/v1/employees", `{}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
	var body struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Fields["email"] != "bad email" {
		t.Errorf("fields not returned: %s", w.Body)
	}
}

func TestCreate_DuplicateEmail_Returns409(t *testing.T) {
	r := newRouter(&stubService{createFn: func(model.EmployeeInput) (*model.Employee, error) {
		return nil, model.ErrDuplicateEmail
	}})
	w := do(r, "POST", "/api/v1/employees", validBody)
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", w.Code)
	}
	var body struct {
		Fields map[string]string `json:"fields"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Fields["email"] == "" {
		t.Errorf("409 should include an email field message: %s", w.Body)
	}
}

func TestGet_StatusCodes(t *testing.T) {
	r := newRouter(&stubService{getFn: func(id int64) (*model.Employee, error) {
		if id == 1 {
			return sample, nil
		}
		return nil, model.ErrNotFound
	}})

	cases := []struct {
		name string
		path string
		want int
	}{
		{"found", "/api/v1/employees/1", http.StatusOK},
		{"not found", "/api/v1/employees/99", http.StatusNotFound},
		{"non-numeric id", "/api/v1/employees/abc", http.StatusBadRequest},
		{"zero id", "/api/v1/employees/0", http.StatusBadRequest},
		{"negative id", "/api/v1/employees/-5", http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if w := do(r, "GET", c.path, ""); w.Code != c.want {
				t.Errorf("want %d, got %d", c.want, w.Code)
			}
		})
	}
}

func TestList_PassesFiltersAndWrapsResponse(t *testing.T) {
	var seen model.ListFilter
	r := newRouter(&stubService{listFn: func(f model.ListFilter) ([]model.Employee, error) {
		seen = f
		return []model.Employee{*sample}, nil
	}})
	w := do(r, "GET", "/api/v1/employees?search=asha&department=Engineering&status=active", "")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if seen.Search != "asha" || seen.Department != "Engineering" || seen.Status != "active" {
		t.Errorf("filters not passed through: %+v", seen)
	}
	var body struct {
		Data  []model.Employee `json:"data"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Total != 1 || len(body.Data) != 1 {
		t.Errorf("unexpected body: %s", w.Body)
	}
}

func TestList_Empty_ReturnsEmptyArrayNotNull(t *testing.T) {
	r := newRouter(&stubService{listFn: func(model.ListFilter) ([]model.Employee, error) {
		return []model.Employee{}, nil
	}})
	w := do(r, "GET", "/api/v1/employees", "")
	if got := w.Body.String(); got != `{"data":[],"total":0}` {
		t.Errorf("unexpected body: %s", got)
	}
}

func TestUpdate_StatusCodes(t *testing.T) {
	cases := []struct {
		name string
		path string
		body string
		err  error
		want int
	}{
		{"ok", "/api/v1/employees/1", validBody, nil, http.StatusOK},
		{"not found", "/api/v1/employees/1", validBody, model.ErrNotFound, http.StatusNotFound},
		{"duplicate email", "/api/v1/employees/1", validBody, model.ErrDuplicateEmail, http.StatusConflict},
		{"bad id", "/api/v1/employees/x", validBody, nil, http.StatusBadRequest},
		{"bad json", "/api/v1/employees/1", `oops`, nil, http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := newRouter(&stubService{updateFn: func(int64, model.EmployeeInput) (*model.Employee, error) {
				if c.err != nil {
					return nil, c.err
				}
				return sample, nil
			}})
			if w := do(r, "PUT", c.path, c.body); w.Code != c.want {
				t.Errorf("want %d, got %d: %s", c.want, w.Code, w.Body)
			}
		})
	}
}

func TestDelete_StatusCodes(t *testing.T) {
	cases := []struct {
		name string
		path string
		err  error
		want int
	}{
		{"deleted", "/api/v1/employees/1", nil, http.StatusNoContent},
		{"not found", "/api/v1/employees/1", model.ErrNotFound, http.StatusNotFound},
		{"bad id", "/api/v1/employees/abc", nil, http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := newRouter(&stubService{deleteFn: func(int64) error { return c.err }})
			if w := do(r, "DELETE", c.path, ""); w.Code != c.want {
				t.Errorf("want %d, got %d", c.want, w.Code)
			}
		})
	}
}

func TestUnexpectedError_Returns500WithoutLeakingDetails(t *testing.T) {
	r := newRouter(&stubService{getFn: func(int64) (*model.Employee, error) {
		return nil, context.DeadlineExceeded
	}})
	w := do(r, "GET", "/api/v1/employees/1", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
	if bytes.Contains(w.Body.Bytes(), []byte("deadline")) {
		t.Errorf("internal error text leaked to client: %s", w.Body)
	}
}
