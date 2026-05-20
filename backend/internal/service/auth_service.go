package service

import (
	"errors"
	"sports-platform/internal/model"
	"sports-platform/internal/repository"
	jwtpkg "sports-platform/pkg/jwt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	memberRepo *repository.MemberRepo
	jwtSecret  string
	expireHours int
}

func NewAuthService(memberRepo *repository.MemberRepo, jwtSecret string, expireHours int) *AuthService {
	return &AuthService{memberRepo: memberRepo, jwtSecret: jwtSecret, expireHours: expireHours}
}

type RegisterInput struct {
	MemberNo string `json:"member_no" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResult struct {
	Token  string        `json:"token"`
	Member *model.Member `json:"member"`
}

func (s *AuthService) Register(input RegisterInput) (*AuthResult, error) {
	// 檢查會員編號唯一
	exists, err := s.memberRepo.ExistsByMemberNo(input.MemberNo)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("會員編號已被使用")
	}

	existing, err := s.memberRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("Email 已被註冊")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	m := &model.Member{
		UUID:         uuid.New().String(),
		MemberNo:     input.MemberNo,
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
		Phone:        input.Phone,
		Status:       model.MemberStatusPending, // 待審核
		CreatedAt:    time.Now(),
	}
	if err := s.memberRepo.Create(m); err != nil {
		return nil, err
	}

	// 回傳不含 token（待審核不能登入）
	m.PasswordHash = ""
	return &AuthResult{Token: "", Member: m}, nil
}

func (s *AuthService) Login(input LoginInput, ip, userAgent string) (*AuthResult, error) {
	m, err := s.memberRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("帳號或密碼錯誤")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(m.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("帳號或密碼錯誤")
	}

	// 檢查狀態
	switch m.Status {
	case model.MemberStatusPending:
		return nil, errors.New("帳號尚未審核，請等待管理員核準")
	case model.MemberStatusDisabled:
		return nil, errors.New("帳號已停用，請聯繫管理員")
	}

	s.memberRepo.LogLogin(m.ID, ip, userAgent)

	token, err := jwtpkg.Generate(m.ID, m.Role, s.jwtSecret, s.expireHours)
	if err != nil {
		return nil, err
	}
	m.PasswordHash = ""
	return &AuthResult{Token: token, Member: m}, nil
}

func (s *AuthService) ChangePassword(memberID uint64, oldPass, newPass string) error {
	m, err := s.memberRepo.FindByID(memberID)
	if err != nil || m == nil {
		return errors.New("會員不存在")
	}
	mWithHash, err := s.memberRepo.FindByEmail(m.Email)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(mWithHash.PasswordHash), []byte(oldPass)); err != nil {
		return errors.New("目前密碼不正確")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.memberRepo.UpdatePassword(memberID, string(hash))
}
