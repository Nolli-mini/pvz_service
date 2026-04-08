package tests

import (
	"context"
	"testing"

	"pvz-service/internal/api"
	"pvz-service/internal/logger"
	"pvz-service/internal/repository"
	"pvz-service/internal/service"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFullPVZWorkflow(t *testing.T) {

	logger.Init("test")

	// Настройка тестовой БД
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	db.AutoMigrate(&api.User{}, &api.PVZ{}, &api.Reception{}, &api.Product{})

	// Инициализация
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)

	ctx := context.Background()

	// 1. Создаем новый ПВЗ
	pvz, err := svc.CreatePVZ(ctx, "Москва")
	assert.NoError(t, err)
	assert.NotNil(t, pvz)
	pvzID := pvz.Id
	t.Logf("✓ PVZ created with ID: %s", pvzID)

	// 2. Добавляем новую приёмку заказов
	reception, err := svc.CreateReception(ctx, *pvzID)
	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, api.InProgress, reception.Status)
	t.Logf("✓ Reception created with ID: %s", reception.Id)

	// 3. Добавляем 50 товаров в рамках текущей приёмки
	productTypes := []api.ProductType{
		api.ProductTypeЭлектроника,
		api.ProductTypeОдежда,
		api.ProductTypeОбувь,
	}

	for i := 1; i <= 50; i++ {
		productType := productTypes[i%len(productTypes)]
		product, err := svc.AddProduct(ctx, *pvzID, productType)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, productType, product.Type)

		if i%10 == 0 {
			t.Logf("✓ Added %d products", i)
		}
	}
	t.Logf("✓ All 50 products added successfully")

	// 4. Закрываем приёмку заказов
	closedReception, err := svc.CloseReception(ctx, *pvzID)
	assert.NoError(t, err)
	assert.Equal(t, api.Close, closedReception.Status)
	t.Logf("✓ Reception closed successfully")

	// 5. Проверяем, что нельзя добавить товар в закрытую приёмку
	_, err = svc.AddProduct(ctx, *pvzID, api.ProductTypeЭлектроника)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no in-progress reception")
	t.Logf("✓ Cannot add product to closed reception (correct behavior)")

	// 6. Проверяем, что нельзя закрыть уже закрытую приёмку
	_, err = svc.CloseReception(ctx, *pvzID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no in-progress reception")
	t.Logf("✓ Cannot close already closed reception (correct behavior)")

	// 7. Проверяем количество созданных записей
	pvzs, total, err := svc.GetPVZList(ctx, nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, pvzs, 1)
	t.Logf("✓ PVZ count: %d", total)

	t.Logf("\n=== INTEGRATION TEST PASSED ===")
	t.Logf("Summary:")
	t.Logf("  - PVZ created: 1")
	t.Logf("  - Receptions created: 1")
	t.Logf("  - Products added: 50")
	t.Logf("  - Reception closed: yes")
}
