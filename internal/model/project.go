package model

import "time"

// Project represents a user's workspace project
type Project struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `json:"user_id" gorm:"index:idx_project_user_id;not null"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null;default:'New Project'"`
	HTML      string    `json:"html" gorm:"type:longtext"`
	CSS       string    `json:"css" gorm:"type:longtext"`
	Messages  string    `json:"messages" gorm:"type:longtext"` // JSON format chat history
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_project_created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for Project model
func (Project) TableName() string {
	return "projects"
}
