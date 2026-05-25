package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"incident-flow/backend/internal/auth"
	"incident-flow/backend/internal/logger"
	"incident-flow/backend/internal/repository"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

type RegisterResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type AuthHandler struct {
	userRepo *repository.UserRepository
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		logger.Warn("register: invalid request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: email must be valid, password must be 8-100 characters",
		})
		return
	}

	// Trim and lowercase email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("register: password hashing failed", "email", req.Email, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process request",
		})
		return
	}

	// Create user in DB
	user, err := h.userRepo.CreateUser(c.Request.Context(), req.Email, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			logger.Warn("register: email already exists", "email", req.Email)
			c.JSON(http.StatusConflict, gin.H{
				"error": "email already registered",
			})
			return
		}

		logger.Error("register: failed to create user", "email", req.Email, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	logger.Info("register: user created successfully", "user_id", user.ID, "email", req.Email)
	c.JSON(http.StatusCreated, RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		logger.Warn("login: invalid request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: email must be valid, password is required",
		})
		return
	}

	// Trim and lowercase email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Get user from DB
	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil || user == nil {
		logger.Warn("login: user not found or db error", "email", req.Email, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		logger.Warn("login: invalid password", "user_id", user.ID, "email", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	// Generate auth token
	token, err := auth.GenerateAuthToken(user.ID)
	if err != nil {
		logger.Error("login: token generation failed", "user_id", user.ID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate auth token",
		})
		return
	}

	logger.Info("login: successful", "user_id", user.ID, "email", req.Email)
	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
	})
}
