package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Check if it's an APIError
			if apiErr, ok := err.(*APIError); ok {
				c.JSON(apiErr.Code, gin.H{
					"error": apiErr.Message,
				})
				return
			}

			// Default error response
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server error",
			})
		}
	}
}
