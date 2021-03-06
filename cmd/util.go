package cmd

import (
	"github.com/RichardKnop/example-api/config"
	"github.com/RichardKnop/example-api/database"
	"github.com/jinzhu/gorm"
)

// initConfigDB loads the configuration and connects to the database
func initConfigDB(mustLoadOnce, keepReloading bool) (*config.Config, *gorm.DB, error) {
	// Config
	cnf := config.NewConfig(mustLoadOnce, keepReloading)

	// Database
	db, err := database.NewDatabase(cnf)
	if err != nil {
		return nil, nil, err
	}

	return cnf, db, nil
}
