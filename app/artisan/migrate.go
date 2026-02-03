package artisan

import (
	"fmt"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/google/uuid"
)

var modelsToMigrate = []interface{}{
	&models.User{},
	&models.Subscription{},
	&models.UserSubscription{},
	&models.Package{},
	&models.Room{},
	&models.Song{},
	&models.Guest{},
	&models.PurchaseLog{},
	&models.SystemConfig{},
	&models.Transaction{},
	&models.CreditLog{},
}

func runMigrate(_ []string) error {
	fmt.Println("Running database migrations...")
	if err := initializers.DbConnection(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := initializers.Db.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}

	// Seed default data
	seedDefaults()

	fmt.Println("Migration completed successfully")
	return nil
}

// seedDefaults creates default configuration and packages
func seedDefaults() {
	// Seed default system configs
	defaultConfigs := []models.SystemConfig{
		{Key: models.ConfigRoomMaxDuration, Value: "120"}, // 2 hours default
		{Key: models.ConfigRoomCreationCost, Value: "1"},  // 1 credit to create room
		{Key: models.ConfigDefaultCredits, Value: "5"},    // 5 credits for new users
	}

	for _, config := range defaultConfigs {
		var existing models.SystemConfig
		if initializers.Db.Where("key = ?", config.Key).First(&existing).RowsAffected == 0 {
			initializers.Db.Create(&config)
			fmt.Printf("Created default config: %s = %s\n", config.Key, config.Value)
		}
	}

	// Seed dummy package: 10 credits for 0 Rupiah
	var dummyPackage models.Package
	if initializers.Db.Where("package_name = ?", "Starter Pack").First(&dummyPackage).RowsAffected == 0 {
		dummyPackage = models.Package{
			ID:            uuid.New().String(),
			PackageName:   "Starter Pack",
			PackageDetail: []byte("Get 10 free credits to start your karaoke journey!"),
			CreditAmount:  10,
			Price:         "0",
			Visibility:    true,
		}
		initializers.Db.Create(&dummyPackage)
		fmt.Println("Created dummy package: Starter Pack (10 credits for Rp 0)")
	}
}
