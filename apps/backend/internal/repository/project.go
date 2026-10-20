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

	CreateJoinRequest(ctx context.Context, req *models.ProjectJoinRequest) error
	GetPendingJoinRequests(ctx context.Context, projectID int) ([]models.ProjectJoinRequest, error)
	GetJoinRequestByID(ctx context.Context, requestID int) (*models.ProjectJoinRequest, error)
	UpdateJoinRequestStatus(ctx context.Context, requestID int, status string) error
	AddProjectMember(ctx context.Context, projectID, userID int) error
	IsProjectMember(ctx context.Context, projectID, userID int) (bool, error)
	GetProjectMember(ctx context.Context, projectID, userID int) (*models.ProjectMember, error)
	GetProjectMembers(ctx context.Context, projectID int) ([]models.ProjectMember, error)
	HasPendingJoinRequest(ctx context.Context, projectID, userID int) (bool, error)
	RemoveProjectMember(ctx context.Context, projectID, userID int) error
	LeaveProject(ctx context.Context, projectID, userID int) error
	UpdateProjectMemberRole(ctx context.Context, input UdpateProjectMemberRoleInput) error
	UpdateProject(ctx context.Context, input UpdateProjectInput) error
	DeleteProject(ctx context.Context, projectID int) error
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

func (r *projectRepository) CreateJoinRequest(ctx context.Context, req *models.ProjectJoinRequest) error {
	query := `
		INSERT INTO project_join_requests (project_id, user_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, status, created_at
	`
	err := r.db.QueryRowContext(ctx, query, req.ProjectID, req.UserID).Scan(&req.ID, &req.Status, &req.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create join request: %w", err)
	}
	return nil
}

func (r *projectRepository) GetPendingJoinRequests(ctx context.Context, projectID int) ([]models.ProjectJoinRequest, error) {
	query := `
		SELECT id, project_id, user_id, status, created_at
		FROM project_join_requests
		WHERE project_id = $1 AND status = 'pending'
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending join requests: %w", err)
	}
	defer rows.Close()

	var requests []models.ProjectJoinRequest
	for rows.Next() {
		var req models.ProjectJoinRequest
		err := rows.Scan(&req.ID, &req.ProjectID, &req.UserID, &req.Status, &req.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan join request: %w", err)
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return requests, nil
}

func (r *projectRepository) GetJoinRequestByID(ctx context.Context, requestID int) (*models.ProjectJoinRequest, error) {
	query := `
		SELECT id, project_id, user_id, status, created_at
		FROM project_join_requests
		WHERE id = $1
	`

	var req models.ProjectJoinRequest
	err := r.db.QueryRowContext(ctx, query, requestID).Scan(&req.ID, &req.ProjectID, &req.UserID, &req.Status, &req.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get join request by id: %w", err)
	}
	return &req, nil
}

func (r *projectRepository) UpdateJoinRequestStatus(ctx context.Context, requestID int, status string) error {
	query := `
		UPDATE project_join_requests
		SET status = $1::join_request_status
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, status, requestID)
	if err != nil {
		return fmt.Errorf("failed to update join request status: %w", err)
	}
	return nil
}

func (r *projectRepository) AddProjectMember(ctx context.Context, projectID, userID int) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'member'::project_role)
		ON CONFLICT (project_id, user_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, projectID, userID)

	if err != nil {
		return fmt.Errorf("failed to add project member: %w", err)
	}
	return nil
}

func (r *projectRepository) IsProjectMember(ctx context.Context, projectID, userID int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check project membership: %w", err)
	}
	return count == 1, nil
}

func (r *projectRepository) GetProjectMember(ctx context.Context, projectID, userID int) (*models.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, joined_at
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	var member models.ProjectMember
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(
		&member.ID,
		&member.ProjectID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project member: %w", err)
	}
	return &member, nil
}

func (r *projectRepository) GetProjectMembers(ctx context.Context, projectID int) ([]models.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, joined_at
		FROM project_members
		WHERE project_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}
	defer rows.Close()

	var members []models.ProjectMember
	for rows.Next() {
		var member models.ProjectMember
		err := rows.Scan(&member.ID, &member.ProjectID, &member.UserID, &member.Role, &member.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return members, nil
}

func (r *projectRepository) HasPendingJoinRequest(ctx context.Context, projectID, userID int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM project_join_requests
		WHERE project_id = $1 AND user_id = $2 AND status = 'pending'
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, projectID, userID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check pending join request: %w", err)
	}
	return count > 0, nil
}

func (r *projectRepository) RemoveProjectMember(ctx context.Context, projectID, userID int) error {
	query := `
		DELETE FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, projectID, userID)

	if err != nil {
		return fmt.Errorf("failed to remove project member: %w", err)
	}

	return nil
}

func (r *projectRepository) LeaveProject(ctx context.Context, projectID, userID int) error {
	query := `
		DELETE FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to leave project: %w", err)
	}

	return nil
}

type UdpateProjectMemberRoleInput struct {
	ProjectID int
	UserID    int
	NewRole   string
}

func (r *projectRepository) UpdateProjectMemberRole(ctx context.Context, input UdpateProjectMemberRoleInput) error {
	query := `
        UPDATE project_members
        SET role = $1::project_role
        WHERE project_id = $2 AND user_id = $3
    `
	_, err := r.db.ExecContext(ctx, query, input.NewRole, input.ProjectID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to update project member role: %w", err)
	}

	return nil
}

type UpdateProjectInput struct {
	ProjectID   int
	Name        string
	Description string
}

func (r *projectRepository) UpdateProject(ctx context.Context, input UpdateProjectInput) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, input.Name, input.Description, input.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

func (r *projectRepository) DeleteProject(ctx context.Context, projectID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	// Delete project members
	_, err = tx.ExecContext(ctx, "DELETE FROM project_members WHERE project_id = $1", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project members: %w", err)
	}

	// Delete the project
	_, err = tx.ExecContext(ctx, "DELETE FROM projects WHERE id = $1", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
