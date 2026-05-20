package model

import "time"

type Order struct {
	ID             uint64     `json:"id"`
	OrderNo        string     `json:"order_no"`
	MemberID       uint64     `json:"member_id"`
	Type           int        `json:"type"`
	Status         int        `json:"status"`
	PaymentMethod  string     `json:"payment_method"`
	PaymentStatus  int        `json:"payment_status"`
	PaymentAt      *time.Time `json:"payment_at"`
	ShippingFee    float64    `json:"shipping_fee"`
	TotalAmount    float64    `json:"total_amount"`
	DiscountAmount float64    `json:"discount_amount"`
	FinalAmount    float64    `json:"final_amount"`
	RecipientName  string     `json:"recipient_name"`
	RecipientPhone string     `json:"recipient_phone"`
	RecipientAddr  string     `json:"recipient_addr"`
	Note           string     `json:"note"`
	ECPayTradeNo   string     `json:"ecpay_trade_no"`
	CreatedAt      time.Time  `json:"created_at"`
	Items          []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        uint64  `json:"id"`
	OrderID   uint64  `json:"order_id"`
	ProductID uint64  `json:"product_id"`
	VariantID uint64  `json:"variant_id"`
	Name      string  `json:"name"`
	Spec      string  `json:"spec"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}
