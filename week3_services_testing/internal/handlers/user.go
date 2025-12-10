package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"week3_services_testing/api/internal/models"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users  map[int]models.User
	nextID int
	mu     sync.RWMutex
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		users:  make(map[int]models.User),
		nextID: 1,
	}
}

// GetAllUsers handles GET /users with pagination filter and sorting
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get paginated params
	var pagination models.PaginationParams
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid pagination parameters",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if pagination.Page == 0 {
		pagination.Page = 1
	}
	if pagination.PageSize == 0 {
		pagination.PageSize = 10
	}

	// Get filter paramters
	var filter models.UserFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convert map to slice for sorting/pagination
	usersList := make([]models.User, 0, len(h.users))
	for _, user := range h.users {
		// Apply search filter if provided
		if filter.Search != "" {
			if !contains(user.Name, filter.Search) && !contains(user.Email, filter.Search) {
				continue
			}
		}

		if filter.MinAge > 0 && user.Age < filter.MinAge {
			continue
		}

		if filter.MaxAge > 0 && user.Age > filter.MaxAge {
			continue
		}
		usersList = append(usersList, user)
	}

	// Apply sorting
	if filter.SortBy != "" {
		sortUsers(usersList, filter.SortBy, filter.Order)
	}

	totalItems := len(usersList)

	// Apply pagination
	offset := pagination.GetOffset()
	end := offset + pagination.PageSize

	if offset > totalItems {
		offset = totalItems
	}
	if end > totalItems {
		end = totalItems
	}

	paginatedUsers := usersList[offset:end]

	// create paginated response
	paginatedResponse := models.NewPaginatedResponse(
		paginatedUsers,
		pagination.Page,
		pagination.PageSize,
		totalItems,
	)

	c.JSON(http.StatusOK, paginatedResponse)
}

func sortUsers(users []models.User, sortBy, order string) {
	sort.Slice(users, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = users[i].Name < users[j].Name
		case "email":
			less = users[i].Email < users[j].Email
		case "age":
			less = users[i].Age < users[j].Age
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

// helper function for search
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr || len(substr) == 0 || strContains(str, substr))
}

func strContains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	h.mu.RLock()
	user, exists := h.users[id]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User with ID:" + idStr + "not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User

	// Bind and validate JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// if err := Validateuser(user); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"error": err.Error(),
	// 	})
	// 	return
	// }

	h.mu.Lock()
	user.ID = h.nextID
	h.nextID++
	h.users[user.ID] = user
	h.mu.Unlock()

	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid id format",
		})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	_, exists := h.users[id]

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// if err := Validateuser(user); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"error": err.Error(),
	// 	})
	// 	return
	// }

	user.ID = id
	h.users[id] = user

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	_, exists := h.users[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User Not Found",
		})
		return
	}

	delete(h.users, id)

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// func Validateuser(user models.User) error {
// 	if user.Name == "" {
// 		return errors.New("name is required")
// 	}

// 	if user.Email == "" {
// 		return errors.New("name is required")
// 	}

// 	if user.Age < 0 || user.Age > 150 {
// 		return errors.New("name is required")
// 	}
// 	return nil
// }
