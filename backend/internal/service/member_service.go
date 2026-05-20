package service

import (
	"sports-platform/internal/model"
	"sports-platform/internal/repository"
	"time"
)

type MemberService struct {
	memberRepo *repository.MemberRepo
}

func NewMemberService(memberRepo *repository.MemberRepo) *MemberService {
	return &MemberService{memberRepo: memberRepo}
}

type UpdateProfileInput struct {
	Name      string  `json:"name" binding:"required"`
	Phone     string  `json:"phone"`
	AvatarURL string  `json:"avatar_url"`
	Gender    int     `json:"gender"`
	Birthday  *string `json:"birthday"` // "2000-01-01"
}

func (s *MemberService) GetProfile(memberID uint64) (*model.Member, error) {
	return s.memberRepo.FindByID(memberID)
}

func (s *MemberService) UpdateProfile(memberID uint64, input UpdateProfileInput) (*model.Member, error) {
	var birthday *time.Time
	if input.Birthday != nil && *input.Birthday != "" {
		t, err := time.Parse("2006-01-02", *input.Birthday)
		if err == nil {
			birthday = &t
		}
	}
	if err := s.memberRepo.UpdateProfile(memberID, input.Name, input.Phone, input.AvatarURL, input.Gender, birthday); err != nil {
		return nil, err
	}
	return s.memberRepo.FindByID(memberID)
}

func (s *MemberService) ListAddresses(memberID uint64) ([]model.MemberAddress, error) {
	return s.memberRepo.ListAddresses(memberID)
}

func (s *MemberService) CreateAddress(memberID uint64, a model.MemberAddress) error {
	a.MemberID = memberID
	return s.memberRepo.CreateAddress(&a)
}

func (s *MemberService) UpdateAddress(memberID uint64, a model.MemberAddress) error {
	a.MemberID = memberID
	return s.memberRepo.UpdateAddress(&a)
}

func (s *MemberService) DeleteAddress(id, memberID uint64) error {
	return s.memberRepo.DeleteAddress(id, memberID)
}
