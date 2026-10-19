// Command platform-operator provisions an operator through trusted database access.
// No public HTTP route can grant this authority.
package main

import (
	"flag"
	"fmt"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"os"
)

func main() {
	user := flag.String("user-id", "", "existing account user ID")
	revoke := flag.Bool("revoke", false, "remove platform operator access")
	flag.Parse()
	if *user == "" || os.Getenv("PLATFORM_DATABASE_URL") == "" {
		fmt.Fprintln(os.Stderr, "provide -user-id and PLATFORM_DATABASE_URL through trusted operator environment")
		os.Exit(1)
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("PLATFORM_DATABASE_URL")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fail()
	}
	var count int64
	if db.Model(&model.AuthIdentities{}).Where("user_id = ?", *user).Count(&count).Error != nil || count == 0 {
		fmt.Fprintln(os.Stderr, "account does not exist")
		os.Exit(1)
	}
	if *revoke {
		err = db.Where("user_id = ?", *user).Delete(&model.PlatformOperator{}).Error
	} else {
		err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.PlatformOperator{UserID: *user}).Error
	}
	if err != nil {
		fail()
	}
	fmt.Println("Platform operator grant updated.")
}
func fail() {
	fmt.Fprintln(os.Stderr, "operator provisioning failed; verify the database and applied migrations")
	os.Exit(1)
}
