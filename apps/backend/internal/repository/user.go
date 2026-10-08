package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"incident-flow/backend/internal/logger"
	"incident-flow/backend/internal/models"
)

var ErrEmailExists = errors.New("email already registered")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, hashedPassword string, firstName, lastName string) (*models.User, error) {
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
		logger.Debug("repository: email already exists", "email", email)
		return nil, ErrEmailExists
	}
	if err != sql.ErrNoRows {
		logger.Error("repository: failed to check email", "email", email, "error", err.Error())
		return nil, fmt.Errorf("check email: %w", err)
	}

	// Insert new user
	var userID int
	var createdAt time.Time
	err = r.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id, created_at`,
		email,
		hashedPassword,
		firstName,
		lastName,
	).Scan(&userID, &createdAt)
	if err != nil {
		logger.Error("repository: failed to insert user", "email", email, "error", err.Error())
		return nil, fmt.Errorf("insert user: %w", err)
	}

	logger.Debug("repository: user created", "user_id", userID, "email", email, "first_name", firstName, "last_name", lastName)
	return &models.User{
		ID:        userID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: createdAt,
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
		"SELECT id, email, first_name, last_name, password_hash, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("repository: user not found", "email", email)
			return nil, nil
		}
		logger.Error("repository: failed to get user", "email", email, "error", err.Error())
		return nil, fmt.Errorf("get user: %w", err)
	}

	logger.Debug("repository: user retrieved", "user_id", user.ID, "email", email)
	return &user, nil
}
