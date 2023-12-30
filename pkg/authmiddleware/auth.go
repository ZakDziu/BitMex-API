package authmiddleware

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/sha3"

	"bitmex-api/pkg/logger"
	"bitmex-api/pkg/model"
)

const (
	AuthSalt = "crm-system"
	MaxAge   = 32000000
)

type AuthMiddleware interface {
	Authorize(c *gin.Context)
	CreateTokens(id uuid.UUID, role model.UserRole) (*Tokens, error)
	Refresh(tokens Tokens) (*Tokens, error)
	ExtractToken(r *http.Request) string
	Validate(raw string) (*AccessClaims, error)
	GetUserRole(accessToken string) (model.UserRole, error)
	GetUserID(accessToken string) (uuid.UUID, error)
}

func IsPasswordMatch(password, hashedPassword string) bool {
	userPasswordHash := H3hash(password + AuthSalt)

	return userPasswordHash == hashedPassword
}

func CreateHashPassword(password string) string {
	return H3hash(password + AuthSalt)
}

func H3hash(s string) string {
	h3 := sha3.New512()
	if _, err := io.WriteString(h3, s); err != nil {
		logger.Errorf("H3hash.WriteString", err)
	}

	return fmt.Sprintf("%x", h3.Sum(nil))
}
