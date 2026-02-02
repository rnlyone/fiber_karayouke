package artisan

import (
	"fmt"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"
)

var modelsToMigrate = []interface{}{
	&models.User{},
	&models.Subscription{},
	&models.UserSubscription{},
	&models.Package{},
}

func runMigrate(_ []string) error {
	fmt.Println("Running database migrations...")
	if err := initializers.DbConnection(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := initializers.Db.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}

	fmt.Println("Migration completed successfully")
	return nil
}
