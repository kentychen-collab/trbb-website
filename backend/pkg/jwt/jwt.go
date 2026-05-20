package jwt

import (
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	MemberID uint64 `json:"member_id,omitempty"`
	AdminID  uint64 `json:"admin_id,omitempty"`
	Role     int    `json:"role"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
	gojwt.RegisteredClaims
}

func Generate(memberID uint64, role int, secret string, expireHours int) (string, error) {
	claims := Claims{
		MemberID: memberID,
		Role:     role,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateAdmin(adminID uint64, role int, secret string, expireHours int) (string, error) {
	claims := Claims{
		AdminID: adminID,
		Role:    role,
		IsAdmin: true,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
