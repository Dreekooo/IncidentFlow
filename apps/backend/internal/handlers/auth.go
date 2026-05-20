package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process request",
		})
		return
	}

	// Create user in DB
	user, err := h.userRepo.CreateUser(c.Request.Context(), req.Email, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "email already registered",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}
