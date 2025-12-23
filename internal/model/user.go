package model

import "time"

// User 用户模型
// 索引说明:
// - idx_email: 邮箱唯一索引，用于登录和查重
// - idx_name: 名称索引，用于搜索
// - idx_created_at: 创建时间索引，用于分页排序
type User struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null;index:idx_name"`
	Age       int       `json:"age" gorm:"default:0"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null;uniqueIndex:idx_email"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"` // 密码不返回给前端
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// UserBasic 用户基础信息（用于列表查询，减少数据传输）
type UserBasic struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// UserListResult 分页结果（预分配切片容量）
type UserListResult struct {
	Users []User
	Total int64
}
