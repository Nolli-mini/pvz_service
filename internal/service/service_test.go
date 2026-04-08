package service

import (
	"context"
	"testing"

	"pvz-service/internal/api"
	"pvz-service/internal/logger"
	"pvz-service/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	logger.Init("test")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	db.AutoMigrate(&api.User{}, &api.PVZ{}, &api.Reception{}, &api.Product{})
	return db
}

func TestService_Register(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	user, err := svc.Register(context.Background(), "test@test.com", "123456", api.UserRoleEmployee)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@test.com", string(user.Email))

	_, err = svc.Register(context.Background(), "test@test.com", "123456", api.UserRoleEmployee)
	assert.Error(t, err)
}

func TestService_Login(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	_, err := svc.Register(context.Background(), "login@test.com", "123456", api.UserRoleEmployee)
	assert.NoError(t, err)

	user, err := svc.Login(context.Background(), "login@test.com", "123456")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	_, err = svc.Login(context.Background(), "login@test.com", "wrong")
	assert.Error(t, err)
}

func TestService_CreatePVZ(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	pvz, err := svc.CreatePVZ(context.Background(), "Москва")
	assert.NoError(t, err)
	assert.NotNil(t, pvz)
	assert.Equal(t, api.PVZCity("Москва"), pvz.City)

	_, err = svc.CreatePVZ(context.Background(), "Новосибирск")
	assert.Error(t, err)
}

func TestService_CreateReception(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	pvz, _ := svc.CreatePVZ(context.Background(), "Москва")

	reception, err := svc.CreateReception(context.Background(), *pvz.Id)
	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, api.InProgress, reception.Status)

	_, err = svc.CreateReception(context.Background(), *pvz.Id)
	assert.Error(t, err)
}

func TestService_AddProduct(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	pvz, _ := svc.CreatePVZ(context.Background(), "Москва")
	_, _ = svc.CreateReception(context.Background(), *pvz.Id)

	product, err := svc.AddProduct(context.Background(), *pvz.Id, api.ProductTypeЭлектроника)
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, api.ProductTypeЭлектроника, product.Type)
}

func TestService_DeleteLastProduct(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	pvz, _ := svc.CreatePVZ(context.Background(), "Москва")
	_, _ = svc.CreateReception(context.Background(), *pvz.Id)

	_, _ = svc.AddProduct(context.Background(), *pvz.Id, api.ProductTypeЭлектроника)
	_, _ = svc.AddProduct(context.Background(), *pvz.Id, api.ProductTypeОдежда)

	err := svc.DeleteLastProduct(context.Background(), *pvz.Id)
	assert.NoError(t, err)

	err = svc.DeleteLastProduct(context.Background(), *pvz.Id)
	assert.NoError(t, err)

	err = svc.DeleteLastProduct(context.Background(), *pvz.Id)
	assert.Error(t, err)
}

func TestService_CloseReception(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	pvz, _ := svc.CreatePVZ(context.Background(), "Москва")
	_, _ = svc.CreateReception(context.Background(), *pvz.Id)
	_, _ = svc.AddProduct(context.Background(), *pvz.Id, api.ProductTypeЭлектроника)

	closed, err := svc.CloseReception(context.Background(), *pvz.Id)
	assert.NoError(t, err)
	assert.Equal(t, api.Close, closed.Status)

	_, err = svc.AddProduct(context.Background(), *pvz.Id, api.ProductTypeОбувь)
	assert.Error(t, err)
}

func TestService_GetPVZList(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewRepository(db)
	svc := NewService(repo)

	svc.CreatePVZ(context.Background(), "Москва")
	svc.CreatePVZ(context.Background(), "Санкт-Петербург")

	pvzs, total, err := svc.GetPVZList(context.Background(), nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, pvzs, 2)
}
