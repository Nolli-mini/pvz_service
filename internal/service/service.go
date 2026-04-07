package service

import (
	"context"
	"errors"
	"pvz-service/internal/api"
	"pvz-service/internal/logger"
	"pvz-service/internal/metrics"
	"pvz-service/internal/repository"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, email, password string, role api.UserRole) (*api.User, error) {
	logger.Log.WithFields(map[string]interface{}{
		"email": email,
		"role":  role,
	}).Debug("Registering user")

	existing, _ := s.repo.GetUserByEmail(ctx, email)
	if existing != nil {
		logger.Log.WithField("email", email).Warn("User already exists")
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to hash password")
		return nil, errors.New("failed to hash password")
	}

	id := uuid.New()
	user := &api.User{
		Id:       &id,
		Email:    openapi_types.Email(email),
		Password: string(hashedPassword),
		Role:     role,
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to create user")
		return nil, err
	}

	logger.Log.WithFields(map[string]interface{}{
		"user_id": user.Id,
		"email":   email,
		"role":    role,
	}).Info("User registered successfully")

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*api.User, error) {
	logger.Log.WithField("email", email).Debug("User login attempt")

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		logger.Log.WithField("email", email).Warn("Login failed: user not found")
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logger.Log.WithField("email", email).Warn("Login failed: invalid password")
		return nil, errors.New("invalid credentials")
	}

	logger.Log.WithFields(map[string]interface{}{
		"user_id": user.Id,
		"email":   email,
	}).Info("User logged in successfully")

	return user, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*api.User, error) {
	logger.Log.WithField("user_id", id).Debug("Getting user by ID")

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get user")
		return nil, err
	}

	if user == nil {
		logger.Log.WithField("user_id", id).Warn("User not found")
	}

	return user, err
}

// PVZ business logic
func (s *Service) CreatePVZ(ctx context.Context, city api.PVZCity) (*api.PVZ, error) {
	logger.Log.WithField("city", city).Debug("Creating PVZ")

	validCities := map[api.PVZCity]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}
	if !validCities[city] {
		logger.Log.WithField("city", city).Warn("Invalid city for PVZ")
		return nil, errors.New("pvz can only be created in Moscow, Saint Petersburg, or Kazan")
	}

	now := time.Now()
	id := uuid.New()
	pvz := &api.PVZ{
		Id:               &id,
		RegistrationDate: &now,
		City:             city,
	}

	err := s.repo.CreatePVZ(ctx, pvz)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to create PVZ")
		return nil, err
	}

	metrics.IncPVZsCreated()

	logger.Log.WithFields(map[string]interface{}{
		"pvz_id": pvz.Id,
		"city":   city,
	}).Info("PVZ created successfully")

	return pvz, nil
}

func (s *Service) GetPVZList(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]api.PVZ, int64, error) {
	logger.Log.WithFields(map[string]interface{}{
		"page":  page,
		"limit": limit,
	}).Debug("Getting PVZ list")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 30 {
		limit = 30
	}

	pvzs, total, err := s.repo.GetPVZList(ctx, startDate, endDate, page, limit)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get PVZ list")
		return nil, 0, err
	}

	logger.Log.WithFields(map[string]interface{}{
		"count": len(pvzs),
		"total": total,
	}).Debug("PVZ list retrieved successfully")

	return pvzs, total, nil
}

func (s *Service) GetPVZWithDetails(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]api.PVZ, error) {
	logger.Log.Debug("Getting PVZ with details")

	pvzs, _, err := s.repo.GetPVZList(ctx, startDate, endDate, page, limit)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get PVZ list")
		return nil, err
	}

	if len(pvzs) == 0 {
		logger.Log.Debug("No PVZs found")
		return []api.PVZ{}, nil
	}

	ids := make([]uuid.UUID, len(pvzs))
	for i, pvz := range pvzs {
		ids[i] = *pvz.Id
	}

	result, err := s.repo.GetPVZWithDetails(ctx, ids)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get PVZ with details")
		return nil, err
	}

	logger.Log.WithField("count", len(result)).Debug("PVZ with details retrieved successfully")
	return result, nil
}

// Reception business logic
func (s *Service) CreateReception(ctx context.Context, pvzID uuid.UUID) (*api.Reception, error) {
	logger.Log.WithField("pvz_id", pvzID).Debug("Creating reception")

	pvz, _ := s.repo.GetPVZByID(ctx, pvzID)
	if pvz == nil {
		logger.Log.WithField("pvz_id", pvzID).Warn("PVZ not found")
		return nil, errors.New("pvz not found")
	}

	existing, _ := s.repo.GetLastInProgressReception(ctx, pvzID)
	if existing != nil {
		logger.Log.WithField("pvz_id", pvzID).Warn("Already has in-progress reception")
		return nil, errors.New("there is already an in-progress reception")
	}

	id := uuid.New()
	reception := &api.Reception{
		Id:       &id,
		DateTime: time.Now(),
		PvzId:    pvzID,
		Status:   api.InProgress,
	}

	err := s.repo.CreateReception(ctx, reception)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to create reception")
		return nil, err
	}

	metrics.IncReceptionsCreated()

	logger.Log.WithFields(map[string]interface{}{
		"reception_id": reception.Id,
		"pvz_id":       pvzID,
	}).Info("Reception created successfully")

	return reception, nil
}

func (s *Service) CloseReception(ctx context.Context, pvzID uuid.UUID) (*api.Reception, error) {
	logger.Log.WithField("pvz_id", pvzID).Debug("Closing reception")

	reception, err := s.repo.GetLastInProgressReception(ctx, pvzID)
	if err != nil || reception == nil {
		logger.Log.WithField("pvz_id", pvzID).Warn("No in-progress reception found")
		return nil, errors.New("no in-progress reception found")
	}

	err = s.repo.CloseReception(ctx, *reception.Id)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to close reception")
		return nil, err
	}

	reception.Status = api.Close

	logger.Log.WithFields(map[string]interface{}{
		"reception_id": reception.Id,
		"pvz_id":       pvzID,
	}).Info("Reception closed successfully")

	return reception, nil
}

// Product business logic
func (s *Service) AddProduct(ctx context.Context, pvzID uuid.UUID, productType api.ProductType) (*api.Product, error) {
	logger.Log.WithFields(map[string]interface{}{
		"pvz_id": pvzID,
		"type":   productType,
	}).Debug("Adding product")

	reception, err := s.repo.GetLastInProgressReception(ctx, pvzID)
	if err != nil || reception == nil {
		logger.Log.WithField("pvz_id", pvzID).Warn("No active reception found")
		return nil, errors.New("no in-progress reception found")
	}

	now := time.Now()
	id := uuid.New()
	product := &api.Product{
		Id:          &id,
		DateTime:    &now,
		Type:        productType,
		ReceptionId: *reception.Id,
	}

	err = s.repo.CreateProduct(ctx, product)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to create product")
		return nil, err
	}

	metrics.IncProductsAdded()

	logger.Log.WithFields(map[string]interface{}{
		"product_id":   product.Id,
		"reception_id": reception.Id,
		"type":         productType,
	}).Info("Product added successfully")

	return product, nil
}

func (s *Service) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	logger.Log.WithField("pvz_id", pvzID).Debug("Deleting last product")

	reception, err := s.repo.GetLastInProgressReception(ctx, pvzID)
	if err != nil || reception == nil {
		logger.Log.WithField("pvz_id", pvzID).Warn("No active reception found")
		return errors.New("no in-progress reception found")
	}

	product, err := s.repo.GetLastProduct(ctx, *reception.Id)
	if err != nil || product == nil {
		logger.Log.WithField("reception_id", reception.Id).Warn("No products to delete")
		return errors.New("no products to delete")
	}

	err = s.repo.DeleteProduct(ctx, *product.Id)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to delete product")
		return err
	}

	logger.Log.WithFields(map[string]interface{}{
		"product_id":   product.Id,
		"reception_id": reception.Id,
	}).Info("Product deleted successfully")

	return nil
}
