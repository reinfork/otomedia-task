package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/ucup/taskmanager/internal/model"
	"gorm.io/gorm"
)

var ErrDuplicateTitle = errors.New("duplicate title")

// TaskRepository abstracts persistence so tests can swap Postgres for SQLite.
type TaskRepository interface {
	Create(task *model.Task) error
	GetByID(id uuid.UUID) (*model.Task, error)
	List(filter model.TaskFilter) ([]model.Task, int64, error)
	Update(task *model.Task) error
	SoftDelete(id uuid.UUID) error
	ExistsByTitleExcludingID(title string, excludeID uuid.UUID) (bool, error)
}

type gormTaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &gormTaskRepository{db: db}
}

func isDuplicateDBError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// MySQL 1062 duplicate entry, SQLite UNIQUE, Postgres 23505 — cover common drivers.
	return strings.Contains(msg, "duplicate") ||
		strings.Contains(msg, "unique") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "1062")
}

func (r *gormTaskRepository) Create(task *model.Task) error {
	if err := r.db.Create(task).Error; err != nil {
		if isDuplicateDBError(err) {
			return ErrDuplicateTitle
		}
		return err
	}
	return nil
}

func (r *gormTaskRepository) GetByID(id uuid.UUID) (*model.Task, error) {
	var t model.Task
	// GORM automatically adds deleted_at IS NULL for models with DeletedAt.
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *gormTaskRepository) List(filter model.TaskFilter) ([]model.Task, int64, error) {
	filter.Normalize()
	db := r.db.Model(&model.Task{})

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.Assignee != "" {
		// MySQL default collation (utf8mb4_*_ai_ci) is case-insensitive,
		// so LIKE behaves like Postgres ILIKE. SQLite is LIKE-insensitive for ASCII.
		db = db.Where("assignee LIKE ?", "%"+filter.Assignee+"%")
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		db = db.Where("(title LIKE ? OR description LIKE ?)", kw, kw)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = applySort(db, filter.Sort)

	var tasks []model.Task
	if err := db.Offset(filter.Offset()).Limit(filter.Limit).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	if tasks == nil {
		tasks = []model.Task{}
	}
	return tasks, total, nil
}

func applySort(db *gorm.DB, sort string) *gorm.DB {
	allowed := map[string]string{
		"created_at": "created_at ASC",
		"-created_at": "created_at DESC",
		"updated_at": "updated_at ASC",
		"-updated_at": "updated_at DESC",
		"title":      "title ASC",
		"-title":     "title DESC",
	}
	if order, ok := allowed[sort]; ok {
		return db.Order(order)
	}
	return db.Order("created_at DESC")
}

func (r *gormTaskRepository) Update(task *model.Task) error {
	if err := r.db.Save(task).Error; err != nil {
		if isDuplicateDBError(err) {
			return ErrDuplicateTitle
		}
		return err
	}
	return nil
}

func (r *gormTaskRepository) SoftDelete(id uuid.UUID) error {
	res := r.db.Where("id = ?", id).Delete(&model.Task{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *gormTaskRepository) ExistsByTitleExcludingID(title string, excludeID uuid.UUID) (bool, error) {
	var count int64
	db := r.db.Model(&model.Task{}).Where("title = ?", title)
	if excludeID != uuid.Nil {
		db = db.Where("id <> ?", excludeID)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
