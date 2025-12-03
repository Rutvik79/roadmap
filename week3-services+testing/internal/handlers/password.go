package handlers

import (
	"net/http"
	"week3-services+testing/api/internal/auth"

	"github.com/gin-gonic/gin"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	email, _ := c.Get("email")

	h.mu.Lock()
	defer h.mu.Unlock()

	user, exists := h.users[email.(string)]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	if !auth.CheckPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid current password",
		})
		return
	}

	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	user.Password = hashedPassword
	h.users[email.(string)] = user

	c.JSON(200, gin.H{
		"message": "Password changed successfully",
	})
}
