package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func TransactionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a new transaction
		tx := db.Begin()
		if tx.Error != nil {
			log.WithFields(log.Fields{
				"error": tx.Error,
			}).Error("Failed to start transaction")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		log.Info("Transaction started")
		// Store the transaction in the context
		c.Set("db_tx", tx)

		// Ensure rollback if panic occurs
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.WithField("panic", r).Error("Panic occurred, transaction rolled back")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unexpected server error"})
			}
		}()

		// Process the request
		c.Next()

		// Check if there are any errors during the request
		if len(c.Errors) > 0 {
			// Rollback the transaction if any errors occurred
			if err := tx.Rollback().Error; err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Failed to rollback transaction")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to rollback transaction"})
				return
			}
			log.Info("Transaction rolled back due to errors")
			return
		}

		// Commit the transaction if no errors occurred
		if err := tx.Commit().Error; err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to commit transaction")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		log.Info("Transaction committed successfully")
	}
}
