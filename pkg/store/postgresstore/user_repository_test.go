package postgresstore_test

import "bitmex-api/pkg/model"

func (s *StoreSuite) TestUserRepository_Get() {
	users := s.UserFixture.List()

	for i := range users {
		authUser := &model.AuthUser{}
		err := s.store.DB.Create(&authUser).Error
		s.Nil(err)
		users[i].UserID = authUser.ID
		err = s.store.DB.Create(&users[i]).Error
		s.Nil(err)
	}

	user, err := s.store.User().Get(users[0].UserID)
	s.Nil(err)
	s.Equal(&users[0], user)
}

func (s *StoreSuite) TestUserRepository_Update() {
	users := s.UserFixture.List()

	for i := range users {
		authUser := &model.AuthUser{}
		err := s.store.DB.Create(&authUser).Error
		s.Nil(err)
		users[i].UserID = authUser.ID
		err = s.store.DB.Create(&users[i]).Error
		s.Nil(err)
	}

	users[0].Name = "editedName"
	users[0].Address = "editedAddress"

	err := s.store.User().Update(&users[0])
	s.Nil(err)

	var actualUser *model.User
	err = s.store.DB.Where("id=?", users[0].ID).Find(&actualUser).Error
	s.Nil(err)

	s.Equal(&users[0], actualUser)
}
