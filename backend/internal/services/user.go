package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"trbb/internal/models"
	"trbb/pkg/database"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already registered")
	ErrUsernameExists   = errors.New("username already taken")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrAccountPending   = errors.New("account is pending admin approval")
	ErrAccountSuspended = errors.New("account is suspended")
	ErrAccountRejected  = errors.New("account registration was rejected")
	ErrPermissionDenied = errors.New("permission denied")
)

type UserService struct {
	db        *database.DB
	jwtSecret string
}

func NewUserService(db *database.DB, jwtSecret string) *UserService {
	return &UserService{db: db, jwtSecret: jwtSecret}
}

// ── Register ─────────────────────────────────────────────────

type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	RealName string `json:"real_name" binding:"required,min=1,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Phone    string `json:"phone"    binding:"required,min=8,max=20"`
	Password string `json:"password" binding:"required,min=8"`
}

// isChinese returns true if s contains any CJK character
func isChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

func (s *UserService) Register(ctx context.Context, in RegisterInput) (*models.User, error) {
	var count int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE email=? AND deleted_at IS NULL", in.Email,
	).Scan(&count); err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if count > 0 {
		return nil, ErrEmailExists
	}
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE username=? AND deleted_at IS NULL", in.Username,
	).Scan(&count); err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if count > 0 {
		return nil, ErrUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	uid := uuid.New().String()
	now := time.Now()

	// 依姓名語言自動分配至 name_zh / name_en
	var nameZh, nameEn, displayName string
	if isChinese(in.RealName) {
		nameZh = in.RealName
	} else {
		nameEn = in.RealName
	}
	displayName = in.RealName // display_name 直接用真實姓名

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO users
		  (uuid,username,email,phone,password_hash,display_name,name_zh,name_en,role,status,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		uid, in.Username, in.Email, in.Phone, string(hash),
		displayName, nullStr(nameZh), nullStr(nameEn),
		models.RoleMember, models.StatusPending, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	id, _ := res.LastInsertId()
	return &models.User{
		ID: uint64(id), UUID: uid, Username: in.Username,
		Email: in.Email, Phone: in.Phone, DisplayName: displayName,
		NameZh: nameZh, NameEn: nameEn,
		Role: models.RoleMember, Status: models.StatusPending,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

// ── Login ────────────────────────────────────────────────────

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResult struct {
	Token   string       `json:"token"`
	Refresh string       `json:"refresh_token"`
	User    *models.User `json:"user"`
}

func (s *UserService) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	user, err := s.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidPassword
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, ErrInvalidPassword
	}
	switch user.Status {
	case models.StatusPending:
		return nil, ErrAccountPending
	case models.StatusSuspended:
		return nil, ErrAccountSuspended
	case models.StatusRejected:
		return nil, ErrAccountRejected
	}
	token, refresh, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, Refresh: refresh, User: user}, nil
}

// ── Update Profile ────────────────────────────────────────────

type UpdateProfileInput struct {
	DisplayName     string  `json:"display_name"`
	NameZh          string  `json:"name_zh"`
	NameEn          string  `json:"name_en"`
	IDNumber        string  `json:"id_number"`
	PassportNumber  string  `json:"passport_number"`
	Gender          *int    `json:"gender"`
	Birthday        string  `json:"birthday"`
	Phone           string  `json:"phone"`
	ShirtSize       string  `json:"shirt_size"`
	FoodType        *int    `json:"food_type"`
	Address         string  `json:"address"`
	EmergencyContact  string `json:"emergency_contact"`
	EmergencyPhone    string `json:"emergency_phone"`
	EmergencyRelation string `json:"emergency_relation"`
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, in UpdateProfileInput) (*models.User, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET
		  display_name=?, name_zh=?, name_en=?,
		  id_number=?, passport_number=?,
		  gender=?, birthday=?,
		  phone=?, shirt_size=?, food_type=?, address=?,
		  emergency_contact=?, emergency_phone=?, emergency_relation=?,
		  updated_at=?
		WHERE id=? AND deleted_at IS NULL`,
		nullStr(in.DisplayName), nullStr(in.NameZh), nullStr(in.NameEn),
		nullStr(in.IDNumber), nullStr(in.PassportNumber),
		in.Gender, nullStr(in.Birthday),
		nullStr(in.Phone), nullStr(in.ShirtSize), in.FoodType, nullStr(in.Address),
		nullStr(in.EmergencyContact), nullStr(in.EmergencyPhone), nullStr(in.EmergencyRelation),
		time.Now(), userID,
	)
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return s.FindByID(ctx, userID)
}

// ── Find helpers ─────────────────────────────────────────────

func (s *UserService) FindByID(ctx context.Context, id uint64) (*models.User, error) {
	row := s.db.QueryRowContext(ctx, userSelectSQL+"WHERE id=? AND deleted_at IS NULL", id)
	return scanUser(row)
}

func (s *UserService) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	row := s.db.QueryRowContext(ctx, userSelectSQL+"WHERE email=? AND deleted_at IS NULL", email)
	return scanUser(row)
}

const userSelectSQL = `
	SELECT id, uuid, username, email, phone, password_hash,
	       display_name, avatar_url, role, status, email_verified,
	       COALESCE(name_zh,''), COALESCE(name_en,''),
	       COALESCE(id_number,''), COALESCE(passport_number,''),
	       gender, birthday,
	       COALESCE(shirt_size,''), food_type,
	       COALESCE(address,''),
	       COALESCE(emergency_contact,''), COALESCE(emergency_phone,''),
	       COALESCE(emergency_relation,''),
	       created_at, updated_at
	FROM users `

func scanUser(row *sql.Row) (*models.User, error) {
	u := &models.User{}
	var (
		phone, displayName, avatarURL sql.NullString
		gender, foodType              sql.NullInt64
		birthday                      sql.NullString
	)
	err := row.Scan(
		&u.ID, &u.UUID, &u.Username, &u.Email,
		&phone, &u.PasswordHash,
		&displayName, &avatarURL,
		&u.Role, &u.Status, &u.EmailVerified,
		&u.NameZh, &u.NameEn,
		&u.IDNumber, &u.PassportNumber,
		&gender, &birthday,
		&u.ShirtSize, &foodType,
		&u.Address,
		&u.EmergencyContact, &u.EmergencyPhone, &u.EmergencyRelation,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.Phone       = phone.String
	u.DisplayName = displayName.String
	u.AvatarURL   = avatarURL.String
	u.Birthday    = birthday.String
	if gender.Valid {
		g := int(gender.Int64)
		u.Gender = &g
	}
	if foodType.Valid {
		f := int(foodType.Int64)
		u.FoodType = &f
	}
	return u, nil
}

// ── List users (admin) ────────────────────────────────────────

type ListUsersInput struct {
	Status   *int   `form:"status"`
	Role     *int   `form:"role"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ListUsersResult struct {
	Users []*models.User `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Pages int            `json:"pages"`
}

func (s *UserService) ListUsers(ctx context.Context, in ListUsersInput) (*ListUsersResult, error) {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.PageSize < 1 || in.PageSize > 100 {
		in.PageSize = 20
	}
	where := "WHERE deleted_at IS NULL"
	args := []any{}
	if in.Status != nil {
		where += " AND status=?"
		args = append(args, *in.Status)
	}
	if in.Role != nil {
		where += " AND role=?"
		args = append(args, *in.Role)
	}
	if in.Keyword != "" {
		where += " AND (username LIKE ? OR email LIKE ? OR display_name LIKE ? OR phone LIKE ? OR name_zh LIKE ?)"
		kw := "%" + in.Keyword + "%"
		args = append(args, kw, kw, kw, kw, kw)
	}

	var total int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users "+where, args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	offset := (in.Page - 1) * in.PageSize
	rows, err := s.db.QueryContext(ctx,
		"SELECT id,uuid,username,email,phone,password_hash,display_name,avatar_url,role,status,email_verified,"+
			"COALESCE(name_zh,''),COALESCE(name_en,''),COALESCE(id_number,''),COALESCE(passport_number,''),"+
			"gender,birthday,COALESCE(shirt_size,''),food_type,COALESCE(address,''),"+
			"COALESCE(emergency_contact,''),COALESCE(emergency_phone,''),COALESCE(emergency_relation,''),"+
			"created_at,updated_at "+
			"FROM users "+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
		append(args, in.PageSize, offset)...,
	)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		u := &models.User{}
		var phone, displayName, avatarURL, birthday sql.NullString
		var gender, foodType sql.NullInt64
		if err := rows.Scan(
			&u.ID, &u.UUID, &u.Username, &u.Email,
			&phone, &u.PasswordHash, &displayName, &avatarURL,
			&u.Role, &u.Status, &u.EmailVerified,
			&u.NameZh, &u.NameEn, &u.IDNumber, &u.PassportNumber,
			&gender, &birthday, &u.ShirtSize, &foodType,
			&u.Address, &u.EmergencyContact, &u.EmergencyPhone, &u.EmergencyRelation,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		u.Phone = phone.String
		u.DisplayName = displayName.String
		u.AvatarURL = avatarURL.String
		u.Birthday = birthday.String
		if gender.Valid {
			g := int(gender.Int64)
			u.Gender = &g
		}
		if foodType.Valid {
			f := int(foodType.Int64)
			u.FoodType = &f
		}
		users = append(users, u)
	}
	pages := (total + in.PageSize - 1) / in.PageSize
	return &ListUsersResult{Users: users, Total: total, Page: in.Page, Pages: pages}, nil
}

// ── Update status ─────────────────────────────────────────────

func (s *UserService) UpdateStatus(ctx context.Context, userID uint64, status, operatorRole int) error {
	if operatorRole < models.RoleAdmin {
		return ErrPermissionDenied
	}
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET status=?,updated_at=? WHERE id=? AND deleted_at IS NULL",
		status, time.Now(), userID,
	)
	return err
}

// ── Create admin ──────────────────────────────────────────────

type CreateAdminInput struct {
	Username    string `json:"username"     binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required"`
	Email       string `json:"email"        binding:"required,email"`
	Password    string `json:"password"     binding:"required,min=8"`
	Role        int    `json:"role"`
}

func (s *UserService) CreateAdmin(ctx context.Context, in CreateAdminInput, operatorRole int) (*models.User, error) {
	if operatorRole < models.RoleSuper {
		return nil, ErrPermissionDenied
	}
	role := models.RoleAdmin
	if in.Role == models.RoleSuper {
		role = models.RoleSuper
	}
	var count int
	_ = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE email=? AND deleted_at IS NULL", in.Email,
	).Scan(&count)
	if count > 0 {
		return nil, ErrEmailExists
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	uid := uuid.New().String()
	now := time.Now()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO users
		  (uuid,username,email,phone,password_hash,display_name,role,status,email_verified,created_at,updated_at)
		VALUES (?,?,?,'',?,?,?,1,1,?,?)`,
		uid, in.Username, in.Email, string(hash), in.DisplayName, role, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}
	id, _ := res.LastInsertId()
	return &models.User{
		ID: uint64(id), UUID: uid, Username: in.Username,
		Email: in.Email, DisplayName: in.DisplayName,
		Role: role, Status: models.StatusActive, CreatedAt: now,
	}, nil
}

// ── Change password ───────────────────────────────────────────

func (s *UserService) ChangePassword(ctx context.Context, userID uint64, oldPwd, newPwd string) error {
	user, err := s.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return ErrInvalidPassword
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET password_hash=?,updated_at=? WHERE id=?",
		string(hash), time.Now(), userID,
	)
	return err
}

// ── JWT ───────────────────────────────────────────────────────

func (s *UserService) generateTokenPair(u *models.User) (string, string, error) {
	now := time.Now()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", u.ID),
		"uuid": u.UUID, "role": u.Role,
		"iat": now.Unix(), "exp": now.Add(24 * time.Hour).Unix(),
	})
	accessStr, err := access.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": fmt.Sprintf("%d", u.ID), "type": "refresh",
		"iat": now.Unix(), "exp": now.Add(7 * 24 * time.Hour).Unix(),
	})
	refreshStr, err := refresh.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}
	return accessStr, refreshStr, nil
}

// ── helpers ───────────────────────────────────────────────────

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// ── Admin update user profile (不能改 email/username) ────────

type AdminUpdateProfileInput struct {
	DisplayName       string `json:"display_name"`
	NameZh            string `json:"name_zh"`
	NameEn            string `json:"name_en"`
	IDNumber          string `json:"id_number"`
	PassportNumber    string `json:"passport_number"`
	Gender            *int   `json:"gender"`
	Birthday          string `json:"birthday"`
	Phone             string `json:"phone"`
	ShirtSize         string `json:"shirt_size"`
	FoodType          *int   `json:"food_type"`
	Address           string `json:"address"`
	EmergencyContact  string `json:"emergency_contact"`
	EmergencyPhone    string `json:"emergency_phone"`
	EmergencyRelation string `json:"emergency_relation"`
}

// AdminUpdateProfile — 管理員修改任何人的資料（不含 email/username）
// operatorRole: 操作者角色；targetRole: 被修改者角色
// 一般管理員(8) 只能改 role<=8 的人；超級管理員(9) 不受限
func (s *UserService) AdminUpdateProfile(ctx context.Context, targetID uint64, in AdminUpdateProfileInput, operatorRole, targetRole int) error {
	if operatorRole < models.RoleAdmin {
		return ErrPermissionDenied
	}
	// 一般管理員不能修改超級管理員
	if operatorRole < models.RoleSuper && targetRole >= models.RoleSuper {
		return ErrPermissionDenied
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET
		  display_name=?, name_zh=?, name_en=?,
		  id_number=?, passport_number=?,
		  gender=?, birthday=?,
		  phone=?, shirt_size=?, food_type=?, address=?,
		  emergency_contact=?, emergency_phone=?, emergency_relation=?,
		  updated_at=?
		WHERE id=? AND deleted_at IS NULL`,
		nullStr(in.DisplayName), nullStr(in.NameZh), nullStr(in.NameEn),
		nullStr(in.IDNumber), nullStr(in.PassportNumber),
		in.Gender, nullStr(in.Birthday),
		nullStr(in.Phone), nullStr(in.ShirtSize), in.FoodType, nullStr(in.Address),
		nullStr(in.EmergencyContact), nullStr(in.EmergencyPhone), nullStr(in.EmergencyRelation),
		time.Now(), targetID,
	)
	return err
}

// AdminSetPassword — 管理員直接設定密碼（不需要舊密碼）
func (s *UserService) AdminSetPassword(ctx context.Context, targetID uint64, newPwd string, operatorRole, targetRole int) error {
	if operatorRole < models.RoleAdmin {
		return ErrPermissionDenied
	}
	if operatorRole < models.RoleSuper && targetRole >= models.RoleSuper {
		return ErrPermissionDenied
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET password_hash=?,updated_at=? WHERE id=? AND deleted_at IS NULL",
		string(hash), time.Now(), targetID,
	)
	return err
}

// CreateMember — 後台新增一般會員（直接 active）
type CreateMemberInput struct {
	Username    string `json:"username"     binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required"`
	Email       string `json:"email"        binding:"required,email"`
	Phone       string `json:"phone"        binding:"required"`
	Password    string `json:"password"     binding:"required,min=8"`
}

func (s *UserService) CreateMember(ctx context.Context, in CreateMemberInput, operatorRole int) (*models.User, error) {
	if operatorRole < models.RoleAdmin {
		return nil, ErrPermissionDenied
	}
	var count int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE email=? AND deleted_at IS NULL", in.Email,
	).Scan(&count); err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrEmailExists
	}
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE username=? AND deleted_at IS NULL", in.Username,
	).Scan(&count); err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUsernameExists
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	uid := uuid.New().String()
	now := time.Now()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO users
		  (uuid,username,email,phone,password_hash,display_name,role,status,email_verified,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,1,1,?,?)`,
		uid, in.Username, in.Email, in.Phone, string(hash), in.DisplayName,
		models.RoleMember, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create member: %w", err)
	}
	id, _ := res.LastInsertId()
	return &models.User{
		ID: uint64(id), UUID: uid, Username: in.Username,
		Email: in.Email, Phone: in.Phone, DisplayName: in.DisplayName,
		Role: models.RoleMember, Status: models.StatusActive, CreatedAt: now,
	}, nil
}

// ── ListAdmins — role >= 8 ────────────────────────────────────

func (s *UserService) ListAdmins(ctx context.Context, in ListUsersInput) (*ListUsersResult, error) {
	if in.Page < 1 { in.Page = 1 }
	if in.PageSize < 1 || in.PageSize > 100 { in.PageSize = 50 }

	where := "WHERE deleted_at IS NULL AND role >= 8"
	args := []any{}
	if in.Keyword != "" {
		where += " AND (username LIKE ? OR email LIKE ? OR display_name LIKE ?)"
		kw := "%" + in.Keyword + "%"
		args = append(args, kw, kw, kw)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admins: %w", err)
	}
	offset := (in.Page - 1) * in.PageSize
	rows, err := s.db.QueryContext(ctx,
		"SELECT id,uuid,username,email,phone,password_hash,display_name,avatar_url,role,status,email_verified,"+
			"COALESCE(name_zh,''),COALESCE(name_en,''),COALESCE(id_number,''),COALESCE(passport_number,''),"+
			"gender,birthday,COALESCE(shirt_size,''),food_type,COALESCE(address,''),"+
			"COALESCE(emergency_contact,''),COALESCE(emergency_phone,''),COALESCE(emergency_relation,''),"+
			"created_at,updated_at "+
			"FROM users "+where+" ORDER BY role DESC, created_at ASC LIMIT ? OFFSET ?",
		append(args, in.PageSize, offset)...,
	)
	if err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	defer rows.Close()
	users := []*models.User{}
	for rows.Next() {
		u := &models.User{}
		var phone, displayName, avatarURL, birthday sql.NullString
		var gender, foodType sql.NullInt64
		if err := rows.Scan(
			&u.ID, &u.UUID, &u.Username, &u.Email,
			&phone, &u.PasswordHash, &displayName, &avatarURL,
			&u.Role, &u.Status, &u.EmailVerified,
			&u.NameZh, &u.NameEn, &u.IDNumber, &u.PassportNumber,
			&gender, &birthday, &u.ShirtSize, &foodType,
			&u.Address, &u.EmergencyContact, &u.EmergencyPhone, &u.EmergencyRelation,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		u.Phone = phone.String; u.DisplayName = displayName.String
		u.AvatarURL = avatarURL.String; u.Birthday = birthday.String
		if gender.Valid { g := int(gender.Int64); u.Gender = &g }
		if foodType.Valid { f := int(foodType.Int64); u.FoodType = &f }
		users = append(users, u)
	}
	pages := (total + in.PageSize - 1) / in.PageSize
	return &ListUsersResult{Users: users, Total: total, Page: in.Page, Pages: pages}, nil
}

// ── SoftDelete ────────────────────────────────────────────────

func (s *UserService) SoftDelete(ctx context.Context, userID uint64) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET deleted_at=?, updated_at=? WHERE id=? AND deleted_at IS NULL",
		time.Now(), time.Now(), userID,
	)
	return err
}
