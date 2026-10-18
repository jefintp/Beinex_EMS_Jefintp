package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"employee-management/internal/model"
	"employee-management/internal/service"

	"github.com/gin-gonic/gin"
)

// EmployeeHandler exposes the employee REST endpoints.
type EmployeeHandler struct {
	svc service.EmployeeService
}

// NewEmployeeHandler creates a handler backed by the given service.
func NewEmployeeHandler(svc service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

// RegisterRoutes mounts the endpoints on the given router group.
func (h *EmployeeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/employees", h.create)
	rg.GET("/employees", h.list)
	rg.GET("/employees/:id", h.get)
	rg.PUT("/employees/:id", h.update)
	rg.DELETE("/employees/:id", h.delete)
	rg.GET("/departments", h.departments)
}

type errorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func (h *EmployeeHandler) create(c *gin.Context) {
	var in model.EmployeeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, errBadJSON)
		return
	}
	e, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, e)
}

func (h *EmployeeHandler) list(c *gin.Context) {
	filter := model.ListFilter{
		Search:     c.Query("search"),
		Department: c.Query("department"),
		Status:     c.Query("status"),
	}
	employees, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": employees, "total": len(employees)})
}

func (h *EmployeeHandler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	e, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *EmployeeHandler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in model.EmployeeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, errBadJSON)
		return
	}
	e, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, e)
}

func (h *EmployeeHandler) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EmployeeHandler) departments(c *gin.Context) {
	departments, err := h.svc.Departments(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": departments})
}

var errBadJSON = errors.New("request body is not valid JSON")

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "id must be a positive integer"})
		return 0, false
	}
	return id, true
}

// respondError maps domain errors to HTTP status codes.
func respondError(c *gin.Context, err error) {
	var ve *model.ValidationError
	switch {
	case errors.As(err, &ve):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "validation failed", Fields: ve.Fields})
	case errors.Is(err, errBadJSON):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, model.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, model.ErrDuplicateEmail):
		c.JSON(http.StatusConflict, errorResponse{
			Error:  err.Error(),
			Fields: map[string]string{"email": err.Error()},
		})
	default:
		log.Printf("internal error: %v", err)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
