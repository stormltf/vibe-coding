package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/test-tt/internal/dao"
	"github.com/test-tt/internal/model"
)

var (
	ErrProjectNotFound    = errors.New("project not found")
	ErrProjectNotOwned    = errors.New("project does not belong to user")
	ErrProjectNameEmpty   = errors.New("project name cannot be empty")
)

type ProjectService struct {
	projectDAO *dao.ProjectDAO
}

func NewProjectService() *ProjectService {
	return &ProjectService{
		projectDAO: dao.NewProjectDAO(),
	}
}

// GetByID retrieves a project by ID with ownership check
func (s *ProjectService) GetByID(ctx context.Context, id, userID uint64) (*model.Project, error) {
	project, err := s.projectDAO.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Check ownership
	if project.UserID != userID {
		return nil, ErrProjectNotOwned
	}

	return project, nil
}

// GetByUserID retrieves all projects for a user
func (s *ProjectService) GetByUserID(ctx context.Context, userID uint64) ([]model.Project, error) {
	return s.projectDAO.GetByUserID(ctx, userID)
}

// GetLatestByUserID retrieves the most recent project for a user
func (s *ProjectService) GetLatestByUserID(ctx context.Context, userID uint64) (*model.Project, error) {
	return s.projectDAO.GetLatestByUserID(ctx, userID)
}

// Create creates a new project
func (s *ProjectService) Create(ctx context.Context, userID uint64, name string) (*model.Project, error) {
	if name == "" {
		name = "New Project"
	}

	project := &model.Project{
		UserID:   userID,
		Name:     name,
		HTML:     "",
		CSS:      "",
		Messages: "[]", // Empty JSON array
	}

	if err := s.projectDAO.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// Update updates a project with ownership check
func (s *ProjectService) Update(ctx context.Context, id, userID uint64, name, html, css, messages string) (*model.Project, error) {
	// Check ownership
	project, err := s.projectDAO.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if project.UserID != userID {
		return nil, ErrProjectNotOwned
	}

	// Update fields
	if name != "" {
		project.Name = name
	}
	project.HTML = html
	project.CSS = css
	project.Messages = messages

	if err := s.projectDAO.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// Delete deletes a project with ownership check
func (s *ProjectService) Delete(ctx context.Context, id, userID uint64) error {
	// Check ownership
	exists, err := s.projectDAO.ExistsByIDAndUserID(ctx, id, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrProjectNotFound
	}

	return s.projectDAO.Delete(ctx, id)
}

// DeleteAllByUserID deletes all projects for a user (used when user deletes account)
func (s *ProjectService) DeleteAllByUserID(ctx context.Context, userID uint64) error {
	return s.projectDAO.DeleteByUserID(ctx, userID)
}
