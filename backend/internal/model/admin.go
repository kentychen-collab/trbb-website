package model

import "time"

type AdminUser struct {
	ID           uint64     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Name         string     `json:"name"`
	Role         int        `json:"role"` // 1=一般管理員 2=超級管理員
	Status       int        `json:"status"`
	CreatedBy    *uint64    `json:"created_by,omitempty"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	Permissions  *AdminPermissions `json:"permissions,omitempty"`
}

type AdminPermissions struct {
	AdminID          uint64 `json:"admin_id"`
	ManageMembers    bool   `json:"manage_members"`
	ManageEvents     bool   `json:"manage_events"`
	ManageProducts   bool   `json:"manage_products"`
	ManageOrders     bool   `json:"manage_orders"`
	ManageSecondHand bool   `json:"manage_second_hand"`
}
