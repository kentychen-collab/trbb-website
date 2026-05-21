package model

import "time"

type Event struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Category    string     `json:"category"`
	CoverImage  string     `json:"cover_image"`
	Description string     `json:"description"`
	EventDate   string     `json:"event_date"`
	Location    string     `json:"location"`
	RegStartAt  *time.Time `json:"reg_start_at"`
	RegEndAt    *time.Time `json:"reg_end_at"`
	MaxQuota    *int       `json:"max_quota"`
	Status      int        `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	Tickets     []EventTicket `json:"tickets,omitempty"`
}

type EventTicket struct {
	ID        uint64     `json:"id"`
	EventID   uint64     `json:"event_id"`
	Name      string     `json:"name"`
	Price     float64    `json:"price"`
	Quota     *int       `json:"quota"`
	SoldCount int        `json:"sold_count"`
	RegStart  *time.Time `json:"reg_start"`
	RegEnd    *time.Time `json:"reg_end"`
}

type EventRegistration struct {
	ID               uint64    `json:"id"`
	OrderID          uint64    `json:"order_id"`
	EventID          uint64    `json:"event_id"`
	TicketID         uint64    `json:"ticket_id"`
	MemberID         uint64    `json:"member_id"`
	Participant      string    `json:"participant"`
	EmergencyContact string    `json:"emergency_contact"`
	EmergencyPhone   string    `json:"emergency_phone"`
	TshirtSize       string    `json:"tshirt_size"`
	Status           int       `json:"status"`
	// Joined fields
	EventTitle    string `json:"event_title,omitempty"`
	EventDate     string `json:"event_date,omitempty"`
	EventLocation string `json:"event_location,omitempty"`
}
