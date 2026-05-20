package repository

import (
	"database/sql"
	"sports-platform/internal/model"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) FindByUsername(username string) (*model.AdminUser, error) {
	a := &model.AdminUser{}
	var createdBy sql.NullInt64
	var lastLoginAt sql.NullTime
	err := r.db.QueryRow(`
		SELECT id, username, email, password_hash, name, role, status, created_by, last_login_at, created_at
		FROM admin_users WHERE username=? AND status=1`, username).
		Scan(&a.ID, &a.Username, &a.Email, &a.PasswordHash, &a.Name,
			&a.Role, &a.Status, &createdBy, &lastLoginAt, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if createdBy.Valid {
		v := uint64(createdBy.Int64)
		a.CreatedBy = &v
	}
	if lastLoginAt.Valid {
		a.LastLoginAt = &lastLoginAt.Time
	}
	return a, nil
}

func (r *AdminRepo) FindByID(id uint64) (*model.AdminUser, error) {
	a := &model.AdminUser{}
	var createdBy sql.NullInt64
	var lastLoginAt sql.NullTime
	err := r.db.QueryRow(`
		SELECT id, username, email, password_hash, name, role, status, created_by, last_login_at, created_at
		FROM admin_users WHERE id=?`, id).
		Scan(&a.ID, &a.Username, &a.Email, &a.PasswordHash, &a.Name,
			&a.Role, &a.Status, &createdBy, &lastLoginAt, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if createdBy.Valid {
		v := uint64(createdBy.Int64)
		a.CreatedBy = &v
	}
	if lastLoginAt.Valid {
		a.LastLoginAt = &lastLoginAt.Time
	}
	return a, nil
}

func (r *AdminRepo) ExistsByUsername(username string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM admin_users WHERE username=?`, username).Scan(&count)
	return count > 0, err
}

func (r *AdminRepo) ExistsByEmail(email string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM admin_users WHERE email=?`, email).Scan(&count)
	return count > 0, err
}

func (r *AdminRepo) Create(a *model.AdminUser) error {
	result, err := r.db.Exec(`
		INSERT INTO admin_users (username, email, password_hash, name, role, status, created_by)
		VALUES (?, ?, ?, ?, ?, 1, ?)`,
		a.Username, a.Email, a.PasswordHash, a.Name, a.Role, a.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	a.ID = uint64(id)

	// 一般管理員預設建立空白權限列
	if a.Role == 1 {
		r.db.Exec(`INSERT IGNORE INTO admin_permissions (admin_id) VALUES (?)`, a.ID)
	}
	return nil
}

func (r *AdminRepo) UpdateLastLogin(id uint64) {
	r.db.Exec(`UPDATE admin_users SET last_login_at=NOW() WHERE id=?`, id)
}

func (r *AdminRepo) UpdateStatus(id uint64, status int) error {
	_, err := r.db.Exec(`UPDATE admin_users SET status=?, updated_at=NOW() WHERE id=?`, status, id)
	return err
}

func (r *AdminRepo) List(limit, offset int) ([]model.AdminUser, int, error) {
	var total int
	r.db.QueryRow(`SELECT COUNT(*) FROM admin_users`).Scan(&total)
	rows, err := r.db.Query(`
		SELECT id, username, email, name, role, status, created_at
		FROM admin_users ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.AdminUser
	for rows.Next() {
		var a model.AdminUser
		rows.Scan(&a.ID, &a.Username, &a.Email, &a.Name, &a.Role, &a.Status, &a.CreatedAt)
		list = append(list, a)
	}
	return list, total, nil
}

func (r *AdminRepo) GetPermissions(adminID uint64) (*model.AdminPermissions, error) {
	p := &model.AdminPermissions{AdminID: adminID}
	err := r.db.QueryRow(`
		SELECT manage_members, manage_events, manage_products, manage_orders, manage_second_hand
		FROM admin_permissions WHERE admin_id=?`, adminID).
		Scan(&p.ManageMembers, &p.ManageEvents, &p.ManageProducts, &p.ManageOrders, &p.ManageSecondHand)
	if err == sql.ErrNoRows {
		return p, nil // 空權限
	}
	return p, err
}

func (r *AdminRepo) SetPermissions(p *model.AdminPermissions) error {
	_, err := r.db.Exec(`
		INSERT INTO admin_permissions (admin_id, manage_members, manage_events, manage_products, manage_orders, manage_second_hand)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			manage_members=VALUES(manage_members),
			manage_events=VALUES(manage_events),
			manage_products=VALUES(manage_products),
			manage_orders=VALUES(manage_orders),
			manage_second_hand=VALUES(manage_second_hand),
			updated_at=NOW()`,
		p.AdminID, p.ManageMembers, p.ManageEvents, p.ManageProducts, p.ManageOrders, p.ManageSecondHand)
	return err
}

// CountSuperAdmins - 確保至少保留一個超管
func (r *AdminRepo) CountSuperAdmins() int {
	var count int
	r.db.QueryRow(`SELECT COUNT(*) FROM admin_users WHERE role=2 AND status=1`).Scan(&count)
	return count
}
