package database

import (
	"log"
	"os"
	"pvz-service/internal/api"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=db user=user password=password dbname=myapp port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	return db
}

func MigrateDB(db *gorm.DB) error {

	if err := db.AutoMigrate(
		&api.User{},
		&api.PVZ{},
		&api.Reception{},
		&api.Product{},
	); err != nil {
		return err
	}

	if err := createIndexes(db); err != nil {
		log.Println("Warning: failed to create indexes:", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

func createIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_receptions_pvz_id ON receptions(pvz_id)",
		"CREATE INDEX IF NOT EXISTS idx_receptions_status ON receptions(status)",
		"CREATE INDEX IF NOT EXISTS idx_receptions_date_time ON receptions(date_time)",
		"CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products(reception_id)",
		"CREATE INDEX IF NOT EXISTS idx_products_date_time ON products(date_time)",
		"CREATE INDEX IF NOT EXISTS idx_pvz_registration_date ON pvz(registration_date)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			return err
		}
	}

	return nil
}

func SetupDB() *gorm.DB {
	db := InitDB()

	if err := MigrateDB(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	return db
}
