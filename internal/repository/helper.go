package repository

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Utility function to get transaction from context
func GetDB(c *gin.Context, db *gorm.DB) *gorm.DB {
	if c == nil {
		return db
	}
	tx, exists := c.Get("db_tx")
	if exists {
		return tx.(*gorm.DB)
	}
	return db // Fallback to main db if no transaction found (for non-transactional operations)
}
