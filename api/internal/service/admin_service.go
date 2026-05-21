package service

import (
	"errors"
	"log"
	"sports-platform/internal/config"
	"sports-platform/internal/model"
	"sports-platform/internal/repository"
	jwtpkg "sports-platform/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	adminRepo *repository.AdminRepo
	cfg       *config.Config
}

func NewAdminService(adminRepo *repository.AdminRepo, cfg *config.Config) *AdminService {
	return &AdminService{adminRepo: adminRepo, cfg: cfg}
}

// SeedSuperAdmin - 首次啟動時建立超管（從 .env 讀取）
func (s *AdminService) SeedSuperAdmin() {
	exists, err := s.adminRepo.ExistsByUsername(s.cfg.AdminInitUsername)
	if err != nil || exists {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.AdminInitPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[SEED] Failed to hash admin password: %v", err)
		return
	}
	admin := &model.AdminUser{
		Username:     s.cfg.AdminInitUsername,
		Email:        s.cfg.AdminInitEmail,
		PasswordHash: string(hash),
		Name:         s.cfg.AdminInitName,
		Role:         2, // 超級管理員
	}
	if err := s.adminRepo.Create(admin); err != nil {
		log.Printf("[SEED] Failed to create super admin: %v", err)
		return
	}
	log.Printf("[SEED] Super admin created: %s", s.cfg.AdminInitUsername)
}

type AdminLoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminAuthResult struct {
	Token string           `json:"token"`
	Admin *model.AdminUser `json:"admin"`
}

func (s *AdminService) Login(input AdminLoginInput) (*AdminAuthResult, error) {
	admin, err := s.adminRepo.FindByUsername(input.Username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, errors.New("帳號或密碼錯誤")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("帳號或密碼錯誤")
	}
	s.adminRepo.UpdateLastLogin(admin.ID)

	// 載入權限
	if admin.Role == 1 {
		perms, _ := s.adminRepo.GetPermissions(admin.ID)
		admin.Permissions = perms
	}

	token, err := jwtpkg.GenerateAdmin(admin.ID, admin.Role, s.cfg.JWTSecret, 24)
	if err != nil {
		return nil, err
	}
	admin.PasswordHash = ""
	return &AdminAuthResult{Token: token, Admin: admin}, nil
}

type CreateAdminInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Role     int    `json:"role"` // 1 or 2
}

func (s *AdminService) CreateAdmin(input CreateAdminInput, creatorID uint64, creatorRole int) (*model.AdminUser, error) {
	// 只有超管能建立管理員
	if creatorRole < 2 {
		return nil, errors.New("權限不足")
	}
	// 不能建立比自己高的角色
	if input.Role > creatorRole {
		return nil, errors.New("無法建立比自己更高權限的管理員")
	}
	if input.Role < 1 {
		input.Role = 1
	}

	exists, _ := s.adminRepo.ExistsByUsername(input.Username)
	if exists {
		return nil, errors.New("帳號已存在")
	}
	existsEmail, _ := s.adminRepo.ExistsByEmail(input.Email)
	if existsEmail {
		return nil, errors.New("Email 已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	admin := &model.AdminUser{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
		Role:         input.Role,
		CreatedBy:    &creatorID,
	}
	if err := s.adminRepo.Create(admin); err != nil {
		return nil, err
	}
	admin.PasswordHash = ""
	return admin, nil
}

func (s *AdminService) ListAdmins(page, limit int) ([]model.AdminUser, int, error) {
	offset := (page - 1) * limit
	return s.adminRepo.List(limit, offset)
}

func (s *AdminService) GetPermissions(adminID uint64) (*model.AdminPermissions, error) {
	return s.adminRepo.GetPermissions(adminID)
}

func (s *AdminService) SetPermissions(creatorRole int, p *model.AdminPermissions) error {
	if creatorRole < 2 {
		return errors.New("權限不足，只有超級管理員可以設定權限")
	}
	target, err := s.adminRepo.FindByID(p.AdminID)
	if err != nil || target == nil {
		return errors.New("管理員不存在")
	}
	if target.Role >= 2 {
		return errors.New("超級管理員不需要設定個別權限")
	}
	return s.adminRepo.SetPermissions(p)
}

func (s *AdminService) SetStatus(creatorRole int, targetID uint64, status int) error {
	if creatorRole < 2 {
		return errors.New("權限不足")
	}
	target, err := s.adminRepo.FindByID(targetID)
	if err != nil || target == nil {
		return errors.New("管理員不存在")
	}
	// 保護：不能停用最後一個超管
	if target.Role == 2 && status == 0 {
		if s.adminRepo.CountSuperAdmins() <= 1 {
			return errors.New("至少需要保留一個啟用中的超級管理員")
		}
	}
	return s.adminRepo.UpdateStatus(targetID, status)
}
