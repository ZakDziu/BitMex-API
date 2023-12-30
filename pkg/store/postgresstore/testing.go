package postgresstore

import (
	"bitmex-api/pkg/model"
	"bitmex-api/pkg/utils"
)

type FixtureAuthUser struct{}

func NewFixtureAuthUser() *FixtureAuthUser {
	return &FixtureAuthUser{}
}

func (f *FixtureAuthUser) One() model.AuthUser {
	return model.AuthUser{
		Username: "authUser",
		Password: "password",
		Role:     "ADMIN",
	}
}

func (f *FixtureAuthUser) List() []model.AuthUser {
	return []model.AuthUser{
		f.One(),
		utils.Mod(f.One(), func(v *model.AuthUser) {
			v.Username = "1"
		}),
		utils.Mod(f.One(), func(v *model.AuthUser) {
			v.Username = "2"
		}),
	}
}

type FixtureUser struct{}

func NewFixtureUser() *FixtureUser {
	return &FixtureUser{}
}

func (f *FixtureUser) One() model.User {
	return model.User{
		Name:    "userName",
		Surname: "Surname",
		Phone:   "+3801231231",
		Address: "Kiev",
	}
}

func (f *FixtureUser) List() []model.User {
	return []model.User{
		f.One(),
		utils.Mod(f.One(), func(v *model.User) {
			v.Name = "newName1"
		}),
		utils.Mod(f.One(), func(v *model.User) {
			v.Name = "newName2"
		}),
	}
}
