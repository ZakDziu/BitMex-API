package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"bitmex-api/pkg/authmiddleware"
	"bitmex-api/pkg/logger"
	"bitmex-api/pkg/model"
	"bitmex-api/pkg/model/ui/auth"
)

type AuthHandler struct {
	api *api
}

func NewAuthHandler(a *api) *AuthHandler {
	return &AuthHandler{
		api: a,
	}
}

// Login
// @Summary user login
// @Produce json
// @Tags Auth
// @Param userInfo  body model.AuthUser  true "User Info"
// @Success 200 {object} authmiddleware.Tokens
// @Failure 400 {object} errors.UIResponseErrorBadRequest
// @Router /api/v1/login [post]
//
//nolint:varnamelen
func (h *AuthHandler) Login(c *gin.Context) {
	user := &model.AuthUser{}
	err := c.ShouldBindJSON(&user)
	if err != nil {
		logger.Errorf("Login.ShouldBindJSON", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	if !user.IsValid(false) {
		logger.Errorf("Login.Empty login or pass", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	userDB, err := h.api.postgresStore.Auth.GetByUsername(user.Username)
	if err != nil {
		logger.Errorf("Login.Get", err)
		if err.Error() == model.NotFound {
			c.JSON(http.StatusUnauthorized, model.ErrUnauthorized)
		} else {
			c.JSON(http.StatusInternalServerError, model.ErrUnhealthy)
		}

		return
	}

	if !authmiddleware.IsPasswordMatch(user.Password, userDB.Password) {
		logger.Errorf("Login.IsPasswordMatch", err)
		c.JSON(http.StatusUnauthorized, model.ErrUnauthorized)

		return
	}

	tokens, err := h.api.auth.CreateTokens(userDB.ID, userDB.Role)
	if err != nil {
		logger.Errorf("Login.CreateTokens", err)
		c.JSON(http.StatusBadRequest, model.ErrUnhealthy)

		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Register
// @Summary user registration
// @Description available user roles: ADMIN/BASE
// @Produce json
// @Tags Auth
// @Security ApiKeyAuth
// @Param userInfo  body model.AuthUser  true "User Info"
// @Success 200 {object} auth.RegistrationResponse
// @Failure 400 {object} errors.UIResponseErrorBadRequest
// @Router /api/v1/registration [post]
//
//nolint:varnamelen
func (h *AuthHandler) Register(c *gin.Context) {
	user := &model.AuthUser{}
	err := c.ShouldBindJSON(&user)
	if err != nil {
		logger.Errorf("Register.ShouldBindJSON", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	userRole, err := h.getUserRoleFromHeader(c)
	if err != nil {
		logger.Errorf("Register.getUserRoleFromHeader", err)
		c.JSON(http.StatusUnauthorized, model.ErrUnauthorized)

		return
	}

	if userRole != model.AdminUserRole {
		logger.Errorf("Register.GetUserRole", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidRole)

		return
	}

	if !user.IsValid(true) {
		logger.Errorf("Register.Empty username or pass", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	user.Password = authmiddleware.CreateHashPassword(user.Password)

	userDB, err := h.api.postgresStore.Auth.GetByUsername(user.Username)
	if err != nil {
		logger.Errorf("Register.GetByUsername", err)
		c.JSON(http.StatusBadRequest, model.ErrUnhealthy)

		return
	}

	if userDB.ID != uuid.Nil {
		logger.Errorf("Register.Username exist", err)
		c.JSON(http.StatusBadRequest, model.ErrUsernameExist)

		return
	}

	err = h.api.postgresStore.Auth.Create(user)
	if err != nil {
		logger.Errorf("Register.Create", err)
		c.JSON(http.StatusInternalServerError, model.ErrUnhealthy)

		return
	}

	c.JSON(http.StatusOK, auth.RegistrationResponse{Status: "user created"})
}

// Refresh
// @Summary user refresh token
// @Produce json
// @Tags Auth
// @Param token  body authmiddleware.Tokens  true "Tokens"
// @Success 200 {object} authmiddleware.Tokens
// @Failure 400 {object} errors.UIResponseErrorBadRequest
// @Router /api/v1/refresh [post]
//
//nolint:varnamelen
func (h *AuthHandler) Refresh(c *gin.Context) {
	oldTokens := authmiddleware.Tokens{}
	err := c.ShouldBindJSON(&oldTokens)
	if err != nil {
		logger.Errorf("Refresh.ShouldBindJSON", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	newTokens, err := h.api.auth.Refresh(oldTokens)
	if err != nil {
		logger.Errorf("Refresh.Refresh", err)
		c.JSON(http.StatusUnauthorized, model.ErrUnauthorized)

		return
	}

	c.JSON(http.StatusOK, newTokens)
}

// ChangePassword
// @Summary user change password
// @Produce json
// @Tags Auth
// @Param ChangePassword  body model.ChangePassword  true "Change Password"
// @Success 200 {object} authmiddleware.Tokens
// @Failure 400 {object} errors.UIResponseErrorBadRequest
// @Router /api/v1/change-password [patch]
//
//nolint:varnamelen
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	changePass := &model.ChangePassword{}
	err := c.ShouldBindJSON(&changePass)
	if err != nil {
		logger.Errorf("ChangePassword.ShouldBindJSON", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	if !changePass.IsValid() {
		logger.Errorf("ChangePassword.Empty old or new pass", err)
		c.JSON(http.StatusBadRequest, model.ErrInvalidBody)

		return
	}

	userID, err := h.getUserIDFromHeader(c)
	if err != nil {
		logger.Errorf("Register.getUserRoleFromHeader", err)
		c.JSON(http.StatusUnauthorized, model.ErrUnauthorized)

		return
	}

	userDB, exists := h.api.postgresStore.Auth.Get(userID)
	if !exists {
		logger.Errorf("ChangePassword.UserNotExist", err)
		c.JSON(http.StatusUnauthorized, model.ErrRefreshExpired)

		return
	}

	if !authmiddleware.IsPasswordMatch(changePass.OldPassword, userDB.Password) {
		logger.Errorf("ChangePassword.IsPasswordMatch", nil)

		c.JSON(http.StatusUnauthorized, model.ErrUnauthorized)

		return
	}

	changePass.NewPassword = authmiddleware.CreateHashPassword(changePass.NewPassword)

	err = h.api.postgresStore.Auth.ChangePassword(userID, changePass.NewPassword)
	if err != nil {
		logger.Errorf("ChangePassword.ChangePassword", err)
		c.JSON(http.StatusInternalServerError, model.ErrUnhealthy)

		return
	}

	tokens, err := h.api.auth.CreateTokens(userDB.ID, userDB.Role)
	if err != nil {
		logger.Errorf("ChangePassword.CreateTokens", err)
		c.JSON(http.StatusInternalServerError, model.ErrUnhealthy)

		return
	}

	c.JSON(http.StatusOK, tokens)
}

//nolint:gocritic
func (h *AuthHandler) getUserIDFromHeader(c *gin.Context) (uuid.UUID, error) {
	token := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", -1)
	userID, err := h.api.auth.GetUserID(token)

	return userID, err
}

//nolint:gocritic
func (h *AuthHandler) getUserRoleFromHeader(c *gin.Context) (model.UserRole, error) {
	token := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", -1)
	userRole, err := h.api.auth.GetUserRole(token)

	return userRole, err
}
