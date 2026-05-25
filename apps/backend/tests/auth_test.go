package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegisterSuccess(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())

	body := []byte(`{
	"email": "` + email + `",
	"password": "password123"
	}`)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestRegisterInvalidEmail(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)
	body := []byte(`{
	"email": "invalid-email",
	"password": "password123"
	}`)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
}

func TestRegisterDuplicateEmail(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)
	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())
	body := []byte(`{
	"email": "` + email + `",
	"password": "password123"
	}`)

	// First registration should succeed
	req1, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)

	req1.Header.Set("Content-Type", "application/json")

	recorder1 := httptest.NewRecorder()
	router.ServeHTTP(recorder1, req1)

	assert.Equal(t, http.StatusCreated, recorder1.Code)

	// Second registration with same email should fail
	req2, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)

	req2.Header.Set("Content-Type", "application/json")

	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)

	assert.Equal(t, http.StatusConflict, recorder2.Code)
}

func TestRegisterShortPassword(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())

	body := []byte(`{
	"email": "` + email + `",
	"password": "short"
	}`)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRegisterMissingPassword(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())

	body := []byte(`{
	"email": "` + email + `"
	}`)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRegisterPasswordHashing(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())

	body := []byte(`{
	"email": "` + email + `",
	"password": "password123"
	}`)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Verify password is hashed in the database
	user, err := userRepo.GetUserByEmail(req.Context(), email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEqual(t, "password123", user.Password)
}

func TestLoginSuccess(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())
	password := "password123"

	// First register the user
	body := []byte(`{
	"email": "` + email + `",
	"password": "` + password + `"
	}`)

	registerReq, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)
	registerReq.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, registerReq)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Now attempt to login
	loginReq, _ := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(body),
	)
	loginReq.Header.Set("Content-Type", "application/json")

	loginRecorder := httptest.NewRecorder()
	router.ServeHTTP(loginRecorder, loginReq)
	assert.Equal(t, http.StatusOK, loginRecorder.Code)
}

func TestLoginInvalidPassword(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())
	password := "password123"

	// First register the user
	body := []byte(`{
	"email": "` + email + `",
	"password": "` + password + `"
	}`)

	registerReq, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)
	registerReq.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, registerReq)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Now attempt to login with wrong password
	loginBody := []byte(`{
	"email": "` + email + `",
	"password": "wrongpassword"
	}`)

	loginReq, _ := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(loginBody),
	)
	loginReq.Header.Set("Content-Type", "application/json")

	loginRecorder := httptest.NewRecorder()
	router.ServeHTTP(loginRecorder, loginReq)
	assert.Equal(t, http.StatusUnauthorized, loginRecorder.Code)
}

func TestLoginUserNotFound(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())
	password := "password123"

	// Attempt to login without registering
	body := []byte(`{
	"email": "` + email + `",
	"password": "` + password + `"
	}`)

	loginReq, _ := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(body),
	)
	loginReq.Header.Set("Content-Type", "application/json")

	loginRecorder := httptest.NewRecorder()
	router.ServeHTTP(loginRecorder, loginReq)
	assert.Equal(t, http.StatusUnauthorized, loginRecorder.Code)
}

func TestLoginMissingPassword(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())

	// First register the user
	body := []byte(`{
	"email": "` + email + `",
	"password": "password123"
	}`)

	registerReq, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)
	registerReq.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, registerReq)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Now attempt to login with missing password
	loginBody := []byte(`{
	"email": "` + email + `"
	}`)

	loginReq, _ := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(loginBody),
	)
	loginReq.Header.Set("Content-Type", "application/json")

	loginRecorder := httptest.NewRecorder()
	router.ServeHTTP(loginRecorder, loginReq)
	assert.Equal(t, http.StatusBadRequest, loginRecorder.Code)
}

func TestLoginMissingEmail(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())
	password := "password123"

	// First register the user
	body := []byte(`{
	"email": "` + email + `",
	"password": "` + password + `"
	}`)

	registerReq, _ := http.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBuffer(body),
	)
	registerReq.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, registerReq)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Now attempt to login with missing email
	loginBody := []byte(`{
	"password": "` + password + `"
	}`)

	loginReq, _ := http.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBuffer(loginBody),
	)
	loginReq.Header.Set("Content-Type", "application/json")

	loginRecorder := httptest.NewRecorder()
	router.ServeHTTP(loginRecorder, loginReq)
	assert.Equal(t, http.StatusBadRequest, loginRecorder.Code)
}
