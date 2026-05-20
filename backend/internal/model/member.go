package model

import "time"

// Member status
const (
	MemberStatusPending  = 0 // 待審核
	MemberStatusActive   = 1 // 啟用
	MemberStatusDisabled = 2 // 停用
)

type Member struct {
	ID           uint64     `json:"id"`
	UUID         string     `json:"uuid"`
	MemberNo     string     `json:"member_no"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	PasswordHash string     `json:"-"`
	Name         string     `json:"name"`
	AvatarURL    string     `json:"avatar_url"`
	Gender       int        `json:"gender"`
	Birthday     *time.Time `json:"birthday,omitempty"`
	Role         int        `json:"role"`
	Status       int        `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

type MemberAddress struct {
	ID        uint64 `json:"id"`
	MemberID  uint64 `json:"member_id"`
	Label     string `json:"label"`
	Recipient string `json:"recipient"`
	Phone     string `json:"phone"`
	Zip       string `json:"zip"`
	City      string `json:"city"`
	District  string `json:"district"`
	Address   string `json:"address"`
	IsDefault int    `json:"is_default"`
}
