package model

import "strings"

type ChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (c *ChangePassword) IsValid() bool {
	oldPass := strings.TrimSpace(c.OldPassword)
	newPass := strings.TrimSpace(c.NewPassword)
	if oldPass == "" || newPass == "" {
		return false
	}

	c.OldPassword = oldPass
	c.NewPassword = newPass

	return true
}
