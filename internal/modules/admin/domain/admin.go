package domain

import "time"

// Admin 管理员账号（独立于 user 表）
type Admin struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Username    string     `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Password    string     `gorm:"size:256;not null" json:"-"`
	Nickname    string     `gorm:"size:64" json:"nickname"`
	Avatar      string     `gorm:"size:256" json:"avatar"`
	Status      int        `gorm:"default:1" json:"status"` // 1正常 0禁用
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (a *Admin) IsDisabled() bool {
	return a.Status == 0
}

func (a *Admin) TableName() string {
	return "admins"
}
