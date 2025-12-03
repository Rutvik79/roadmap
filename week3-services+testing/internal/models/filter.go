package models

type UserFilter struct {
	MinAge int    `form:"min_age" binding:"omitempty,min=0"`
	MaxAge int    `form:"max_age" binding:"omitempty,max=130"`
	Search string `form:"search"`
	SortBy string `form:"sort_by" binding:"omitempty,oneof=name email age"`
	Order  string `form:"order" binding:"omitempty,oneof=asc desc"`
}
