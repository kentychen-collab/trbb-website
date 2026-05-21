package service

import (
	"errors"
	"fmt"
	"sports-platform/internal/model"
	"sports-platform/internal/repository"
	"strings"
	"time"
	"unicode"
)

type ProductService struct {
	productRepo *repository.ProductRepo
}

func NewProductService(productRepo *repository.ProductRepo) *ProductService {
	return &ProductService{productRepo: productRepo}
}

type ProductInput struct {
	CategoryID  uint32   `json:"category_id" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	CoverImage  string   `json:"cover_image"`
	Images      []string `json:"images"`
	Price       float64  `json:"price" binding:"required,gt=0"`
	SalePrice   *float64 `json:"sale_price"`
	Stock       int      `json:"stock"`
	Status      int      `json:"status"`
	IsFeatured  int      `json:"is_featured"`
}

func toSlug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		slug = fmt.Sprintf("product-%d", time.Now().Unix())
	}
	return slug
}

func (s *ProductService) List(page, limit int, categoryID *int, featuredOnly bool) ([]model.Product, int, error) {
	status := 1
	return s.productRepo.List(limit, (page-1)*limit, categoryID, &status, featuredOnly)
}

func (s *ProductService) AdminList(page, limit int, categoryID *int, status *int) ([]model.Product, int, error) {
	return s.productRepo.List(limit, (page-1)*limit, categoryID, status, false)
}

func (s *ProductService) GetBySlug(slug string) (*model.Product, error) {
	return s.productRepo.FindBySlug(slug)
}

func (s *ProductService) GetByID(id uint64) (*model.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *ProductService) Create(input ProductInput) (*model.Product, error) {
	if input.Slug == "" {
		input.Slug = toSlug(input.Name)
	}
	if input.SalePrice != nil && *input.SalePrice >= input.Price {
		return nil, errors.New("特價必須低於原價")
	}
	p := &model.Product{
		CategoryID:  input.CategoryID,
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		CoverImage:  input.CoverImage,
		Images:      input.Images,
		Price:       input.Price,
		SalePrice:   input.SalePrice,
		Stock:       input.Stock,
		Status:      input.Status,
		IsFeatured:  input.IsFeatured,
	}
	if err := s.productRepo.Create(p); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			p.Slug = fmt.Sprintf("%s-%d", input.Slug, time.Now().Unix())
			err = s.productRepo.Create(p)
		}
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (s *ProductService) Update(id uint64, input ProductInput) (*model.Product, error) {
	p, err := s.productRepo.FindByID(id)
	if err != nil || p == nil {
		return nil, errors.New("商品不存在")
	}
	if input.Slug == "" {
		input.Slug = p.Slug
	}
	if input.SalePrice != nil && *input.SalePrice >= input.Price {
		return nil, errors.New("特價必須低於原價")
	}
	p.CategoryID  = input.CategoryID
	p.Name        = input.Name
	p.Slug        = input.Slug
	p.Description = input.Description
	p.CoverImage  = input.CoverImage
	p.Images      = input.Images
	p.Price       = input.Price
	p.SalePrice   = input.SalePrice
	p.Stock       = input.Stock
	p.Status      = input.Status
	p.IsFeatured  = input.IsFeatured
	if err := s.productRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) SetStatus(id uint64, status int) error {
	p, err := s.productRepo.FindByID(id)
	if err != nil || p == nil {
		return errors.New("商品不存在")
	}
	p.Status = status
	return s.productRepo.Update(p)
}

func (s *ProductService) ListCategories() ([]model.Category, error) {
	return s.productRepo.ListCategories()
}
