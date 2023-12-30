package postgresstore_test

import (
	uuid "github.com/satori/go.uuid"

	"bitmex-api/pkg/model"
)

func (s *StoreSuite) TestAuthRepository_GetByUsername() {
	users := s.AuthUserFixture.List()

	for i := range users {
		err := s.store.DB.Create(&users[i]).Error
		s.Nil(err)

	}

	user, err := s.store.Auth().GetByUsername(users[0].Username)
	s.Nil(err)
	s.Equal(&users[0], user)
}

func (s *StoreSuite) TestAuthRepository_Create() {
	user := s.AuthUserFixture.One()

	err := s.store.Auth().Create(&user)
	s.Nil(err)

	var actualUser *model.AuthUser
	err = s.store.DB.Where("id=?", user.ID).Find(&actualUser).Error
	s.Nil(err)
	s.Equal(&user, actualUser)
}

func (s *StoreSuite) TestAuthRepository_Delete() {
	users := s.AuthUserFixture.List()

	for i := range users {
		err := s.store.DB.Create(&users[i]).Error
		s.Nil(err)

		err = s.store.DB.Create(&model.User{UserID: users[i].ID}).Error
		s.Nil(err)

	}

	err := s.store.Auth().Delete(users[0].ID)
	s.Nil(err)

	var actualUser *model.AuthUser

	result := s.store.DB.Where("id=?", users[0].ID).Find(&actualUser)
	s.Equal(int64(0), result.RowsAffected)
}

func (s *StoreSuite) TestAuthRepository_Get() {
	users := s.AuthUserFixture.List()

	for i := range users {
		err := s.store.DB.Create(&users[i]).Error
		s.Nil(err)

	}

	user, exists := s.store.Auth().Get(users[0].ID)
	s.Equal(true, exists)
	s.Equal(&users[0], user)

	_, exists = s.store.Auth().Get(uuid.NewV4())
	s.Equal(false, exists)
}

func (s *StoreSuite) TestAuthRepository_ChangePassword() {
	users := s.AuthUserFixture.List()

	for i := range users {
		err := s.store.DB.Create(&users[i]).Error
		s.Nil(err)

	}
	newPass := "newPass"
	err := s.store.Auth().ChangePassword(users[0].ID, newPass)
	s.Nil(err)

	var actualUser *model.AuthUser

	err = s.store.DB.Where("id=?", users[0].ID).Find(&actualUser).Error
	s.Nil(err)

	s.Equal(newPass, actualUser.Password)

}
