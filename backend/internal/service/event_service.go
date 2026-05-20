package service

import (
	"errors"
	"fmt"
	"sports-platform/internal/model"
	"sports-platform/internal/repository"
	"time"
)

type EventService struct {
	eventRepo *repository.EventRepo
	orderRepo *repository.OrderRepo
}

func NewEventService(eventRepo *repository.EventRepo, orderRepo *repository.OrderRepo) *EventService {
	return &EventService{eventRepo: eventRepo, orderRepo: orderRepo}
}

func (s *EventService) List(page, limit int) ([]model.Event, int, error) {
	offset := (page - 1) * limit
	return s.eventRepo.List(limit, offset)
}

func (s *EventService) GetBySlug(slug string) (*model.Event, error) {
	e, err := s.eventRepo.FindBySlug(slug)
	if err != nil || e == nil {
		return nil, err
	}
	e.Tickets, _ = s.eventRepo.ListTickets(e.ID)
	return e, nil
}

type RegisterEventInput struct {
	TicketID         uint64 `json:"ticket_id" binding:"required"`
	Participant      string `json:"participant" binding:"required"`
	EmergencyContact string `json:"emergency_contact"`
	EmergencyPhone   string `json:"emergency_phone"`
	TshirtSize       string `json:"tshirt_size"`
}

func (s *EventService) Register(eventID, memberID uint64, input RegisterEventInput) (*model.EventRegistration, error) {
	event, err := s.eventRepo.FindByID(eventID)
	if err != nil || event == nil {
		return nil, errors.New("event not found")
	}

	ticket, err := s.eventRepo.FindTicket(input.TicketID, eventID)
	if err != nil || ticket == nil {
		return nil, errors.New("ticket not found")
	}

	now := time.Now()
	if ticket.RegStart != nil && now.Before(*ticket.RegStart) {
		return nil, errors.New("registration not started yet")
	}
	if ticket.RegEnd != nil && now.After(*ticket.RegEnd) {
		return nil, errors.New("registration has ended")
	}
	if ticket.Quota != nil && ticket.SoldCount >= *ticket.Quota {
		return nil, errors.New("ticket sold out")
	}

	// Create order
	orderNo := fmt.Sprintf("EVT%s%06d", time.Now().Format("20060102150405"), memberID%1000000)
	order := &model.Order{
		OrderNo:     orderNo,
		MemberID:    memberID,
		Type:        2, // 團報
		PaymentMethod: "pending",
		TotalAmount:   ticket.Price,
		FinalAmount:   ticket.Price,
	}
	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	reg := &model.EventRegistration{
		OrderID:          order.ID,
		EventID:          eventID,
		TicketID:         input.TicketID,
		MemberID:         memberID,
		Participant:      input.Participant,
		EmergencyContact: input.EmergencyContact,
		EmergencyPhone:   input.EmergencyPhone,
		TshirtSize:       input.TshirtSize,
	}
	if err := s.eventRepo.CreateRegistration(reg); err != nil {
		return nil, err
	}

	s.eventRepo.IncrTicketSold(input.TicketID)
	return reg, nil
}

func (s *EventService) MyRegistrations(memberID uint64) ([]model.EventRegistration, error) {
	return s.eventRepo.ListRegistrationsByMember(memberID)
}

// Admin
func (s *EventService) AdminList(page, limit int) ([]model.Event, int, error) {
	offset := (page - 1) * limit
	return s.eventRepo.AdminList(limit, offset)
}

type CreateEventInput struct {
	Title      string  `json:"title" binding:"required"`
	Slug       string  `json:"slug" binding:"required"`
	Category   string  `json:"category"`
	CoverImage string  `json:"cover_image"`
	Description string `json:"description"`
	EventDate  string  `json:"event_date" binding:"required"`
	Location   string  `json:"location"`
	RegStartAt *string `json:"reg_start_at"`
	RegEndAt   *string `json:"reg_end_at"`
	MaxQuota   *int    `json:"max_quota"`
}

func (s *EventService) AdminCreate(input CreateEventInput) (*model.Event, error) {
	e := &model.Event{
		Title:       input.Title,
		Slug:        input.Slug,
		Category:    input.Category,
		CoverImage:  input.CoverImage,
		Description: input.Description,
		EventDate:   input.EventDate,
		Location:    input.Location,
		MaxQuota:    input.MaxQuota,
	}
	if input.RegStartAt != nil {
		t, err := time.Parse("2006-01-02 15:04:05", *input.RegStartAt)
		if err == nil {
			e.RegStartAt = &t
		}
	}
	if input.RegEndAt != nil {
		t, err := time.Parse("2006-01-02 15:04:05", *input.RegEndAt)
		if err == nil {
			e.RegEndAt = &t
		}
	}
	if err := s.eventRepo.Create(e); err != nil {
		return nil, err
	}
	return e, nil
}
