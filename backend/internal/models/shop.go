package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ── Product ──────────────────────────────────────────────────

const (
	ProductDraft     = 0
	ProductPublished = 1
	ProductSoldOut   = 2
)

const (
	CatApparel     = 1
	CatEquipment   = 2
	CatNutrition   = 3
	CatAccessories = 4
)

// StringSlice — JSON-serializable []string for MySQL JSON column
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}
func (s *StringSlice) Scan(src any) error {
	if src == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal(b, s)
}

type Product struct {
	ID          uint64      `json:"id"`
	UUID        string      `json:"uuid"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Category    int         `json:"category"`
	Price       float64     `json:"price"`
	Stock       int         `json:"stock"`
	Images      StringSlice `json:"images"` // []url stored as JSON
	Specs       StringSlice `json:"specs"`  // []spec string stored as JSON
	Status      int         `json:"status"`
	CreatorID   uint64      `json:"creator_id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ── Order ─────────────────────────────────────────────────────

const (
	OrderPending   = 0
	OrderPaid      = 1
	OrderShipped   = 2
	OrderCompleted = 3
	OrderCancelled = 4
	OrderRefunded  = 5
)

const (
	PayCreditCard = 1
	PayTransfer   = 2
	PayLinePay    = 3
	PayCash       = 4
)

const (
	DeliveryShipping = 1
	DeliveryPickup   = 2
)

const (
	PaymentUnpaid   = 0
	PaymentPaid     = 1
	PaymentRefunded = 2
)

type Order struct {
	ID             uint64     `json:"id"`
	UUID           string     `json:"uuid"`
	UserID         uint64     `json:"user_id"`
	OrderType      int        `json:"order_type"`
	TotalAmount    float64    `json:"total_amount"`
	Status         int        `json:"status"`
	ShippingName   string     `json:"shipping_name"`
	ShippingPhone  string     `json:"shipping_phone"`
	ShippingAddr   string     `json:"shipping_addr"`
	Note           string     `json:"note"`
	PaymentMethod  *int       `json:"payment_method"`
	DeliveryMethod *int       `json:"delivery_method"`
	PickupLocation string     `json:"pickup_location"`
	PaymentStatus  int        `json:"payment_status"`
	PaidAt         *time.Time `json:"paid_at"`
	ShippedAt      *time.Time `json:"shipped_at"`
	TrackingNumber string     `json:"tracking_number"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// joined
	Items    []*OrderItem `json:"items,omitempty"`
	Username string       `json:"username,omitempty"`
}

type OrderItem struct {
	ID      uint64  `json:"id"`
	OrderID uint64  `json:"order_id"`
	RefType string  `json:"ref_type"`
	RefID   uint64  `json:"ref_id"`
	Title   string  `json:"title"`
	Price   float64 `json:"price"`
	Qty     int     `json:"qty"`
	Spec    string  `json:"spec"`
}
