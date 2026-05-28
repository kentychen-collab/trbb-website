package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"trbb/internal/models"
	"trbb/pkg/database"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrOrderNotFound   = errors.New("order not found")
	ErrOutOfStock      = errors.New("product out of stock")
)

type ShopService struct {
	db *database.DB
}

func NewShopService(db *database.DB) *ShopService {
	return &ShopService{db: db}
}

// ── List products (public) ────────────────────────────────────

type ListProductsInput struct {
	Status   *int   `form:"status"`
	Category *int   `form:"category"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ListProductsResult struct {
	Products []*models.Product `json:"products"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	Pages    int               `json:"pages"`
}

func (s *ShopService) ListPublic(ctx context.Context, in ListProductsInput) (*ListProductsResult, error) {
	published := models.ProductPublished
	in.Status = &published
	return s.list(ctx, in)
}

func (s *ShopService) ListAdmin(ctx context.Context, in ListProductsInput) (*ListProductsResult, error) {
	return s.list(ctx, in)
}

func (s *ShopService) list(ctx context.Context, in ListProductsInput) (*ListProductsResult, error) {
	if in.Page < 1 { in.Page = 1 }
	if in.PageSize < 1 || in.PageSize > 100 { in.PageSize = 20 }

	where := "WHERE 1=1"
	args := []any{}
	if in.Status != nil { where += " AND status=?"; args = append(args, *in.Status) }
	if in.Category != nil { where += " AND category=?"; args = append(args, *in.Category) }
	if in.Keyword != "" {
		where += " AND (title LIKE ? OR description LIKE ?)"
		kw := "%" + in.Keyword + "%"
		args = append(args, kw, kw)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count products: %w", err)
	}

	offset := (in.Page - 1) * in.PageSize
	rows, err := s.db.QueryContext(ctx, `
		SELECT id,uuid,title,COALESCE(description,''),category,price,stock,
		       COALESCE(images,'[]'),COALESCE(specs,'[]'),status,creator_id,created_at,updated_at
		FROM products `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		append(args, in.PageSize, offset)...,
	)
	if err != nil { return nil, fmt.Errorf("list products: %w", err) }
	defer rows.Close()

	products := []*models.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil { return nil, err }
		products = append(products, p)
	}
	pages := (total + in.PageSize - 1) / in.PageSize
	if pages < 1 { pages = 1 }
	return &ListProductsResult{Products: products, Total: total, Page: in.Page, Pages: pages}, nil
}

func (s *ShopService) GetProductByID(ctx context.Context, id uint64) (*models.Product, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id,uuid,title,COALESCE(description,''),category,price,stock,
		       COALESCE(images,'[]'),COALESCE(specs,'[]'),status,creator_id,created_at,updated_at
		FROM products WHERE id=?`, id)
	if err != nil { return nil, err }
	defer rows.Close()
	if !rows.Next() { return nil, ErrProductNotFound }
	return scanProduct(rows)
}

func scanProduct(rows *sql.Rows) (*models.Product, error) {
	p := &models.Product{}
	var imagesJSON, specsJSON string
	if err := rows.Scan(
		&p.ID, &p.UUID, &p.Title, &p.Description, &p.Category,
		&p.Price, &p.Stock, &imagesJSON, &specsJSON,
		&p.Status, &p.CreatorID, &p.CreatedAt, &p.UpdatedAt,
	); err != nil { return nil, fmt.Errorf("scan product: %w", err) }
	_ = json.Unmarshal([]byte(imagesJSON), &p.Images)
	_ = json.Unmarshal([]byte(specsJSON), &p.Specs)
	return p, nil
}

// ── Create / Update / Delete product ─────────────────────────

type ProductInput struct {
	Title       string   `json:"title"       binding:"required"`
	Description string   `json:"description"`
	Category    int      `json:"category"`
	Price       float64  `json:"price"       binding:"required"`
	Stock       int      `json:"stock"`
	Images      []string `json:"images"`
	Specs       []string `json:"specs"`
	Status      int      `json:"status"`
}

func (s *ShopService) CreateProduct(ctx context.Context, in ProductInput, creatorID uint64) (*models.Product, error) {
	uid := uuid.New().String()
	now := time.Now()
	imagesJSON, _ := json.Marshal(in.Images)
	specsJSON, _ := json.Marshal(in.Specs)

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO products (uuid,title,description,category,price,stock,images,specs,status,creator_id,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		uid, in.Title, nullStr(in.Description), in.Category, in.Price, in.Stock,
		string(imagesJSON), string(specsJSON), in.Status, creatorID, now, now,
	)
	if err != nil { return nil, fmt.Errorf("create product: %w", err) }
	id, _ := res.LastInsertId()
	return s.GetProductByID(ctx, uint64(id))
}

func (s *ShopService) UpdateProduct(ctx context.Context, id uint64, in ProductInput) (*models.Product, error) {
	imagesJSON, _ := json.Marshal(in.Images)
	specsJSON, _ := json.Marshal(in.Specs)
	_, err := s.db.ExecContext(ctx, `
		UPDATE products SET title=?,description=?,category=?,price=?,stock=?,images=?,specs=?,status=?,updated_at=?
		WHERE id=?`,
		in.Title, nullStr(in.Description), in.Category, in.Price, in.Stock,
		string(imagesJSON), string(specsJSON), in.Status, time.Now(), id,
	)
	if err != nil { return nil, fmt.Errorf("update product: %w", err) }
	return s.GetProductByID(ctx, id)
}

func (s *ShopService) DeleteProduct(ctx context.Context, id uint64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM products WHERE id=?", id)
	return err
}

// ── Create order ──────────────────────────────────────────────

type OrderInput struct {
	Items []struct {
		ProductID uint64 `json:"product_id" binding:"required"`
		Qty       int    `json:"qty"        binding:"required,min=1"`
		Spec      string `json:"spec"`
	} `json:"items" binding:"required,min=1"`
	ShippingName   string `json:"shipping_name"   binding:"required"`
	ShippingPhone  string `json:"shipping_phone"  binding:"required"`
	ShippingAddr   string `json:"shipping_addr"`
	Note           string `json:"note"`
	PaymentMethod  int    `json:"payment_method"  binding:"required"`
	DeliveryMethod int    `json:"delivery_method" binding:"required"`
	PickupLocation string `json:"pickup_location"`
}

func (s *ShopService) CreateOrder(ctx context.Context, userID uint64, in OrderInput) (*models.Order, error) {
	// Validate items & compute total
	type itemRow struct {
		product *models.Product
		qty     int
		spec    string
	}
	var rows []itemRow
	var total float64

	for _, it := range in.Items {
		p, err := s.GetProductByID(ctx, it.ProductID)
		if err != nil { return nil, fmt.Errorf("product %d: %w", it.ProductID, err) }
		if p.Status != models.ProductPublished { return nil, fmt.Errorf("product %d not available", it.ProductID) }
		if p.Stock < it.Qty { return nil, fmt.Errorf("%w: %s", ErrOutOfStock, p.Title) }
		rows = append(rows, itemRow{p, it.Qty, it.Spec})
		total += p.Price * float64(it.Qty)
	}

	// Begin tx
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil { return nil, err }
	defer tx.Rollback()

	uid := uuid.New().String()
	now := time.Now()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO orders
		  (uuid,user_id,order_type,total_amount,status,
		   shipping_name,shipping_phone,shipping_addr,note,
		   payment_method,delivery_method,pickup_location,payment_status,
		   created_at,updated_at)
		VALUES (?,?,1,?,0, ?,?,?,?, ?,?,?,0, ?,?)`,
		uid, userID, total,
		nullStr(in.ShippingName), nullStr(in.ShippingPhone), nullStr(in.ShippingAddr), nullStr(in.Note),
		in.PaymentMethod, in.DeliveryMethod, nullStr(in.PickupLocation),
		now, now,
	)
	if err != nil { return nil, fmt.Errorf("insert order: %w", err) }
	orderID, _ := res.LastInsertId()

	// Insert items & deduct stock
	for _, r := range rows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id,ref_type,ref_id,title,price,qty,spec)
			VALUES (?,?,?,?,?,?,?)`,
			orderID, "product", r.product.ID, r.product.Title, r.product.Price, r.qty, nullStr(r.spec),
		)
		if err != nil { return nil, fmt.Errorf("insert order item: %w", err) }
		_, err = tx.ExecContext(ctx,
			"UPDATE products SET stock=stock-? WHERE id=?", r.qty, r.product.ID)
		if err != nil { return nil, fmt.Errorf("deduct stock: %w", err) }
	}

	if err := tx.Commit(); err != nil { return nil, err }
	return s.GetOrderByID(ctx, uint64(orderID))
}

// ── Get / list orders ─────────────────────────────────────────

func (s *ShopService) GetOrderByID(ctx context.Context, id uint64) (*models.Order, error) {
	o, err := s.scanOrder(s.db.QueryRowContext(ctx, orderSelectSQL+"WHERE o.id=?", id))
	if err != nil { return nil, err }
	o.Items, _ = s.getOrderItems(ctx, o.ID)
	return o, nil
}

func (s *ShopService) GetOrderByIDAndUser(ctx context.Context, id, userID uint64) (*models.Order, error) {
	o, err := s.scanOrder(s.db.QueryRowContext(ctx, orderSelectSQL+"WHERE o.id=? AND o.user_id=?", id, userID))
	if err != nil { return nil, err }
	o.Items, _ = s.getOrderItems(ctx, o.ID)
	return o, nil
}

const orderSelectSQL = `
	SELECT o.id,o.uuid,o.user_id,o.order_type,o.total_amount,o.status,
	       COALESCE(o.shipping_name,''),COALESCE(o.shipping_phone,''),COALESCE(o.shipping_addr,''),
	       COALESCE(o.note,''),o.payment_method,o.delivery_method,COALESCE(o.pickup_location,''),
	       o.payment_status,o.paid_at,o.shipped_at,COALESCE(o.tracking_number,''),
	       o.created_at,o.updated_at,COALESCE(u.username,'')
	FROM orders o LEFT JOIN users u ON u.id=o.user_id `

func (s *ShopService) scanOrder(row *sql.Row) (*models.Order, error) {
	o := &models.Order{}
	var pm, dm sql.NullInt64
	var paidAt, shippedAt sql.NullTime
	err := row.Scan(
		&o.ID, &o.UUID, &o.UserID, &o.OrderType, &o.TotalAmount, &o.Status,
		&o.ShippingName, &o.ShippingPhone, &o.ShippingAddr,
		&o.Note, &pm, &dm, &o.PickupLocation,
		&o.PaymentStatus, &paidAt, &shippedAt, &o.TrackingNumber,
		&o.CreatedAt, &o.UpdatedAt, &o.Username,
	)
	if errors.Is(err, sql.ErrNoRows) { return nil, ErrOrderNotFound }
	if err != nil { return nil, fmt.Errorf("scan order: %w", err) }
	if pm.Valid { v := int(pm.Int64); o.PaymentMethod = &v }
	if dm.Valid { v := int(dm.Int64); o.DeliveryMethod = &v }
	if paidAt.Valid { o.PaidAt = &paidAt.Time }
	if shippedAt.Valid { o.ShippedAt = &shippedAt.Time }
	return o, nil
}

func (s *ShopService) getOrderItems(ctx context.Context, orderID uint64) ([]*models.OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id,order_id,ref_type,ref_id,title,price,qty,COALESCE(spec,'')
		FROM order_items WHERE order_id=?`, orderID)
	if err != nil { return nil, err }
	defer rows.Close()
	var items []*models.OrderItem
	for rows.Next() {
		it := &models.OrderItem{}
		if err := rows.Scan(&it.ID, &it.OrderID, &it.RefType, &it.RefID, &it.Title, &it.Price, &it.Qty, &it.Spec); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

// ── List orders ───────────────────────────────────────────────

type ListOrdersInput struct {
	UserID   *uint64 `form:"-"`
	Status   *int    `form:"status"`
	Keyword  string  `form:"keyword"`
	Page     int     `form:"page"`
	PageSize int     `form:"page_size"`
}

type ListOrdersResult struct {
	Orders []*models.Order `json:"orders"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Pages  int             `json:"pages"`
}

func (s *ShopService) ListOrders(ctx context.Context, in ListOrdersInput) (*ListOrdersResult, error) {
	if in.Page < 1 { in.Page = 1 }
	if in.PageSize < 1 || in.PageSize > 100 { in.PageSize = 20 }

	where := "WHERE 1=1"
	args := []any{}
	if in.UserID != nil { where += " AND o.user_id=?"; args = append(args, *in.UserID) }
	if in.Status != nil { where += " AND o.status=?"; args = append(args, *in.Status) }
	if in.Keyword != "" {
		where += " AND (u.username LIKE ? OR o.shipping_name LIKE ?)"
		kw := "%" + in.Keyword + "%"
		args = append(args, kw, kw)
	}

	var total int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM orders o LEFT JOIN users u ON u.id=o.user_id "+where, args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	offset := (in.Page - 1) * in.PageSize
	rows, err := s.db.QueryContext(ctx,
		"SELECT o.id,o.uuid,o.user_id,o.order_type,o.total_amount,o.status,"+
			"COALESCE(o.shipping_name,''),COALESCE(o.shipping_phone,''),COALESCE(o.shipping_addr,''),"+
			"COALESCE(o.note,''),o.payment_method,o.delivery_method,COALESCE(o.pickup_location,''),"+
			"o.payment_status,o.paid_at,o.shipped_at,COALESCE(o.tracking_number,''),"+
			"o.created_at,o.updated_at,COALESCE(u.username,'') "+
			"FROM orders o LEFT JOIN users u ON u.id=o.user_id "+
			where+" ORDER BY o.created_at DESC LIMIT ? OFFSET ?",
		append(args, in.PageSize, offset)...,
	)
	if err != nil { return nil, fmt.Errorf("list orders: %w", err) }
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{}
		var pm, dm sql.NullInt64
		var paidAt, shippedAt sql.NullTime
		if err := rows.Scan(
			&o.ID, &o.UUID, &o.UserID, &o.OrderType, &o.TotalAmount, &o.Status,
			&o.ShippingName, &o.ShippingPhone, &o.ShippingAddr,
			&o.Note, &pm, &dm, &o.PickupLocation,
			&o.PaymentStatus, &paidAt, &shippedAt, &o.TrackingNumber,
			&o.CreatedAt, &o.UpdatedAt, &o.Username,
		); err != nil { return nil, err }
		if pm.Valid { v := int(pm.Int64); o.PaymentMethod = &v }
		if dm.Valid { v := int(dm.Int64); o.DeliveryMethod = &v }
		if paidAt.Valid { o.PaidAt = &paidAt.Time }
		if shippedAt.Valid { o.ShippedAt = &shippedAt.Time }
		orders = append(orders, o)
	}
	pages := (total + in.PageSize - 1) / in.PageSize
	if pages < 1 { pages = 1 }
	return &ListOrdersResult{Orders: orders, Total: total, Page: in.Page, Pages: pages}, nil
}

// ── Update order (admin) ──────────────────────────────────────

type UpdateOrderInput struct {
	Status         *int   `json:"status"`
	PaymentStatus  *int   `json:"payment_status"`
	ShippingName   string `json:"shipping_name"`
	ShippingPhone  string `json:"shipping_phone"`
	ShippingAddr   string `json:"shipping_addr"`
	Note           string `json:"note"`
	TrackingNumber string `json:"tracking_number"`
	PickupLocation string `json:"pickup_location"`
}

func (s *ShopService) UpdateOrder(ctx context.Context, id uint64, in UpdateOrderInput) (*models.Order, error) {
	now := time.Now()
	setClauses := "updated_at=?"
	args := []any{now}

	if in.Status != nil {
		setClauses += ",status=?"
		args = append(args, *in.Status)
		if *in.Status == models.OrderShipped {
			setClauses += ",shipped_at=?"
			args = append(args, now)
		}
	}
	if in.PaymentStatus != nil {
		setClauses += ",payment_status=?"
		args = append(args, *in.PaymentStatus)
		if *in.PaymentStatus == models.PaymentPaid {
			setClauses += ",paid_at=?"
			args = append(args, now)
		}
	}
	if in.ShippingName != ""   { setClauses += ",shipping_name=?";   args = append(args, in.ShippingName) }
	if in.ShippingPhone != ""  { setClauses += ",shipping_phone=?";  args = append(args, in.ShippingPhone) }
	if in.ShippingAddr != ""   { setClauses += ",shipping_addr=?";   args = append(args, in.ShippingAddr) }
	if in.Note != ""           { setClauses += ",note=?";            args = append(args, in.Note) }
	if in.TrackingNumber != "" { setClauses += ",tracking_number=?"; args = append(args, in.TrackingNumber) }
	if in.PickupLocation != "" { setClauses += ",pickup_location=?"; args = append(args, in.PickupLocation) }

	args = append(args, id)
	_, err := s.db.ExecContext(ctx, "UPDATE orders SET "+setClauses+" WHERE id=?", args...)
	if err != nil { return nil, fmt.Errorf("update order: %w", err) }
	return s.GetOrderByID(ctx, id)
}
