package models

import "time"

type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	JoinCode    string    `json:"join_code"`
	OwnerID     int       `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectMember struct {
	ID        int         `json:"id"`
	ProjectID int         `json:"project_id"`
	UserID    int         `json:"user_id"`
	Role      ProjectRole `json:"role"`
	JoinedAt  time.Time   `json:"joined_at"`
}

type ProjectJoinRequest struct {
	ID        int               `json:"id"`
	ProjectID int               `json:"project_id"`
	UserID    int               `json:"user_id"`
	Status    JoinRequestStatus `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
}

type Incident struct {
	ID          int              `json:"id"`
	ProjectID   int              `json:"project_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Priority    IncidentPriority `json:"priority"`
	Status      IncidentStatus   `json:"status"`
	AssignedTo  int              `json:"assigned_to"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
