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
