package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ucup/taskmanager/internal/model"
	"github.com/ucup/taskmanager/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) TaskService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}))
	repo := repository.NewTaskRepository(db)
	return NewTaskService(repo)
}

func strPtr(s string) *string { return &s }

func TestCreateAndGet(t *testing.T) {
	svc := newTestService(t)
	created, err := svc.Create(CreateTaskInput{Title: "First task", Status: "todo", Assignee: "budi"})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	got, err := svc.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "First task", got.Title)
	assert.Equal(t, "todo", got.Status)
}

func TestCreateDuplicateTitleReturns409Conflict(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Create(CreateTaskInput{Title: "Unique title"})
	require.NoError(t, err)
	_, err = svc.Create(CreateTaskInput{Title: "Unique title"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestUpdateTask(t *testing.T) {
	svc := newTestService(t)
	created, err := svc.Create(CreateTaskInput{Title: "To update", Status: "todo"})
	require.NoError(t, err)

	updated, err := svc.Update(created.ID, UpdateTaskInput{
		Title:  strPtr("Updated title"),
		Status: strPtr("done"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated title", updated.Title)
	assert.Equal(t, "done", updated.Status)

	// Updating a missing task -> 404.
	_, err = svc.Update(uuid.New(), UpdateTaskInput{Title: strPtr("x")})
	assert.ErrorIs(t, err, ErrNotFound)

	// Updating to a duplicate title -> 409.
	_, err = svc.Create(CreateTaskInput{Title: "Other"})
	require.NoError(t, err)
	_, err = svc.Update(created.ID, UpdateTaskInput{Title: strPtr("Other")})
	assert.ErrorIs(t, err, ErrConflict)

	// Invalid status -> validation error.
	_, err = svc.Update(created.ID, UpdateTaskInput{Status: strPtr("nope")})
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestSearchFiltering(t *testing.T) {
	svc := newTestService(t)
	_, _ = svc.Create(CreateTaskInput{Title: "Fix login bug", Description: "auth fails", Status: "todo", Assignee: "budi"})
	_, _ = svc.Create(CreateTaskInput{Title: "Write docs", Description: "readme", Status: "done", Assignee: "sari"})
	_, _ = svc.Create(CreateTaskInput{Title: "Fix payment bug", Description: "midtrans", Status: "in_progress", Assignee: "budi"})

	// Keyword search.
	tasks, total, err := svc.List(model.TaskFilter{Keyword: "fix", Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, tasks, 2)

	// Status filter.
	tasks, total, err = svc.List(model.TaskFilter{Status: "done", Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Write docs", tasks[0].Title)

	// Assignee filter.
	_, total, err = svc.List(model.TaskFilter{Assignee: "budi", Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	// Pagination.
	tasks, total, err = svc.List(model.TaskFilter{Page: 2, Limit: 2, Sort: "created_at"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, tasks, 1)

	// Invalid status rejected.
	_, _, err = svc.List(model.TaskFilter{Status: "bogus"})
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestSoftDeleteHidesTask(t *testing.T) {
	svc := newTestService(t)
	created, err := svc.Create(CreateTaskInput{Title: "Temp"})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(created.ID))

	_, err = svc.GetByID(created.ID)
	assert.ErrorIs(t, err, ErrNotFound)

	tasks, total, err := svc.List(model.TaskFilter{Page: 1, Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, tasks)
}
