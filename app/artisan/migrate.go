package artisan

import (
	"fmt"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"

	"github.com/google/uuid"
)

var modelsToMigrate = []interface{}{
	&models.User{},
	&models.SubscriptionPlan{},
	&models.Package{},
	&models.Room{},
	&models.Song{},
	&models.Guest{},
	&models.PurchaseLog{},
	&models.SystemConfig{},
	&models.Transaction{},
	&models.CreditLog{},
	&models.Session{},
	&models.TVToken{},
}

func runMigrate(args []string) error {
	// Check for "fresh" argument
	isFresh := len(args) > 0 && args[0] == "fresh"

	if isFresh {
		fmt.Println("Running fresh migration (dropping all tables)...")
	} else {
		fmt.Println("Running database migrations...")
	}

	if err := initializers.DbConnection(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Drop all tables if fresh migration
	if isFresh {
		fmt.Println("Dropping all tables...")
		for i := len(modelsToMigrate) - 1; i >= 0; i-- {
			if err := initializers.Db.Migrator().DropTable(modelsToMigrate[i]); err != nil {
				fmt.Printf("Warning: failed to drop table for %T: %v\n", modelsToMigrate[i], err)
			}
		}
		fmt.Println("All tables dropped successfully")
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
		{ID: uuid.New().String(), Key: models.ConfigRoomMaxDuration, Value: "120"}, // 2 hours default
		{ID: uuid.New().String(), Key: models.ConfigRoomCreationCost, Value: "1"},  // 1 credit to create room
		{ID: uuid.New().String(), Key: models.ConfigDefaultCredits, Value: "5"},    // 5 credits for new users
		{ID: uuid.New().String(), Key: models.ConfigDailyFreeCredits, Value: "5"},  // 5 daily free credits
	}

	for _, config := range defaultConfigs {
		var existing models.SystemConfig
		if initializers.Db.Where("key = ?", config.Key).First(&existing).RowsAffected == 0 {
			initializers.Db.Create(&config)
			fmt.Printf("Created default config: %s = %s\n", config.Key, config.Value)
		}
	}

	// Seed dummy package: 10 credits for IDR 0
	var dummyPackage models.Package
	if initializers.Db.Where("package_name = ?", "Starter Pack").First(&dummyPackage).RowsAffected == 0 {
		dummyPackage = models.Package{
			ID:            uuid.New().String(),
			PackageName:   "Starter Pack",
			PackageDetail: []byte("Get 10 free credits to start your karaoke journey!"),
			CreditAmount:  10,
			Price:         0,
			Visibility:    true,
		}
		initializers.Db.Create(&dummyPackage)
		fmt.Println("Created dummy package: Starter Pack (10 credits for IDR 0)")
	}
}
