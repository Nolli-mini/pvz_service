package repository

import (
	"context"
	"pvz-service/internal/api"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// User
func (r *Repository) CreateUser(ctx context.Context, user *api.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*api.User, error) {
	var user api.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*api.User, error) {
	var user api.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// PVZ
func (r *Repository) CreatePVZ(ctx context.Context, pvz *api.PVZ) error {
	return r.db.WithContext(ctx).Create(pvz).Error
}

func (r *Repository) GetPVZByID(ctx context.Context, id uuid.UUID) (*api.PVZ, error) {
	var pvz api.PVZ
	err := r.db.WithContext(ctx).First(&pvz, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &pvz, err
}

func (r *Repository) GetPVZList(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]api.PVZ, int64, error) {
	var pvzs []api.PVZ
	var total int64

	query := r.db.WithContext(ctx).Model(&api.PVZ{})

	if startDate != nil || endDate != nil {
		subQuery := r.db.Table("receptions").
			Select("DISTINCT pvz_id").
			Where("1=1")

		if startDate != nil {
			subQuery = subQuery.Where("date_time >= ?", startDate)
		}
		if endDate != nil {
			subQuery = subQuery.Where("date_time <= ?", endDate)
		}

		query = query.Where("id IN (?)", subQuery)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("registration_date DESC").Find(&pvzs).Error

	return pvzs, total, err
}

func (r *Repository) GetPVZWithDetails(ctx context.Context, pvzIDs []uuid.UUID) ([]api.PVZ, error) {
	var pvzs []api.PVZ
	err := r.db.WithContext(ctx).
		Preload("Receptions", func(db *gorm.DB) *gorm.DB {
			return db.Order("date_time DESC")
		}).
		Preload("Receptions.Products", func(db *gorm.DB) *gorm.DB {
			return db.Order("date_time ASC")
		}).
		Where("id IN ?", pvzIDs).
		Find(&pvzs).Error
	return pvzs, err
}

// Reception
func (r *Repository) CreateReception(ctx context.Context, reception *api.Reception) error {
	return r.db.WithContext(ctx).Create(reception).Error
}

func (r *Repository) GetLastInProgressReception(ctx context.Context, pvzID uuid.UUID) (*api.Reception, error) {
	var reception api.Reception
	err := r.db.WithContext(ctx).
		Where("pvz_id = ? AND status = ?", pvzID, api.InProgress).
		Order("date_time DESC").
		First(&reception).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &reception, err
}

func (r *Repository) CloseReception(ctx context.Context, receptionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&api.Reception{}).
		Where("id = ?", receptionID).
		Update("status", api.Close).Error
}

// Product
func (r *Repository) CreateProduct(ctx context.Context, product *api.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *Repository) GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*api.Product, error) {
	var product api.Product
	err := r.db.WithContext(ctx).
		Where("reception_id = ?", receptionID).
		Order("date_time DESC").
		First(&product).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &product, err
}

func (r *Repository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&api.Product{}, "id = ?", productID).Error
}
