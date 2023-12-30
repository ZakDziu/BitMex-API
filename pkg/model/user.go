package model

import (
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                  uuid.UUID      `json:"-"`
	UserID              uuid.UUID      `json:"-"`
	Name                string         `json:"name"`
	Surname             string         `json:"surname"`
	Phone               string         `json:"phone"`
	Address             string         `json:"address"`
	Subscription        bool           `json:"subscription"`
	SubscriptionSymbols pq.StringArray `gorm:"type:text[]"  json:"subscriptionSymbols"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	uuid := uuid.NewV4().String()
	tx.Statement.SetColumn("ID", uuid)

	return nil
}

func (u *User) TableName() string {
	return "users"
}
