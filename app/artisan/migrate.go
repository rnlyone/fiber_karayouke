package artisan

import (
	"fmt"

	"GoFiberMVC/app/initializers"
	"GoFiberMVC/app/models"
)

var modelsToMigrate = []interface{}{
	&models.Account{},
	&models.Client{},
	&models.ClientAccount{},
	&models.ClientDocument{},
	&models.Events{},
	&models.Service{},
	&models.ServiceProvider{},
	&models.TransactionLog{},
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
