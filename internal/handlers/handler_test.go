package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"pvz-service/internal/api"
	"pvz-service/internal/logger"
	"pvz-service/internal/repository"
	"pvz-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*service.Service, *Handler, *gin.Engine) {
	logger.Init("test")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	db.AutoMigrate(&api.User{}, &api.PVZ{}, &api.Reception{}, &api.Product{})

	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	handler := NewHandler(svc)

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	return svc, handler, router
}

func TestHandler_DummyLogin(t *testing.T) {
	_, handler, router := setupTestDB(t)
	router.POST("/dummyLogin", handler.DummyLogin)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "employee role",
			body:       `{"role":"employee"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "moderator role",
			body:       `{"role":"moderator"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid role",
			body:       `{"role":"invalid"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestHandler_Register(t *testing.T) {
	_, handler, router := setupTestDB(t)
	router.POST("/register", handler.Register)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "successful registration",
			body:       `{"email":"test@test.com","password":"123","role":"employee"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid role",
			body:       `{"email":"test@test.com","password":"123","role":"admin"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	svc, handler, router := setupTestDB(t)

	// Регистрируем пользователя
	_, err := svc.Register(context.Background(), "logintest@test.com", "123456", api.UserRoleEmployee)
	assert.NoError(t, err)

	router.POST("/login", handler.Login)

	// Успешный логин
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"logintest@test.com","password":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())

	// Неправильный пароль
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"logintest@test.com","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
