package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TransactionMiddleware is a middleware that wraps each request in a serializable database transaction.
func TransactionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Begin a GORM transaction
		tx := db.Begin()

		// Check for any errors in beginning the transaction
		if tx.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "could not start transaction",
			})
			return
		}

		// Attach the transaction to the context so it can be used in the handlers
		c.Set("db_tx", tx)

		// Proceed with the request
		c.Next()

		// After the request is completed, decide to commit or rollback the transaction
		if len(c.Errors) > 0 {
			// If there are any errors, rollback the transaction
			if err := tx.Rollback().Error; err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "failed to rollback transaction",
				})
				return
			}
		} else {
			// Otherwise, commit the transaction
			if err := tx.Commit().Error; err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "failed to commit transaction",
				})
				return
			}
		}
	}
}
