package repository

import (
	"database/sql"
	"sports-platform/internal/model"
	"time"
)

type MemberRepo struct {
	db *sql.DB
}

func NewMemberRepo(db *sql.DB) *MemberRepo {
	return &MemberRepo{db: db}
}

func (r *MemberRepo) Create(m *model.Member) error {
	_, err := r.db.Exec(`
		INSERT INTO members (uuid, member_no, email, phone, password_hash, name, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, 0, NOW(), NOW())`,
		m.UUID, m.MemberNo, m.Email, m.Phone, m.PasswordHash, m.Name)
	return err
}

func (r *MemberRepo) scanMember(row *sql.Row) (*model.Member, error) {
	m := &model.Member{}
	var memberNo, phone, avatarURL sql.NullString
	err := row.Scan(&m.ID, &m.UUID, &memberNo, &m.Email, &phone,
		&m.PasswordHash, &m.Name, &avatarURL, &m.Gender, &m.Role, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.MemberNo = memberNo.String
	m.Phone = phone.String
	m.AvatarURL = avatarURL.String
	return m, nil
}

func (r *MemberRepo) FindByEmail(email string) (*model.Member, error) {
	row := r.db.QueryRow(`
		SELECT id, uuid, member_no, email, phone, password_hash, name, avatar_url, gender, role, status, created_at
		FROM members WHERE email = ?`, email)
	return r.scanMember(row)
}

func (r *MemberRepo) FindByID(id uint64) (*model.Member, error) {
	row := r.db.QueryRow(`
		SELECT id, uuid, member_no, email, phone, password_hash, name, avatar_url, gender, role, status, created_at
		FROM members WHERE id = ?`, id)
	return r.scanMember(row)
}

func (r *MemberRepo) ExistsByMemberNo(memberNo string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM members WHERE member_no=?`, memberNo).Scan(&count)
	return count > 0, err
}

func (r *MemberRepo) UpdateProfile(id uint64, name, phone, avatarURL string, gender int, birthday *time.Time) error {
	_, err := r.db.Exec(`
		UPDATE members SET name=?, phone=?, avatar_url=?, gender=?, birthday=?, updated_at=NOW()
		WHERE id=?`,
		name, phone, avatarURL, gender, birthday, id)
	return err
}

func (r *MemberRepo) UpdatePassword(id uint64, hash string) error {
	_, err := r.db.Exec(`UPDATE members SET password_hash=?, updated_at=NOW() WHERE id=?`, hash, id)
	return err
}

func (r *MemberRepo) UpdateStatus(id uint64, status int) error {
	_, err := r.db.Exec(`UPDATE members SET status=?, updated_at=NOW() WHERE id=?`, status, id)
	return err
}

// Admin: list members with optional status filter
func (r *MemberRepo) AdminList(limit, offset int, status *int) ([]model.Member, int, error) {
	where := ""
	args := []interface{}{}
	if status != nil {
		where = "WHERE status=?"
		args = append(args, *status)
	}
	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM members "+where, args...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := r.db.Query(`
		SELECT id, uuid, member_no, email, phone, password_hash, name, avatar_url, gender, role, status, created_at
		FROM members `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Member
	for rows.Next() {
		var m model.Member
		var memberNo, phone, avatarURL sql.NullString
		rows.Scan(&m.ID, &m.UUID, &memberNo, &m.Email, &phone,
			&m.PasswordHash, &m.Name, &avatarURL, &m.Gender, &m.Role, &m.Status, &m.CreatedAt)
		m.MemberNo = memberNo.String
		m.Phone = phone.String
		m.AvatarURL = avatarURL.String
		m.PasswordHash = "" // 不回傳
		list = append(list, m)
	}
	return list, total, nil
}

// Addresses
func (r *MemberRepo) ListAddresses(memberID uint64) ([]model.MemberAddress, error) {
	rows, err := r.db.Query(`
		SELECT id, member_id, label, recipient, phone, zip, city, district, address, is_default
		FROM member_addresses WHERE member_id=? ORDER BY is_default DESC, id DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.MemberAddress
	for rows.Next() {
		var a model.MemberAddress
		var label, zip, city, district sql.NullString
		rows.Scan(&a.ID, &a.MemberID, &label, &a.Recipient, &a.Phone,
			&zip, &city, &district, &a.Address, &a.IsDefault)
		a.Label = label.String
		a.Zip = zip.String
		a.City = city.String
		a.District = district.String
		list = append(list, a)
	}
	return list, nil
}

func (r *MemberRepo) CreateAddress(a *model.MemberAddress) error {
	if a.IsDefault == 1 {
		r.db.Exec(`UPDATE member_addresses SET is_default=0 WHERE member_id=?`, a.MemberID)
	}
	_, err := r.db.Exec(`
		INSERT INTO member_addresses (member_id, label, recipient, phone, zip, city, district, address, is_default)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.MemberID, a.Label, a.Recipient, a.Phone, a.Zip, a.City, a.District, a.Address, a.IsDefault)
	return err
}

func (r *MemberRepo) UpdateAddress(a *model.MemberAddress) error {
	if a.IsDefault == 1 {
		r.db.Exec(`UPDATE member_addresses SET is_default=0 WHERE member_id=?`, a.MemberID)
	}
	_, err := r.db.Exec(`
		UPDATE member_addresses SET label=?, recipient=?, phone=?, zip=?, city=?, district=?, address=?, is_default=?
		WHERE id=? AND member_id=?`,
		a.Label, a.Recipient, a.Phone, a.Zip, a.City, a.District, a.Address, a.IsDefault, a.ID, a.MemberID)
	return err
}

func (r *MemberRepo) DeleteAddress(id, memberID uint64) error {
	_, err := r.db.Exec(`DELETE FROM member_addresses WHERE id=? AND member_id=?`, id, memberID)
	return err
}

func (r *MemberRepo) LogLogin(memberID uint64, ip, userAgent string) {
	r.db.Exec(`INSERT INTO login_logs (member_id, ip, user_agent, created_at) VALUES (?, ?, ?, NOW())`,
		memberID, ip, userAgent)
}
