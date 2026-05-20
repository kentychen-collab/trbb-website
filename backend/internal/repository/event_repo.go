package repository

import (
	"database/sql"
	"sports-platform/internal/model"
)

type EventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

func scanEvent(row *sql.Row) (*model.Event, error) {
	e := &model.Event{}
	var slug, category, coverImage, description, location sql.NullString
	var regStartAt, regEndAt sql.NullTime
	var maxQuota sql.NullInt64
	err := row.Scan(&e.ID, &e.Title, &slug, &category, &coverImage,
		&description, &e.EventDate, &location,
		&regStartAt, &regEndAt, &maxQuota, &e.Status, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.Slug = slug.String
	e.Category = category.String
	e.CoverImage = coverImage.String
	e.Description = description.String
	e.Location = location.String
	if regStartAt.Valid {
		e.RegStartAt = &regStartAt.Time
	}
	if regEndAt.Valid {
		e.RegEndAt = &regEndAt.Time
	}
	if maxQuota.Valid {
		q := int(maxQuota.Int64)
		e.MaxQuota = &q
	}
	return e, nil
}

func scanEventRow(rows *sql.Rows) (model.Event, error) {
	e := model.Event{}
	var slug, category, coverImage, description, location sql.NullString
	var regStartAt, regEndAt sql.NullTime
	var maxQuota sql.NullInt64
	err := rows.Scan(&e.ID, &e.Title, &slug, &category, &coverImage,
		&description, &e.EventDate, &location,
		&regStartAt, &regEndAt, &maxQuota, &e.Status, &e.CreatedAt)
	if err != nil {
		return e, err
	}
	e.Slug = slug.String
	e.Category = category.String
	e.CoverImage = coverImage.String
	e.Description = description.String
	e.Location = location.String
	if regStartAt.Valid {
		e.RegStartAt = &regStartAt.Time
	}
	if regEndAt.Valid {
		e.RegEndAt = &regEndAt.Time
	}
	if maxQuota.Valid {
		q := int(maxQuota.Int64)
		e.MaxQuota = &q
	}
	return e, nil
}

func (r *EventRepo) List(limit, offset int) ([]model.Event, int, error) {
	var total int
	r.db.QueryRow(`SELECT COUNT(*) FROM events WHERE status=1`).Scan(&total)

	rows, err := r.db.Query(`
		SELECT id, title, slug, category, cover_image, description, event_date, location,
		       reg_start_at, reg_end_at, max_quota, status, created_at
		FROM events WHERE status=1
		ORDER BY event_date ASC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Event
	for rows.Next() {
		e, err := scanEventRow(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, nil
}

func (r *EventRepo) FindBySlug(slug string) (*model.Event, error) {
	row := r.db.QueryRow(`
		SELECT id, title, slug, category, cover_image, description, event_date, location,
		       reg_start_at, reg_end_at, max_quota, status, created_at
		FROM events WHERE slug=? AND status=1`, slug)
	return scanEvent(row)
}

func (r *EventRepo) FindByID(id uint64) (*model.Event, error) {
	row := r.db.QueryRow(`
		SELECT id, title, slug, category, cover_image, description, event_date, location,
		       reg_start_at, reg_end_at, max_quota, status, created_at
		FROM events WHERE id=? AND status=1`, id)
	return scanEvent(row)
}

func (r *EventRepo) ListTickets(eventID uint64) ([]model.EventTicket, error) {
	rows, err := r.db.Query(`
		SELECT id, event_id, name, price, quota, sold_count, reg_start, reg_end
		FROM event_tickets WHERE event_id=?`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.EventTicket
	for rows.Next() {
		var t model.EventTicket
		var quota sql.NullInt64
		var regStart, regEnd sql.NullTime
		rows.Scan(&t.ID, &t.EventID, &t.Name, &t.Price, &quota, &t.SoldCount, &regStart, &regEnd)
		if quota.Valid {
			q := int(quota.Int64)
			t.Quota = &q
		}
		if regStart.Valid {
			t.RegStart = &regStart.Time
		}
		if regEnd.Valid {
			t.RegEnd = &regEnd.Time
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *EventRepo) FindTicket(ticketID, eventID uint64) (*model.EventTicket, error) {
	t := &model.EventTicket{}
	var quota sql.NullInt64
	var regStart, regEnd sql.NullTime
	err := r.db.QueryRow(`
		SELECT id, event_id, name, price, quota, sold_count, reg_start, reg_end
		FROM event_tickets WHERE id=? AND event_id=?`, ticketID, eventID).
		Scan(&t.ID, &t.EventID, &t.Name, &t.Price, &quota, &t.SoldCount, &regStart, &regEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if quota.Valid {
		q := int(quota.Int64)
		t.Quota = &q
	}
	if regStart.Valid {
		t.RegStart = &regStart.Time
	}
	if regEnd.Valid {
		t.RegEnd = &regEnd.Time
	}
	return t, nil
}

func (r *EventRepo) IncrTicketSold(ticketID uint64) error {
	_, err := r.db.Exec(`UPDATE event_tickets SET sold_count=sold_count+1 WHERE id=?`, ticketID)
	return err
}

func (r *EventRepo) CreateRegistration(reg *model.EventRegistration) error {
	result, err := r.db.Exec(`
		INSERT INTO event_registrations
		(order_id, event_id, ticket_id, member_id, participant, emergency_contact, emergency_phone, tshirt_size, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		reg.OrderID, reg.EventID, reg.TicketID, reg.MemberID,
		reg.Participant, reg.EmergencyContact, reg.EmergencyPhone, reg.TshirtSize)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	reg.ID = uint64(id)
	return nil
}

func (r *EventRepo) ListRegistrationsByMember(memberID uint64) ([]model.EventRegistration, error) {
	rows, err := r.db.Query(`
		SELECT r.id, r.order_id, r.event_id, r.ticket_id, r.member_id,
		       r.participant, r.emergency_contact, r.emergency_phone, r.tshirt_size, r.status,
		       e.title, e.event_date, e.location
		FROM event_registrations r
		JOIN events e ON e.id = r.event_id
		WHERE r.member_id=? ORDER BY r.id DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.EventRegistration
	for rows.Next() {
		var reg model.EventRegistration
		var emergencyContact, emergencyPhone, tshirtSize, location sql.NullString
		rows.Scan(&reg.ID, &reg.OrderID, &reg.EventID, &reg.TicketID, &reg.MemberID,
			&reg.Participant, &emergencyContact, &emergencyPhone, &tshirtSize, &reg.Status,
			&reg.EventTitle, &reg.EventDate, &location)
		reg.EmergencyContact = emergencyContact.String
		reg.EmergencyPhone = emergencyPhone.String
		reg.TshirtSize = tshirtSize.String
		reg.EventLocation = location.String
		list = append(list, reg)
	}
	return list, nil
}

// Admin
func (r *EventRepo) AdminList(limit, offset int) ([]model.Event, int, error) {
	var total int
	r.db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&total)
	rows, err := r.db.Query(`
		SELECT id, title, slug, category, cover_image, description, event_date, location,
		       reg_start_at, reg_end_at, max_quota, status, created_at
		FROM events ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Event
	for rows.Next() {
		e, err := scanEventRow(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, nil
}

func (r *EventRepo) Create(e *model.Event) error {
	result, err := r.db.Exec(`
		INSERT INTO events (title, slug, category, cover_image, description, event_date, location,
		                    reg_start_at, reg_end_at, max_quota, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		e.Title, e.Slug, e.Category, e.CoverImage, e.Description,
		e.EventDate, e.Location, e.RegStartAt, e.RegEndAt, e.MaxQuota)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	e.ID = uint64(id)
	return nil
}

func (r *EventRepo) Update(e *model.Event) error {
	_, err := r.db.Exec(`
		UPDATE events SET title=?, category=?, cover_image=?, description=?,
		event_date=?, location=?, reg_start_at=?, reg_end_at=?, max_quota=?, status=?
		WHERE id=?`,
		e.Title, e.Category, e.CoverImage, e.Description,
		e.EventDate, e.Location, e.RegStartAt, e.RegEndAt, e.MaxQuota, e.Status, e.ID)
	return err
}
