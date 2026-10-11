package repository

import (
	"context"
	"database/sql"
	"fmt"

	"incident-flow/backend/internal/models"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *models.Project) error
	GetByJoinCode(ctx context.Context, joinCode string) (*models.Project, error)
	GetByID(ctx context.Context, id int) (*models.Project, error)
}

type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	projectQuery := `
		INSERT INTO projects (name, description, join_code, owner_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err = tx.QueryRowContext(ctx, projectQuery, project.Name, project.Description, project.JoinCode, project.OwnerID).
		Scan(&project.ID, &project.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}

	// Add the owner to the project_members table
	memberQuery := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`
	_, err = tx.ExecContext(ctx, memberQuery, project.ID, project.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to add owner to project_members: %w", err)
	}

	// Zatwierdzenie obu operacji w bazie
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *projectRepository) GetByJoinCode(ctx context.Context, joinCode string) (*models.Project, error) {
	query := `
		SELECT id, name, description, join_code, owner_id, created_at
		FROM projects
		WHERE join_code = $1
	`
	var project models.Project
	err := r.db.QueryRowContext(ctx, query, joinCode).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.JoinCode,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project by join code: %w", err)
	}

	return &project, nil
}

func (r *projectRepository) GetByID(ctx context.Context, id int) (*models.Project, error) {
	query := `
		SELECT id, name, description, join_code, owner_id, created_at
		FROM projects
		WHERE id = $1
	`
	var project models.Project
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.JoinCode,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project by id: %w", err)
	}

	return &project, nil
}
