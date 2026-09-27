package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ucup/taskmanager/internal/model"
	"github.com/ucup/taskmanager/internal/repository"
	"github.com/ucup/taskmanager/internal/service"
	"github.com/ucup/taskmanager/pkg/cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestRouter(t *testing.T) (*gin.Engine, *miniredis.Miniredis) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}))

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rc := cache.NewRedisCacheFromClient(redis.NewClient(&redis.Options{Addr: mr.Addr()}))

	repo := repository.NewTaskRepository(db)
	svc := service.NewTaskService(repo)
	h := NewTaskHandler(svc, rc)

	r := gin.New()
	h.RegisterRoutes(r)
	return r, mr
}

func doRequest(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestListCachingMissThenHit(t *testing.T) {
	r, _ := newTestRouter(t)

	w := doRequest(t, r, "POST", "/api/tasks", map[string]string{"title": "Cached task", "status": "todo"})
	require.Equal(t, http.StatusCreated, w.Code)

	w1 := doRequest(t, r, "GET", "/api/tasks?status=todo", nil)
	require.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, "MISS", w1.Header().Get("X-Cache"))

	w2 := doRequest(t, r, "GET", "/api/tasks?status=todo", nil)
	require.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "HIT", w2.Header().Get("X-Cache"))
	assert.JSONEq(t, w1.Body.String(), w2.Body.String())
}

func TestWriteInvalidatesCache(t *testing.T) {
	r, _ := newTestRouter(t)
	w := doRequest(t, r, "POST", "/api/tasks", map[string]string{"title": "T1"})
	require.Equal(t, http.StatusCreated, w.Code)

	w1 := doRequest(t, r, "GET", "/api/tasks", nil)
	require.Equal(t, "MISS", w1.Header().Get("X-Cache"))
	w2 := doRequest(t, r, "GET", "/api/tasks", nil)
	require.Equal(t, "HIT", w2.Header().Get("X-Cache"))

	// Create invalidates -> next read is MISS again.
	w3 := doRequest(t, r, "POST", "/api/tasks", map[string]string{"title": "T2"})
	require.Equal(t, http.StatusCreated, w3.Code)
	w4 := doRequest(t, r, "GET", "/api/tasks", nil)
	assert.Equal(t, "MISS", w4.Header().Get("X-Cache"))
}

func TestDuplicateTitleReturns409(t *testing.T) {
	r, _ := newTestRouter(t)
	w := doRequest(t, r, "POST", "/api/tasks", map[string]string{"title": "Same"})
	require.Equal(t, http.StatusCreated, w.Code)

	w2 := doRequest(t, r, "POST", "/api/tasks", map[string]string{"title": "Same"})
	assert.Equal(t, http.StatusConflict, w2.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &payload))
	assert.Contains(t, w2.Body.String(), "CONFLICT")
}

func TestUpdateAndDeleteFlow(t *testing.T) {
	r, _ := newTestRouter(t)
	w := doRequest(t, r, "POST", "/api/tasks", map[string]string{"title": "Flow", "status": "todo"})
	require.Equal(t, http.StatusCreated, w.Code)
	var created struct {
		Data model.Task `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	// PUT update.
	w2 := doRequest(t, r, "PUT", "/api/tasks/"+created.Data.ID.String(), map[string]string{"title": "Flow v2", "status": "done"})
	assert.Equal(t, http.StatusOK, w2.Code)

	// GET reflects update.
	w3 := doRequest(t, r, "GET", "/api/tasks/"+created.Data.ID.String(), nil)
	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), "Flow v2")

	// DELETE soft-deletes.
	w4 := doRequest(t, r, "DELETE", "/api/tasks/"+created.Data.ID.String(), nil)
	assert.Equal(t, http.StatusOK, w4.Code)

	// GET after delete -> 404, list hides it.
	w5 := doRequest(t, r, "GET", "/api/tasks/"+created.Data.ID.String(), nil)
	assert.Equal(t, http.StatusNotFound, w5.Code)
	w6 := doRequest(t, r, "GET", "/api/tasks", nil)
	assert.NotContains(t, w6.Body.String(), "Flow v2")
}

func TestConsistentErrorShape(t *testing.T) {
	r, _ := newTestRouter(t)
	w := doRequest(t, r, "GET", "/api/tasks/not-a-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code"`)
	assert.Contains(t, w.Body.String(), `"message"`)
}
