package postgresstore

import (
	uuid "github.com/satori/go.uuid"

	"bitmex-api/pkg/model"
)

type AuthRepository struct {
	store *PostgresStore
}

func NewAuthRepository(store *PostgresStore) *AuthRepository {
	return &AuthRepository{store: store}
}

func (r *AuthRepository) GetByUsername(username string) (*model.AuthUser, error) {
	var authUser *model.AuthUser

	err := r.store.DB.Where("username=?", username).Find(&authUser).Error
	if err != nil {
		return nil, err
	}

	return authUser, nil
}

func (r *AuthRepository) Create(user *model.AuthUser) error {
	tx := r.store.DB.Begin()

	err := tx.Create(user).Error
	if err != nil {
		return err
	}

	err = tx.Create(&model.User{UserID: user.ID}).Error
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (r *AuthRepository) Delete(userID uuid.UUID) error {
	tx := r.store.DB.Begin()
	err := tx.Delete(&model.User{}, "user_id=?", userID).Error
	if err != nil {
		return err
	}

	err = tx.Delete(&model.AuthUser{}, "id=?", userID).Error
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func (r *AuthRepository) Get(id uuid.UUID) (*model.AuthUser, bool) {
	var user *model.AuthUser

	result := r.store.DB.Where("id=?", id).Find(&user)
	if result.RowsAffected == 0 {
		return nil, false
	}

	return user, true
}

func (r *AuthRepository) ChangePassword(id uuid.UUID, pass string) error {
	var user *model.AuthUser
	tx := r.store.DB.Begin()

	err := tx.Where("id=?", id).Find(&user).Error
	if err != nil {
		return err
	}

	user.Password = pass

	err = tx.Updates(user).Error
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}
