package repository

import (
	"database/sql"
	"sports-platform/internal/model"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(o *model.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	result, err := tx.Exec(`
		INSERT INTO orders (order_no, member_id, type, status, payment_method, shipping_fee,
		                    total_amount, discount_amount, final_amount,
		                    recipient_name, recipient_phone, recipient_addr, note)
		VALUES (?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.OrderNo, o.MemberID, o.Type, o.PaymentMethod, o.ShippingFee,
		o.TotalAmount, o.DiscountAmount, o.FinalAmount,
		o.RecipientName, o.RecipientPhone, o.RecipientAddr, o.Note)
	if err != nil {
		tx.Rollback()
		return err
	}
	id, _ := result.LastInsertId()
	o.ID = uint64(id)

	for i := range o.Items {
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, variant_id, name, spec, price, quantity, subtotal)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			o.ID, o.Items[i].ProductID, o.Items[i].VariantID,
			o.Items[i].Name, o.Items[i].Spec, o.Items[i].Price,
			o.Items[i].Quantity, o.Items[i].Subtotal)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *OrderRepo) FindByOrderNo(orderNo string) (*model.Order, error) {
	o := &model.Order{}
	var paymentMethod, recipientName, recipientPhone, recipientAddr, note sql.NullString
	var paymentAt sql.NullTime
	err := r.db.QueryRow(`
		SELECT id, order_no, member_id, type, status, payment_method, payment_status,
		       shipping_fee, total_amount, discount_amount, final_amount,
		       recipient_name, recipient_phone, recipient_addr, note, created_at
		FROM orders WHERE order_no=?`, orderNo).
		Scan(&o.ID, &o.OrderNo, &o.MemberID, &o.Type, &o.Status, &paymentMethod, &o.PaymentStatus,
			&o.ShippingFee, &o.TotalAmount, &o.DiscountAmount, &o.FinalAmount,
			&recipientName, &recipientPhone, &recipientAddr, &note, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.PaymentMethod = paymentMethod.String
	o.RecipientName = recipientName.String
	o.RecipientPhone = recipientPhone.String
	o.RecipientAddr = recipientAddr.String
	o.Note = note.String
	_ = paymentAt
	o.Items, _ = r.listItems(o.ID)
	return o, nil
}

func (r *OrderRepo) ListByMember(memberID uint64, limit, offset int) ([]model.Order, int, error) {
	var total int
	r.db.QueryRow(`SELECT COUNT(*) FROM orders WHERE member_id=?`, memberID).Scan(&total)

	rows, err := r.db.Query(`
		SELECT id, order_no, type, status, final_amount, created_at
		FROM orders WHERE member_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		memberID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Order
	for rows.Next() {
		var o model.Order
		rows.Scan(&o.ID, &o.OrderNo, &o.Type, &o.Status, &o.FinalAmount, &o.CreatedAt)
		list = append(list, o)
	}
	return list, total, nil
}

func (r *OrderRepo) UpdateStatus(id uint64, status int) error {
	_, err := r.db.Exec(`UPDATE orders SET status=?, updated_at=NOW() WHERE id=?`, status, id)
	return err
}

func (r *OrderRepo) UpdatePayment(orderNo, tradeNo string) error {
	_, err := r.db.Exec(`
		UPDATE orders SET payment_status=1, payment_at=NOW(), ecpay_trade_no=?, updated_at=NOW()
		WHERE order_no=?`, tradeNo, orderNo)
	return err
}

func (r *OrderRepo) listItems(orderID uint64) ([]model.OrderItem, error) {
	rows, err := r.db.Query(`
		SELECT id, order_id, product_id, variant_id, name, spec, price, quantity, subtotal
		FROM order_items WHERE order_id=?`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		var productID, variantID sql.NullInt64
		var spec sql.NullString
		rows.Scan(&item.ID, &item.OrderID, &productID, &variantID,
			&item.Name, &spec, &item.Price, &item.Quantity, &item.Subtotal)
		item.ProductID = uint64(productID.Int64)
		item.VariantID = uint64(variantID.Int64)
		item.Spec = spec.String
		list = append(list, item)
	}
	return list, nil
}

// Admin
func (r *OrderRepo) AdminList(limit, offset int) ([]model.Order, int, error) {
	var total int
	r.db.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&total)
	rows, err := r.db.Query(`
		SELECT id, order_no, member_id, type, status, payment_status, final_amount, created_at
		FROM orders ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Order
	for rows.Next() {
		var o model.Order
		rows.Scan(&o.ID, &o.OrderNo, &o.MemberID, &o.Type, &o.Status, &o.PaymentStatus, &o.FinalAmount, &o.CreatedAt)
		list = append(list, o)
	}
	return list, total, nil
}
