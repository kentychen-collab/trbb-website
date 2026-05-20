package model

import "time"

type Product struct {
	ID          uint64    `json:"id"`
	CategoryID  uint32    `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CoverImage  string    `json:"cover_image"`
	Price       float64   `json:"price"`
	SalePrice   float64   `json:"sale_price"`
	Stock       int       `json:"stock"`
	Status      int       `json:"status"`
	IsFeatured  int       `json:"is_featured"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductVariant struct {
	ID        uint64  `json:"id"`
	ProductID uint64  `json:"product_id"`
	SKU       string  `json:"sku"`
	SpecName  string  `json:"spec_name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
}

type Category struct {
	ID       uint32 `json:"id"`
	ParentID uint32 `json:"parent_id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}
