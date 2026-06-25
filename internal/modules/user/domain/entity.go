package domain

import "time"

type User struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Username    string     `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Email       string     `gorm:"uniqueIndex;size:128;not null" json:"email"`
	Password    string     `gorm:"size:256;not null" json:"-"`
	Nickname    string     `gorm:"size:64" json:"nickname"`
	Avatar      string     `gorm:"size:256" json:"avatar"`
	Gender      int        `gorm:"default:0" json:"gender"`
	Birthday    *time.Time `json:"birthday,omitempty"`
	Status      int        `gorm:"default:1" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP string     `gorm:"size:45" json:"last_login_ip,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (u *User) IsDisabled() bool {
	return u.Status == 0
}
