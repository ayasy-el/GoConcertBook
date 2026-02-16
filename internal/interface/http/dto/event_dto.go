package dto

type CreateEventRequest struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

type CreateCategoryRequest struct {
	Name       string `json:"name"`
	TotalStock int    `json:"total_stock"`
	Price      int64  `json:"price"`
}
