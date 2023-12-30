package model

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"strings"
)

type UserRole string

const (
	AdminUserRole UserRole = "ADMIN"
	BaseUserRole  UserRole = "BASE"
)

type AuthUser struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;" json:"-"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Role     UserRole  `json:"role"`
}

func (a *AuthUser) TableName() string {
	return "auth_users"
}

func (a *AuthUser) BeforeCreate(tx *gorm.DB) error {
	uuid := uuid.NewV4().String()
	tx.Statement.SetColumn("ID", uuid)

	return nil
}

func (a *AuthUser) IsValid(forRegister bool) bool {
	username := strings.TrimSpace(a.Username)
	pass := strings.TrimSpace(a.Password)
	if username == "" || pass == "" || (forRegister && a.Role == "") {
		return false
	}

	a.Username = username
	a.Password = pass

	return true
}
