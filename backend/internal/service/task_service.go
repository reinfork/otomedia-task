package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ucup/taskmanager/internal/model"
	"github.com/ucup/taskmanager/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("task not found")
	ErrConflict       = errors.New("task title already exists")
	ErrValidation     = errors.New("validation failed")
	ErrInvalidStatus  = errors.New("invalid status: must be todo, in_progress, or done")
	ErrTitleRequired  = errors.New("title is required")
	ErrInvalidPayload = errors.New("invalid request payload")
)

type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Assignee    string `json:"assignee"`
}

type UpdateTaskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Assignee    *string `json:"assignee"`
}

type TaskService interface {
	Create(input CreateTaskInput) (*model.Task, error)
	GetByID(id uuid.UUID) (*model.Task, error)
	List(filter model.TaskFilter) ([]model.Task, int64, error)
	Update(id uuid.UUID, input UpdateTaskInput) (*model.Task, error)
	Delete(id uuid.UUID) error
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) Create(input CreateTaskInput) (*model.Task, error) {
	if input.Title == "" {
		return nil, ErrTitleRequired
	}
	status := input.Status
	if status == "" {
		status = model.StatusTodo
	}
	if !model.IsValidStatus(status) {
		return nil, ErrInvalidStatus
	}
	exists, err := s.repo.ExistsByTitleExcludingID(input.Title, uuid.Nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict
	}
	t := &model.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      status,
		Assignee:    input.Assignee,
	}
	if err := s.repo.Create(t); err != nil {
		if errors.Is(err, repository.ErrDuplicateTitle) {
			return nil, ErrConflict
		}
		return nil, err
	}
	return t, nil
}

func (s *taskService) GetByID(id uuid.UUID) (*model.Task, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *taskService) List(filter model.TaskFilter) ([]model.Task, int64, error) {
	filter.Normalize()
	if filter.Status != "" && !model.IsValidStatus(filter.Status) {
		return nil, 0, ErrInvalidStatus
	}
	return s.repo.List(filter)
}

func (s *taskService) Update(id uuid.UUID, input UpdateTaskInput) (*model.Task, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if input.Title != nil {
		if *input.Title == "" {
			return nil, ErrTitleRequired
		}
		if *input.Title != t.Title {
			exists, err := s.repo.ExistsByTitleExcludingID(*input.Title, id)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrConflict
			}
			t.Title = *input.Title
		}
	}
	if input.Description != nil {
		t.Description = *input.Description
	}
	if input.Status != nil {
		if !model.IsValidStatus(*input.Status) {
			return nil, ErrInvalidStatus
		}
		t.Status = *input.Status
	}
	if input.Assignee != nil {
		t.Assignee = *input.Assignee
	}
	if err := s.repo.Update(t); err != nil {
		if errors.Is(err, repository.ErrDuplicateTitle) {
			return nil, ErrConflict
		}
		return nil, err
	}
	return t, nil
}

func (s *taskService) Delete(id uuid.UUID) error {
	if err := s.repo.SoftDelete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
