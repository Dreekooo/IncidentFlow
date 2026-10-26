package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"incident-flow/backend/internal/models"
)

type IncidentRepository interface {
	CreateIncident(ctx context.Context, incident *models.Incident) error
	GetIncidentByID(ctx context.Context, projectID int, incidentID int) (*models.Incident, error)
	ListIncidentsByProjectID(ctx context.Context, projectID int) ([]*models.Incident, error)
	UpdateIncident(ctx context.Context, incident *models.Incident) error
	AssignIncident(ctx context.Context, projectID int, incidentID int, userID int) error
	DeleteIncident(ctx context.Context, projectID int, incidentID int) error
}

type incidentRepository struct {
	db *sql.DB
}

func NewIncidentRepository(db *sql.DB) IncidentRepository {
	return &incidentRepository{db: db}
}

func (r *incidentRepository) CreateIncident(ctx context.Context, incident *models.Incident) error {
	query := `
		INSERT INTO incidents (project_id, assigned_to, title, description, priority, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::incident_priority, $6::incident_status, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	var id int
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(
		ctx,
		query,
		incident.ProjectID,
		incident.AssignedTo,
		incident.Title,
		incident.Description,
		incident.Priority,
		incident.Status,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	incident.ID = id
	incident.CreatedAt = createdAt
	incident.UpdatedAt = updatedAt

	return nil
}

func (r *incidentRepository) GetIncidentByID(ctx context.Context, projectID int, incidentID int) (*models.Incident, error) {
	query := `
		SELECT id, project_id, assigned_to, title, description, priority, status, created_at, updated_at
		FROM incidents
		WHERE id = $1 AND project_id = $2
	`

	var incident models.Incident
	err := r.db.QueryRowContext(ctx, query, incidentID, projectID).Scan(
		&incident.ID,
		&incident.ProjectID,
		&incident.AssignedTo,
		&incident.Title,
		&incident.Description,
		&incident.Priority,
		&incident.Status,
		&incident.CreatedAt,
		&incident.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("incident not found")
		}
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return &incident, nil
}

func (r *incidentRepository) ListIncidentsByProjectID(ctx context.Context, projectID int) ([]*models.Incident, error) {
	query := `
		SELECT id, project_id, title, description, priority, status, assigned_to, created_at, updated_at
		FROM incidents
		WHERE project_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer rows.Close()

	incidents := make([]*models.Incident, 0)
	for rows.Next() {
		var incident models.Incident
		err := rows.Scan(
			&incident.ID,
			&incident.ProjectID,
			&incident.Title,
			&incident.Description,
			&incident.Priority,
			&incident.Status,
			&incident.AssignedTo,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over incidents: %w", err)
	}

	return incidents, nil
}

func (r *incidentRepository) UpdateIncident(ctx context.Context, incident *models.Incident) error {
	query := `
		UPDATE incidents
		SET title = $1, description = $2, priority = $3::incident_priority, status = $4::incident_status, assigned_to = $5, updated_at = NOW()
		WHERE id = $6 AND project_id = $7
	`
	result, err := r.db.ExecContext(ctx, query, incident.Title, incident.Description, incident.Priority, incident.Status, incident.AssignedTo, incident.ID, incident.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}

func (r *incidentRepository) AssignIncident(ctx context.Context, projectID int, incidentID int, userID int) error {
	query := `
		UPDATE incidents
		SET assigned_to = $1, updated_at = NOW()
		WHERE id = $2 AND project_id = $3
	`
	result, err := r.db.ExecContext(ctx, query, userID, incidentID, projectID)
	if err != nil {
		return fmt.Errorf("failed to assign incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}

func (r *incidentRepository) DeleteIncident(ctx context.Context, projectID int, incidentID int) error {
	query := `
		DELETE FROM incidents
		WHERE id = $1 AND project_id = $2
	`
	result, err := r.db.ExecContext(ctx, query, incidentID, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}
