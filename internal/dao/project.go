package dao

import (
	"context"

	"gorm.io/gorm"

	"github.com/test-tt/internal/model"
	"github.com/test-tt/pkg/database"
)

type ProjectDAO struct{}

func NewProjectDAO() *ProjectDAO {
	return &ProjectDAO{}
}

// GetByID retrieves a project by ID
func (d *ProjectDAO) GetByID(ctx context.Context, id uint64) (*model.Project, error) {
	var project model.Project
	if err := database.DB.WithContext(ctx).First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByUserID retrieves all projects for a user, ordered by updated_at desc
func (d *ProjectDAO) GetByUserID(ctx context.Context, userID uint64) ([]model.Project, error) {
	var projects []model.Project
	if err := database.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// Create creates a new project
func (d *ProjectDAO) Create(ctx context.Context, project *model.Project) error {
	return database.DB.WithContext(ctx).Create(project).Error
}

// Update updates an existing project
func (d *ProjectDAO) Update(ctx context.Context, project *model.Project) error {
	return database.DB.WithContext(ctx).Save(project).Error
}

// UpdateFields updates specific fields of a project
func (d *ProjectDAO) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return database.DB.WithContext(ctx).Model(&model.Project{}).Where("id = ?", id).Updates(fields).Error
}

// Delete deletes a project by ID
func (d *ProjectDAO) Delete(ctx context.Context, id uint64) error {
	return database.DB.WithContext(ctx).Delete(&model.Project{}, id).Error
}

// DeleteByUserID deletes all projects for a user
func (d *ProjectDAO) DeleteByUserID(ctx context.Context, userID uint64) error {
	return database.DB.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Project{}).Error
}

// ExistsByIDAndUserID checks if a project exists and belongs to a user
func (d *ProjectDAO) ExistsByIDAndUserID(ctx context.Context, id, userID uint64) (bool, error) {
	var count int64
	err := database.DB.WithContext(ctx).Model(&model.Project{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&count).Error
	return count > 0, err
}

// GetLatestByUserID retrieves the most recently updated project for a user
func (d *ProjectDAO) GetLatestByUserID(ctx context.Context, userID uint64) (*model.Project, error) {
	var project model.Project
	if err := database.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}
