package domain

import "time"

// OperationLog 管理员操作日志
type OperationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AdminID    uint      `gorm:"not null;index" json:"admin_id"`
	Method     string    `gorm:"size:10;not null" json:"method"`
	Path       string    `gorm:"size:512;not null" json:"path"`
	StatusCode int       `gorm:"default:0" json:"status_code"`
	ClientIP   string    `gorm:"size:45" json:"client_ip"`
	CreatedAt  time.Time `json:"created_at"`
}
