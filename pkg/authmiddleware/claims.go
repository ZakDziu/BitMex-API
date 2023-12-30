package authmiddleware

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"

	"bitmex-api/pkg/model"
)

const (
	AccessTokenTTL  = time.Hour * 8
	RefreshTokenTTL = time.Hour * 24 * 7
)

type BaseClaims struct {
	jwt.StandardClaims
	ID   uuid.UUID `json:"id"`
	Role model.UserRole
}

type AccessClaims struct {
	BaseClaims
	AccessUUID string `json:"access_uuid"`
}

type RefreshClaims struct {
	BaseClaims
	RefreshUUID string `json:"refresh_uuid"`
}

type Tokens struct {
	Access  string `json:"accessToken"`
	Refresh string `json:"refreshToken"`
}

func NewClaims(idClaims uuid.UUID, role model.UserRole, ttl time.Duration) BaseClaims {
	return BaseClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			Id:        uuid.NewV4().String(),
			IssuedAt:  time.Now().Unix(),
		},
		ID:   idClaims,
		Role: role,
	}
}

func GenerateClaims(idClaims uuid.UUID, role model.UserRole) (*AccessClaims, *RefreshClaims) {
	access := AccessClaims{
		BaseClaims: NewClaims(idClaims, role, AccessTokenTTL),
	}

	refresh := RefreshClaims{
		BaseClaims: NewClaims(idClaims, role, RefreshTokenTTL),
	}

	access.AccessUUID = refresh.Id
	refresh.RefreshUUID = access.Id

	return &access, &refresh
}
