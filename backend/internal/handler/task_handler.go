package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ucup/taskmanager/internal/model"
	"github.com/ucup/taskmanager/internal/service"
	"github.com/ucup/taskmanager/pkg/cache"
)

// ErrorResponse is the consistent envelope for all API errors.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}

func serviceErrorToHTTP(err error) (int, string, string) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", err.Error()
	case errors.Is(err, service.ErrConflict):
		return http.StatusConflict, "CONFLICT", err.Error()
	case errors.Is(err, service.ErrTitleRequired),
		errors.Is(err, service.ErrInvalidStatus),
		errors.Is(err, service.ErrValidation),
		errors.Is(err, service.ErrInvalidPayload):
		return http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"
	}
}

type TaskHandler struct {
	svc   service.TaskService
	cache cache.Cache
}

func NewTaskHandler(svc service.TaskService, c cache.Cache) *TaskHandler {
	return &TaskHandler{svc: svc, cache: c}
}

func (h *TaskHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/tasks", h.ListTasks)
		api.GET("/tasks/:id", h.GetTask)
		api.POST("/tasks", h.CreateTask)
		api.PUT("/tasks/:id", h.UpdateTask)
		api.DELETE("/tasks/:id", h.DeleteTask)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func parseUUIDParam(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid task id")
		return uuid.Nil, false
	}
	return id, true
}

// ListTasks implements filtering (status, keyword, assignee, page, limit, sort)
// with a 60s Redis cache whose key includes query parameters.
func (h *TaskHandler) ListTasks(c *gin.Context) {
	filter := model.TaskFilter{
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
		Assignee: c.Query("assignee"),
		Sort:     c.DefaultQuery("sort", "-created_at"),
	}
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter.Normalize()

	if filter.Status != "" && !model.IsValidStatus(filter.Status) {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid status: must be todo, in_progress, or done")
		return
	}

	cacheKey := ""
	if h.cache != nil {
		cacheKey = cache.BuildTasksKey(c.Request.URL.Query())
		if cached, err := h.cache.Get(c.Request.Context(), cacheKey); err == nil {
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json", []byte(cached))
			return
		} else if !errors.Is(err, redis.Nil) {
			// Cache errors must not break the request; fall through to DB.
		}
	}

	tasks, total, err := h.svc.List(filter)
	if err != nil {
		status, code, msg := serviceErrorToHTTP(err)
		respondError(c, status, code, msg)
		return
	}

	totalPages := 0
	if filter.Limit > 0 {
		totalPages = int((total + int64(filter.Limit) - 1) / int64(filter.Limit))
	}
	body := gin.H{
		"data": tasks,
		"meta": gin.H{
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
	}

	// Serialize once so the exact bytes served are what we cache.
	// (Keeps HIT responses byte-identical to MISS responses.)
	if h.cache != nil && cacheKey != "" {
		// Best-effort cache write; ignore errors.
		// We render via JSON first using gin's renderer indirectly by caching
		// the raw payload through a small helper: re-encode deterministically
		// is unnecessary — instead we cache after c.JSON via a writer wrapper
		// would add complexity, so we store via a direct marshal below.
		// Simpler: let c.JSON serve, and cache the same struct.
		_ = cacheKey
	}

	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, body)

	// Populate cache asynchronously-best-effort is overkill; do it inline
	// by re-marshalling the same payload (identical semantics).
	if h.cache != nil && cacheKey != "" {
		// Marshal the same body we just sent.
		// Use gin's JSON marshal path via encoding/json for stability.
		payload := marshalBody(body)
		if payload != "" {
			_ = h.cache.Set(c.Request.Context(), cacheKey, payload, cache.TaskListTTL)
		}
	}
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	t, err := h.svc.GetByID(id)
	if err != nil {
		status, code, msg := serviceErrorToHTTP(err)
		respondError(c, status, code, msg)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var input service.CreateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request payload")
		return
	}
	t, err := h.svc.Create(input)
	if err != nil {
		status, code, msg := serviceErrorToHTTP(err)
		respondError(c, status, code, msg)
		return
	}
	_ = cache.InvalidateTaskLists(c.Request.Context(), h.cache)
	c.JSON(http.StatusCreated, gin.H{"data": t})
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	var input service.UpdateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request payload")
		return
	}
	t, err := h.svc.Update(id, input)
	if err != nil {
		status, code, msg := serviceErrorToHTTP(err)
		respondError(c, status, code, msg)
		return
	}
	_ = cache.InvalidateTaskLists(c.Request.Context(), h.cache)
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, ok := parseUUIDParam(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		status, code, msg := serviceErrorToHTTP(err)
		respondError(c, status, code, msg)
		return
	}
	_ = cache.InvalidateTaskLists(c.Request.Context(), h.cache)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "task deleted"}})
}
