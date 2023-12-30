package store

import (
	uuid "github.com/satori/go.uuid"

	"bitmex-api/pkg/model"
)

type UserRepository interface {
	Get(id uuid.UUID) (*model.User, error)
	Update(user *model.User) error
	GetAll() ([]*model.User, error)
}

type AuthRepository interface {
	GetByUsername(username string) (*model.AuthUser, error)
	Get(id uuid.UUID) (*model.AuthUser, bool)
	Create(user *model.AuthUser) error
	Delete(id uuid.UUID) error
	ChangePassword(id uuid.UUID, pass string) error
}
