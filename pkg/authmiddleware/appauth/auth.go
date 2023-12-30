package appauth

import (
	"crypto/ecdsa"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"bitmex-api/pkg/authmiddleware"
	"bitmex-api/pkg/logger"
	"bitmex-api/pkg/model"
	"bitmex-api/pkg/store"
)

const StringsNumber = 2

type AuthMiddleware struct {
	postgres *store.Store
	loc      *time.Location
	atKey    *ecdsa.PrivateKey
	rtKey    *ecdsa.PrivateKey
}

func NewAuthMiddleware(postgres *store.Store, atKey, rtKey *ecdsa.PrivateKey) *AuthMiddleware {
	loc, _ := time.LoadLocation("Europe/Moscow")
	var middleware = &AuthMiddleware{
		loc:      loc,
		postgres: postgres,
		atKey:    atKey,
		rtKey:    rtKey,
	}

	return middleware
}

//nolint:varnamelen
func (m *AuthMiddleware) Authorize(c *gin.Context) {
	tokenString := m.ExtractToken(c.Request)
	claims, err := m.Validate(tokenString)
	if err != nil {
		logger.Errorf("Authorize.Validate", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrRefreshExpired)

		return
	}

	if claims == nil {
		logger.Errorf("Authorize.Not nil claims", "empty claims")
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrUnauthorized)

		return
	}

	userDB, exists := m.postgres.Auth.Get(claims.BaseClaims.ID)
	if !exists {
		logger.Errorf("Authorize.Get", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrRefreshExpired)

		return
	}

	if claims.BaseClaims.Role != userDB.Role {
		logger.Errorf("Authorize.User role", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrRefreshExpired)

		return
	}

	c.Next()
}

func (m *AuthMiddleware) GetUserID(accessToken string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(accessToken, &authmiddleware.AccessClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				logger.Errorf("GetUserId.unexpected signing method: ", token.Header["alg"])

				return nil, model.ErrUnauthorized
			}

			return &m.atKey.PublicKey, nil
		})
	if err != nil {
		return uuid.Nil, model.ErrUnauthorized
	}

	claims, ok := token.Claims.(*authmiddleware.AccessClaims)
	if !ok {
		logger.Errorf("Refresh.invalid token claims: ", token.Claims)

		return uuid.Nil, model.ErrUnauthorized
	}

	if !token.Valid {
		return uuid.Nil, model.ErrUnauthorized
	}

	return claims.BaseClaims.ID, nil
}

func (m *AuthMiddleware) GetUserRole(accessToken string) (model.UserRole, error) {
	// Create new pairs of refresh and access tokens
	token, err := jwt.ParseWithClaims(accessToken, &authmiddleware.AccessClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				logger.Errorf("GetUserId.unexpected signing method", token.Header["alg"])

				return nil, model.ErrUnauthorized
			}

			return &m.atKey.PublicKey, nil
		})
	if err != nil {
		return "", model.ErrUnauthorized
	}

	claims, ok := token.Claims.(*authmiddleware.AccessClaims)
	if !ok {
		logger.Errorf("Refresh.invalid token claims", token.Claims)

		return "", model.ErrUnauthorized
	}

	if !token.Valid {
		return "", model.ErrUnauthorized
	}

	return claims.BaseClaims.Role, nil
}

func (m *AuthMiddleware) CreateTokens(id uuid.UUID, role model.UserRole) (*authmiddleware.Tokens, error) {
	accessClaims, refreshClaims := authmiddleware.GenerateClaims(id, role)

	at := jwt.NewWithClaims(jwt.SigningMethodES256, accessClaims)
	accessToken, err := at.SignedString(m.atKey)
	if err != nil {
		return nil, err
	}

	rt := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims)
	refreshToken, err := rt.SignedString(m.rtKey)
	if err != nil {
		return nil, err
	}

	return &authmiddleware.Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (m *AuthMiddleware) Refresh(tokens authmiddleware.Tokens) (*authmiddleware.Tokens, error) {
	token, err := jwt.ParseWithClaims(tokens.Refresh, &authmiddleware.RefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				logger.Errorf("Refresh.unexpected signing method", token.Header["alg"])

				return nil, model.ErrUnauthorized
			}

			return &m.rtKey.PublicKey, nil
		})
	if err != nil {
		return nil, model.ErrUnauthorized
	}

	claims, ok := token.Claims.(*authmiddleware.RefreshClaims)
	if !ok {
		logger.Errorf("Refresh.invalid token claims", token.Claims)

		return nil, model.ErrUnauthorized
	}

	if !token.Valid {
		return nil, model.ErrUnauthorized
	}

	userDB, exists := m.postgres.Auth.Get(claims.BaseClaims.ID)
	if !exists {
		logger.Errorf("Refresh.Get", err)

		return nil, model.ErrUnauthorized
	}

	return m.CreateTokens(claims.ID, userDB.Role)
}

func (m *AuthMiddleware) ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == StringsNumber {
		return strArr[1]
	}

	return ""
}

// Validate verifies token signature.
func (m *AuthMiddleware) Validate(raw string) (*authmiddleware.AccessClaims, error) {
	token, err := jwt.ParseWithClaims(raw, &authmiddleware.AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			logger.Errorf("Validate.unexpected signing method", token.Header["alg"])

			return nil, model.ErrUnauthorized
		}

		return &m.atKey.PublicKey, nil
	})

	if err != nil {
		return nil, model.ErrUnauthorized
	}

	claims, ok := token.Claims.(*authmiddleware.AccessClaims)
	if !ok {
		logger.Errorf("Validate.invalid token claims", token.Claims)

		return nil, model.ErrUnauthorized
	}

	if !token.Valid {
		return nil, model.ErrUnauthorized
	}

	return claims, nil
}
