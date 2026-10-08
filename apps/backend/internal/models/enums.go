package models

type ProjectRole string

const (
	ProjectRoleOwner  ProjectRole = "owner"
	ProjectRoleMember ProjectRole = "member"
)

type JoinRequestStatus string

const (
	JoinRequestPending  JoinRequestStatus = "pending"
	JoinRequestApproved JoinRequestStatus = "approved"
	JoinRequestRejected JoinRequestStatus = "rejected"
)

type IncidentPriority string

const (
	IncidentPriorityLow    IncidentPriority = "low"
	IncidentPriorityMedium IncidentPriority = "medium"
	IncidentPriorityHigh   IncidentPriority = "high"
	IncidentPriorityCritical IncidentPriority = "critical"
)
type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "open"
	IncidentStatusInProgress IncidentStatus = "in_progress"
	IncidentStatusResolved   IncidentStatus = "submitted"
	IncidentStatusClosed     IncidentStatus = "finished"
)
