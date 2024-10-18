package helpers

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TruncateTables truncates the specified tables and resets their IDs.
// It also ensures that the tables are properly emptied before the tests proceed.
func TruncateTables(db *gorm.DB, tables ...string) error {
	for _, table := range tables {
		log.WithField("table", table).Info("Truncating table")

		// Execute the truncation query
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}

		// Verify that the table is empty
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to count rows in table %s: %w", table, err)
		}
		if count != 0 {
			return fmt.Errorf("table %s was not properly truncated, %d rows still present", table, count)
		}
	}
	return nil
}
