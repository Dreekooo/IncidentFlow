package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"incident-flow/backend/internal/models"
)

var ErrEmailExists = errors.New("email already registered")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, hashedPassword string) (*models.User, error) {
	// Ensure context has a timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
	}

	// Check if email already exists
	var existingID int
	err := r.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&existingID)
	if err == nil {
		return nil, ErrEmailExists
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("check email: %w", err)
	}

	// Insert new user
	var userID int
	var role string
	var createdAt time.Time
	err = r.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password, role, created_at, updated_at)
		 VALUES ($1, $2, 'viewer', NOW(), NOW())
		 RETURNING id, role, created_at`,
		email,
		hashedPassword,
	).Scan(&userID, &role, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &models.User{
		ID:        userID,
		Email:     email,
		Role:      role,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
	}

	var user models.User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, email, role, password, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Role, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}
