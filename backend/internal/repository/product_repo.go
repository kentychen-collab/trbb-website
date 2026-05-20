package repository

import (
	"database/sql"
	"encoding/json"
	"sports-platform/internal/model"
)

type ProductRepo struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) scanProduct(rows *sql.Rows) (model.Product, error) {
	var p model.Product
	var slug, description, coverImage, imagesJSON, categoryName sql.NullString
	var salePrice sql.NullFloat64
	err := rows.Scan(
		&p.ID, &p.CategoryID, &categoryName, &p.Name, &slug,
		&description, &coverImage, &imagesJSON,
		&p.Price, &salePrice, &p.Stock, &p.Status, &p.IsFeatured, &p.CreatedAt,
	)
	if err != nil {
		return p, err
	}
	p.Slug = slug.String
	p.Description = description.String
	p.CoverImage = coverImage.String
	p.CategoryName = categoryName.String
	if salePrice.Valid {
		p.SalePrice = &salePrice.Float64
	}
	if imagesJSON.Valid && imagesJSON.String != "" {
		json.Unmarshal([]byte(imagesJSON.String), &p.Images)
	}
	return p, nil
}

func (r *ProductRepo) List(limit, offset int, categoryID *int, status *int, featured bool) ([]model.Product, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if categoryID != nil {
		where += " AND p.category_id=?"
		args = append(args, *categoryID)
	}
	if status != nil {
		where += " AND p.status=?"
		args = append(args, *status)
	} else {
		where += " AND p.status=1"
	}
	if featured {
		where += " AND p.is_featured=1"
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	r.db.QueryRow("SELECT COUNT(*) FROM products p "+where, countArgs...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := r.db.Query(`
		SELECT p.id, p.category_id, c.name, p.name, p.slug,
		       p.description, p.cover_image, p.images,
		       p.price, p.sale_price, p.stock, p.status, p.is_featured, p.created_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		`+where+` ORDER BY p.created_at DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Product
	for rows.Next() {
		p, err := r.scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, nil
}

func (r *ProductRepo) FindByID(id uint64) (*model.Product, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.category_id, c.name, p.name, p.slug,
		       p.description, p.cover_image, p.images,
		       p.price, p.sale_price, p.stock, p.status, p.is_featured, p.created_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id=?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		p, err := r.scanProduct(rows)
		if err != nil {
			return nil, err
		}
		return &p, nil
	}
	return nil, nil
}

func (r *ProductRepo) FindBySlug(slug string) (*model.Product, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.category_id, c.name, p.name, p.slug,
		       p.description, p.cover_image, p.images,
		       p.price, p.sale_price, p.stock, p.status, p.is_featured, p.created_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.slug=? AND p.status=1`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		p, err := r.scanProduct(rows)
		if err != nil {
			return nil, err
		}
		return &p, nil
	}
	return nil, nil
}

func (r *ProductRepo) Create(p *model.Product) error {
	imagesJSON := "[]"
	if len(p.Images) > 0 {
		b, _ := json.Marshal(p.Images)
		imagesJSON = string(b)
	}
	result, err := r.db.Exec(`
		INSERT INTO products (category_id, name, slug, description, cover_image, images,
		                      price, sale_price, stock, status, is_featured)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.CategoryID, p.Name, p.Slug, p.Description, p.CoverImage, imagesJSON,
		p.Price, p.SalePrice, p.Stock, p.Status, p.IsFeatured)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	p.ID = uint64(id)
	return nil
}

func (r *ProductRepo) Update(p *model.Product) error {
	imagesJSON := "[]"
	if len(p.Images) > 0 {
		b, _ := json.Marshal(p.Images)
		imagesJSON = string(b)
	}
	_, err := r.db.Exec(`
		UPDATE products SET category_id=?, name=?, slug=?, description=?, cover_image=?,
		images=?, price=?, sale_price=?, stock=?, status=?, is_featured=?, updated_at=NOW()
		WHERE id=?`,
		p.CategoryID, p.Name, p.Slug, p.Description, p.CoverImage,
		imagesJSON, p.Price, p.SalePrice, p.Stock, p.Status, p.IsFeatured, p.ID)
	return err
}

func (r *ProductRepo) Delete(id uint64) error {
	_, err := r.db.Exec(`UPDATE products SET status=0, updated_at=NOW() WHERE id=?`, id)
	return err
}

func (r *ProductRepo) ListCategories() ([]model.Category, error) {
	rows, err := r.db.Query(`
		SELECT id, name, slug, sort, status FROM categories
		WHERE status=1 ORDER BY sort ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Category
	for rows.Next() {
		var c model.Category
		rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Sort, &c.Status)
		list = append(list, c)
	}
	return list, nil
}
