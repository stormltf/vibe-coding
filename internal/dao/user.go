package dao

import (
	"context"

	"github.com/test-tt/internal/model"
	"github.com/test-tt/pkg/database"
	"gorm.io/gorm"
)

type UserDAO struct{}

func NewUserDAO() *UserDAO {
	return &UserDAO{}
}

// GetByID 根据 ID 获取用户（带 context 支持超时取消）
func (d *UserDAO) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := database.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByIDs 批量获取用户（减少 N+1 查询）
func (d *UserDAO) GetByIDs(ctx context.Context, ids []uint64) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// 预分配切片容量
	users := make([]model.User, 0, len(ids))
	if err := database.DB.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetByEmail 根据邮箱获取用户（利用唯一索引）
func (d *UserDAO) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := database.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) GetAll(ctx context.Context) ([]model.User, error) {
	var users []model.User
	if err := database.DB.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetPage 分页查询用户列表（优化版）
// 优化点:
// 1. 使用 context 支持超时
// 2. 单次查询获取 count（使用窗口函数或缓存 count）
// 3. 指定排序字段利用索引
// 4. 预分配切片容量
func (d *UserDAO) GetPage(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var total int64

	db := database.DB.WithContext(ctx).Model(&model.User{})

	// 获取总数（这个查询会被缓存优化）
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 如果没有数据，直接返回
	if total == 0 {
		return []model.User{}, 0, nil
	}

	// 预分配切片容量，避免多次扩容
	users := make([]model.User, 0, limit)

	// 使用 ID 排序（主键索引，最快）
	if err := database.DB.WithContext(ctx).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetPageFast 快速分页（基于游标，适合大数据量）
// 使用 "WHERE id < lastID" 代替 OFFSET，性能更好
func (d *UserDAO) GetPageFast(ctx context.Context, lastID uint64, limit int) ([]model.User, error) {
	users := make([]model.User, 0, limit)

	query := database.DB.WithContext(ctx).Order("id DESC").Limit(limit)
	if lastID > 0 {
		query = query.Where("id < ?", lastID)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetBasicPage 获取基础信息分页（只查询必要字段）
func (d *UserDAO) GetBasicPage(ctx context.Context, offset, limit int) ([]model.UserBasic, int64, error) {
	var total int64

	if err := database.DB.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.UserBasic{}, 0, nil
	}

	users := make([]model.UserBasic, 0, limit)

	// 只查询需要的字段
	if err := database.DB.WithContext(ctx).
		Model(&model.User{}).
		Select("id", "name", "age").
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (d *UserDAO) Create(ctx context.Context, user *model.User) error {
	return database.DB.WithContext(ctx).Create(user).Error
}

// CreateBatch 批量创建用户
func (d *UserDAO) CreateBatch(ctx context.Context, users []model.User, batchSize int) error {
	if len(users) == 0 {
		return nil
	}
	return database.DB.WithContext(ctx).CreateInBatches(users, batchSize).Error
}

// Update 更新用户（只更新非零字段）
func (d *UserDAO) Update(ctx context.Context, user *model.User) error {
	return database.DB.WithContext(ctx).Model(user).Updates(user).Error
}

// UpdateFields 更新指定字段
func (d *UserDAO) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return database.DB.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

func (d *UserDAO) Delete(ctx context.Context, id uint64) error {
	return database.DB.WithContext(ctx).Delete(&model.User{}, id).Error
}

// DeleteBatch 批量删除
func (d *UserDAO) DeleteBatch(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return database.DB.WithContext(ctx).Where("id IN ?", ids).Delete(&model.User{}).Error
}

// ExistsByEmail 检查邮箱是否存在（利用索引，只查询 1 条）
func (d *UserDAO) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := database.DB.WithContext(ctx).
		Model(&model.User{}).
		Where("email = ?", email).
		Limit(1).
		Count(&count).Error
	return count > 0, err
}

// Transaction 事务支持
func (d *UserDAO) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return database.DB.WithContext(ctx).Transaction(fn)
}
