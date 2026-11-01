package postgres

import (
	"context"
	"devboard/internal/domain"
	generated "devboard/internal/repository/postgres/generated"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TaskRepository implements UseCase.TaskRepository
type TaskRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewTaskRepository constructor
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

func toDomainTask(g generated.Task) *domain.Task {
	task := &domain.Task{
		ID:          g.ID.String(),
		ProjectID:   g.ProjectID.String(),
		Title:       g.Title,
		Description: g.Description,
		Status:      domain.TaskStatus(g.Status),
		CreatedBy:   g.CreatedBy.String(),
		CreatedAt:   g.CreatedAt.Time,
		UpdatedAt:   g.UpdatedAt.Time,
	}

	if g.AssigneeID.Valid {
		id := g.AssigneeID.String()
		task.AssigneeID = &id
	}

	return task
}

// Create new task in database
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	var projectID pgtype.UUID
	if err := projectID.Scan(task.ProjectID); err != nil {
		return fmt.Errorf("project_id invalido: %w", err)
	}

	var createdBy pgtype.UUID
	if err := createdBy.Scan(task.CreatedBy); err != nil {
		return fmt.Errorf("created_by invalido: %w", err)
	}

	result, err := r.queries.CreateTask(ctx, generated.CreateTaskParams{
		ProjectID:   projectID,
		CreatedBy:   createdBy,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
	})

	if err != nil {
		return fmt.Errorf("insertando tarea en BD: %w", err)
	}
	task.ID = result.ID.String()
	task.CreatedAt = result.CreatedAt.Time
	task.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// Update update task in database
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	var id pgtype.UUID
	if err := id.Scan(task.ID); err != nil {
		return fmt.Errorf("project_id invalido: %w", err)
	}

	var assigneeID pgtype.UUID
	if task.AssigneeID != nil {
		if err := assigneeID.Scan(*task.AssigneeID); err != nil {
			return fmt.Errorf("created_by invalido: %w", err)
		}
	}

	result, err := r.queries.UpdateTask(ctx, generated.UpdateTaskParams{
		ID:          id,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		AssigneeID:  assigneeID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscando tarea: %w", domain.ErrNotFound)

		}
		return fmt.Errorf("actualizando tarea %s: %w", id, err)
	}
	task.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID get one user by id
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	var pgID pgtype.UUID
	if err := pgID.Scan(id); err != nil {
		return nil, fmt.Errorf("id invalido %s: %w", id, domain.ErrInvalidInput)
	}

	result, err := r.queries.GetTaskByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("tarea: %w", domain.ErrNotFound)

		}
		return nil, fmt.Errorf("buscando tarea %s: %w", id, err)
	}

	return toDomainTask(result), nil
}

// ListByProject function
func (r *TaskRepository) ListByProject(ctx context.Context, projectID string) ([]*domain.Task, error) {
	var pgProjectID pgtype.UUID
	if err := pgProjectID.Scan(projectID); err != nil {
		return nil, fmt.Errorf("id invalido %s: %w", projectID, domain.ErrInvalidInput)
	}

	tasks, err := r.queries.ListTaskByProject(ctx, pgProjectID)
	if err != nil {
		return nil, fmt.Errorf("buscando tareas de proyecto %s: %w", projectID, err)
	}

	result := make([]*domain.Task, len(tasks))

	for i, task := range tasks {
		result[i] = toDomainTask(task)
	}
	return result, nil
}
