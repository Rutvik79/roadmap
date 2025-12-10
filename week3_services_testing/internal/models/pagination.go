package models

type PaginationParams struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

func NewPaginationParams(page, pageSize int) *PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

func NewPaginatedResponse(data interface{}, page, pageSize, totalItems int) *PaginatedResponse {
	totalPages := (totalItems + pageSize - 1) / pageSize

	return &PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
