package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"trbb/internal/models"
	"trbb/pkg/database"
)

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrAlreadyRegistered = errors.New("already registered")
	ErrRegClosed         = errors.New("registration is closed")
	ErrEventFull         = errors.New("event is full")
	ErrRegNotFound       = errors.New("registration not found")
)

type EventService struct {
	db *database.DB
}

func NewEventService(db *database.DB) *EventService {
	return &EventService{db: db}
}

// ── List ─────────────────────────────────────────────────────

type ListEventsInput struct {
	Status   *int   `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ListEventsResult struct {
	Events []*models.Event `json:"events"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Pages  int             `json:"pages"`
}

func (s *EventService) ListPublic(ctx context.Context, in ListEventsInput) (*ListEventsResult, error) {
	published := models.EventPublished
	in.Status = &published
	return s.list(ctx, in)
}

func (s *EventService) ListAdmin(ctx context.Context, in ListEventsInput) (*ListEventsResult, error) {
	return s.list(ctx, in)
}

func (s *EventService) list(ctx context.Context, in ListEventsInput) (*ListEventsResult, error) {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.PageSize < 1 || in.PageSize > 100 {
		in.PageSize = 20
	}

	// 條件用 e. 前綴，避免 JOIN 後欄位名稱模糊
	where := "WHERE 1=1"
	args := []any{}
	if in.Status != nil {
		where += " AND e.status=?"
		args = append(args, *in.Status)
	}
	if in.Keyword != "" {
		where += " AND (e.title LIKE ? OR e.location LIKE ?)"
		kw := "%" + in.Keyword + "%"
		args = append(args, kw, kw)
	}

	// COUNT 查詢不需要 JOIN
	var total int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM events e "+where, args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count events: %w", err)
	}

	offset := (in.Page - 1) * in.PageSize
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.uuid, e.title,
		       COALESCE(e.description,''),
		       e.event_type,
		       COALESCE(e.location,''),
		       COALESCE(e.cover_url,''),
		       e.start_at, e.end_at,
		       e.reg_start_at, e.reg_end_at,
		       e.max_participants,
		       COALESCE(e.fee,0),
		       e.status, e.creator_id,
		       e.created_at, e.updated_at,
		       COUNT(r.id) AS reg_count
		FROM events e
		LEFT JOIN event_registrations r
		       ON r.event_id=e.id AND r.status IN (0,1)
		`+where+`
		GROUP BY e.id
		ORDER BY e.start_at ASC
		LIMIT ? OFFSET ?`,
		append(args, in.PageSize, offset)...,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := []*models.Event{}
	for rows.Next() {
		ev, err := scanEventRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	pages := (total + in.PageSize - 1) / in.PageSize
	if pages < 1 {
		pages = 1
	}
	return &ListEventsResult{Events: events, Total: total, Page: in.Page, Pages: pages}, nil
}

// ── Get single ────────────────────────────────────────────────

func (s *EventService) GetByID(ctx context.Context, id uint64) (*models.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.uuid, e.title,
		       COALESCE(e.description,''),
		       e.event_type,
		       COALESCE(e.location,''),
		       COALESCE(e.cover_url,''),
		       e.start_at, e.end_at,
		       e.reg_start_at, e.reg_end_at,
		       e.max_participants,
		       COALESCE(e.fee,0),
		       e.status, e.creator_id,
		       e.created_at, e.updated_at,
		       COUNT(r.id) AS reg_count
		FROM events e
		LEFT JOIN event_registrations r
		       ON r.event_id=e.id AND r.status IN (0,1)
		WHERE e.id=?
		GROUP BY e.id`, id)
	if err != nil {
		return nil, fmt.Errorf("query event: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrEventNotFound
	}
	ev, err := scanEventRow(rows)
	if err != nil {
		return nil, fmt.Errorf("scan event: %w", err)
	}
	return ev, nil
}

// scanEventRow — 通用的 row scan，同時支援 *sql.Rows
func scanEventRow(rows *sql.Rows) (*models.Event, error) {
	ev := &models.Event{}
	var maxP sql.NullInt64
	err := rows.Scan(
		&ev.ID, &ev.UUID, &ev.Title,
		&ev.Description,
		&ev.EventType,
		&ev.Location,
		&ev.CoverURL,
		&ev.StartAt, &ev.EndAt,
		&ev.RegStartAt, &ev.RegEndAt,
		&maxP,
		&ev.Fee,
		&ev.Status, &ev.CreatorID,
		&ev.CreatedAt, &ev.UpdatedAt,
		&ev.RegisteredCount,
	)
	if err != nil {
		return nil, err
	}
	if maxP.Valid {
		v := int(maxP.Int64)
		ev.MaxParticipants = &v
	}
	return ev, nil
}

// ── Create / Update / Delete ──────────────────────────────────

type EventInput struct {
	Title           string    `json:"title"            binding:"required"`
	Description     string    `json:"description"`
	EventType       int       `json:"event_type"`
	Location        string    `json:"location"         binding:"required"`
	CoverURL        string    `json:"cover_url"`
	StartAt         time.Time `json:"start_at"         binding:"required"`
	EndAt           time.Time `json:"end_at"`
	RegStartAt      time.Time `json:"reg_start_at"     binding:"required"`
	RegEndAt        time.Time `json:"reg_end_at"       binding:"required"`
	MaxParticipants *int      `json:"max_participants"`
	Fee             float64   `json:"fee"`
	Status          int       `json:"status"`
}

func (s *EventService) Create(ctx context.Context, in EventInput, creatorID uint64) (*models.Event, error) {
	uid := uuid.New().String()
	now := time.Now()
	endAt := in.EndAt
	if endAt.IsZero() {
		endAt = in.StartAt
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO events
		  (uuid,title,description,event_type,location,cover_url,
		   start_at,end_at,reg_start_at,reg_end_at,max_participants,
		   fee,status,creator_id,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		uid,
		in.Title, nullStr(in.Description),
		in.EventType, in.Location, nullStr(in.CoverURL),
		in.StartAt, endAt, in.RegStartAt, in.RegEndAt,
		in.MaxParticipants, in.Fee, in.Status, creatorID,
		now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, uint64(id))
}

func (s *EventService) Update(ctx context.Context, id uint64, in EventInput) (*models.Event, error) {
	endAt := in.EndAt
	if endAt.IsZero() {
		endAt = in.StartAt
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE events SET
		  title=?,description=?,event_type=?,location=?,cover_url=?,
		  start_at=?,end_at=?,reg_start_at=?,reg_end_at=?,max_participants=?,
		  fee=?,status=?,updated_at=?
		WHERE id=?`,
		in.Title, nullStr(in.Description),
		in.EventType, in.Location, nullStr(in.CoverURL),
		in.StartAt, endAt, in.RegStartAt, in.RegEndAt,
		in.MaxParticipants, in.Fee, in.Status,
		time.Now(), id,
	)
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	return s.GetByID(ctx, id)
}

func (s *EventService) Delete(ctx context.Context, id uint64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM events WHERE id=?", id)
	return err
}

// ── Register ──────────────────────────────────────────────────

type RegistrationInput struct {
	NameZh            string `json:"name_zh"`
	NameEn            string `json:"name_en"`
	IDNumber          string `json:"id_number"`
	PassportNumber    string `json:"passport_number"`
	Gender            *int   `json:"gender"`
	Birthday          string `json:"birthday"`
	Phone             string `json:"phone"  binding:"required"`
	Email             string `json:"email"  binding:"required,email"`
	ShirtSize         string `json:"shirt_size"`
	FoodType          *int   `json:"food_type"`
	Address           string `json:"address"`
	EmergencyContact  string `json:"emergency_contact"`
	EmergencyPhone    string `json:"emergency_phone"`
	EmergencyRelation string `json:"emergency_relation"`
	Note              string `json:"note"`
}

func (s *EventService) Register(ctx context.Context, eventID, userID uint64, in RegistrationInput) (*models.EventRegistration, error) {
	ev, err := s.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if !ev.IsOpen() {
		return nil, ErrRegClosed
	}
	if ev.MaxParticipants != nil && ev.RegisteredCount >= *ev.MaxParticipants {
		return nil, ErrEventFull
	}

	var count int
	s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM event_registrations WHERE event_id=? AND user_id=? AND status NOT IN (2,3)",
		eventID, userID,
	).Scan(&count)
	if count > 0 {
		return nil, ErrAlreadyRegistered
	}

	uid := uuid.New().String()
	now := time.Now()

	// birthday: 空字串存 NULL
	var birthdayVal any
	if in.Birthday != "" {
		birthdayVal = in.Birthday
	}

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO event_registrations
		  (uuid,event_id,user_id,status,note,
		   reg_name_zh,reg_name_en,reg_id_number,reg_passport_number,
		   reg_gender,reg_birthday,reg_phone,reg_email,
		   reg_shirt_size,reg_food_type,reg_address,
		   reg_emergency_contact,reg_emergency_phone,reg_emergency_relation,
		   created_at,updated_at)
		VALUES (?,?,?,0,?,
		        ?,?,?,?,
		        ?,?,?,?,
		        ?,?,?,
		        ?,?,?,
		        ?,?)`,
		uid, eventID, userID, nullStr(in.Note),
		in.NameZh, nullStr(in.NameEn), nullStr(in.IDNumber), nullStr(in.PassportNumber),
		in.Gender, birthdayVal, in.Phone, in.Email,
		nullStr(in.ShirtSize), in.FoodType, nullStr(in.Address),
		in.EmergencyContact, in.EmergencyPhone, in.EmergencyRelation,
		now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	id, _ := res.LastInsertId()
	return &models.EventRegistration{
		ID: uint64(id), UUID: uid, EventID: eventID, UserID: userID,
		Status: models.RegPending, CreatedAt: now,
		RegNameZh: in.NameZh, RegPhone: in.Phone, RegEmail: in.Email,
	}, nil
}

// ── Cancel ────────────────────────────────────────────────────

func (s *EventService) CancelRegistration(ctx context.Context, eventID, userID uint64) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE event_registrations SET status=2,updated_at=? WHERE event_id=? AND user_id=? AND status IN (0,1)",
		time.Now(), eventID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRegNotFound
	}
	return nil
}

// ── Get my registration ───────────────────────────────────────

func (s *EventService) GetMyRegistration(ctx context.Context, eventID, userID uint64) (*models.EventRegistration, error) {
	reg := &models.EventRegistration{}
	var gender, foodType sql.NullInt64
	var birthday, note sql.NullString

	err := s.db.QueryRowContext(ctx, `
		SELECT id,uuid,event_id,user_id,status,
		       COALESCE(note,''),
		       COALESCE(reg_name_zh,''),COALESCE(reg_name_en,''),
		       COALESCE(reg_id_number,''),COALESCE(reg_passport_number,''),
		       reg_gender,reg_birthday,
		       COALESCE(reg_phone,''),COALESCE(reg_email,''),
		       COALESCE(reg_shirt_size,''),reg_food_type,
		       COALESCE(reg_address,''),
		       COALESCE(reg_emergency_contact,''),COALESCE(reg_emergency_phone,''),
		       COALESCE(reg_emergency_relation,''),
		       created_at,updated_at
		FROM event_registrations
		WHERE event_id=? AND user_id=? AND status NOT IN (2,3)`,
		eventID, userID,
	).Scan(
		&reg.ID, &reg.UUID, &reg.EventID, &reg.UserID, &reg.Status,
		&note,
		&reg.RegNameZh, &reg.RegNameEn,
		&reg.RegIDNumber, &reg.RegPassportNumber,
		&gender, &birthday,
		&reg.RegPhone, &reg.RegEmail,
		&reg.RegShirtSize, &foodType,
		&reg.RegAddress,
		&reg.RegEmergencyContact, &reg.RegEmergencyPhone, &reg.RegEmergencyRelation,
		&reg.CreatedAt, &reg.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRegNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get my reg: %w", err)
	}
	reg.Note = note.String
	reg.RegBirthday = birthday.String
	if gender.Valid {
		v := int(gender.Int64)
		reg.RegGender = &v
	}
	if foodType.Valid {
		v := int(foodType.Int64)
		reg.RegFoodType = &v
	}
	return reg, nil
}

// ── List registrations (admin) ────────────────────────────────

func (s *EventService) ListRegistrations(ctx context.Context, eventID uint64) ([]*models.EventRegistration, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.uuid, r.event_id, r.user_id, r.status,
		       COALESCE(r.note,''),
		       COALESCE(r.reg_name_zh,''), COALESCE(r.reg_name_en,''),
		       COALESCE(r.reg_id_number,''), COALESCE(r.reg_passport_number,''),
		       r.reg_gender, r.reg_birthday,
		       COALESCE(r.reg_phone,''), COALESCE(r.reg_email,''),
		       COALESCE(r.reg_shirt_size,''), r.reg_food_type,
		       COALESCE(r.reg_address,''),
		       COALESCE(r.reg_emergency_contact,''),
		       COALESCE(r.reg_emergency_phone,''),
		       COALESCE(r.reg_emergency_relation,''),
		       r.created_at, r.updated_at,
		       u.username, COALESCE(u.display_name,''), u.email
		FROM event_registrations r
		JOIN users u ON u.id=r.user_id
		WHERE r.event_id=?
		ORDER BY r.created_at ASC`, eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}
	defer rows.Close()

	list := []*models.EventRegistration{}
	for rows.Next() {
		reg := &models.EventRegistration{}
		var gender, foodType sql.NullInt64
		var birthday sql.NullString
		if err := rows.Scan(
			&reg.ID, &reg.UUID, &reg.EventID, &reg.UserID, &reg.Status,
			&reg.Note,
			&reg.RegNameZh, &reg.RegNameEn,
			&reg.RegIDNumber, &reg.RegPassportNumber,
			&gender, &birthday,
			&reg.RegPhone, &reg.RegEmail,
			&reg.RegShirtSize, &foodType,
			&reg.RegAddress,
			&reg.RegEmergencyContact, &reg.RegEmergencyPhone, &reg.RegEmergencyRelation,
			&reg.CreatedAt, &reg.UpdatedAt,
			&reg.Username, &reg.DisplayName, &reg.Email,
		); err != nil {
			return nil, fmt.Errorf("scan registration: %w", err)
		}
		reg.RegBirthday = birthday.String
		if gender.Valid {
			v := int(gender.Int64)
			reg.RegGender = &v
		}
		if foodType.Valid {
			v := int(foodType.Int64)
			reg.RegFoodType = &v
		}
		list = append(list, reg)
	}
	return list, nil
}

// ── Update registration (admin) ───────────────────────────────

func (s *EventService) UpdateRegistration(ctx context.Context, regID uint64, in RegistrationInput, status int) error {
	var birthdayVal any
	if in.Birthday != "" {
		birthdayVal = in.Birthday
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE event_registrations SET
		  status=?,note=?,
		  reg_name_zh=?,reg_name_en=?,reg_id_number=?,reg_passport_number=?,
		  reg_gender=?,reg_birthday=?,reg_phone=?,reg_email=?,
		  reg_shirt_size=?,reg_food_type=?,reg_address=?,
		  reg_emergency_contact=?,reg_emergency_phone=?,reg_emergency_relation=?,
		  updated_at=?
		WHERE id=?`,
		status, nullStr(in.Note),
		in.NameZh, nullStr(in.NameEn), nullStr(in.IDNumber), nullStr(in.PassportNumber),
		in.Gender, birthdayVal, in.Phone, in.Email,
		nullStr(in.ShirtSize), in.FoodType, nullStr(in.Address),
		in.EmergencyContact, in.EmergencyPhone, in.EmergencyRelation,
		time.Now(), regID,
	)
	return err
}
