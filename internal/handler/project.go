package handler

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/internal/middleware"
	"github.com/test-tt/internal/service"
	"github.com/test-tt/pkg/errcode"
	"github.com/test-tt/pkg/logger"
	"github.com/test-tt/pkg/response"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler() *ProjectHandler {
	return &ProjectHandler{
		projectService: service.NewProjectService(),
	}
}

// CreateProjectRequest create project request
type CreateProjectRequest struct {
	Name string `json:"name"`
}

// UpdateProjectRequest update project request
type UpdateProjectRequest struct {
	Name     string `json:"name"`
	HTML     string `json:"html"`
	CSS      string `json:"css"`
	Messages string `json:"messages"`
}

// List godoc
// @Summary      List user projects
// @Description  Get all projects for the authenticated user
// @Tags         Projects
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.Response{data=[]model.Project}
// @Failure      401  {object}  response.Response
// @Router       /projects [get]
func (h *ProjectHandler) List(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserIDFromContext(c)

	projects, err := h.projectService.GetByUserID(ctx, userID)
	if err != nil {
		logger.ErrorCtxf(ctx, "failed to list projects", "error", err, "userID", userID)
		response.Fail(c, errcode.ErrDatabase)
		return
	}

	response.Success(c, projects)
}

// Get godoc
// @Summary      Get project
// @Description  Get a specific project by ID
// @Tags         Projects
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Project ID"
// @Success      200  {object}  response.Response{data=model.Project}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /projects/{id} [get]
func (h *ProjectHandler) Get(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserIDFromContext(c)
	projectID, _ := c.Params.Get("id")

	var id uint64
	if _, err := parseUint64(projectID, &id); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	project, err := h.projectService.GetByID(ctx, id, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNotFound):
			response.Fail(c, errcode.ErrNotFound.WithMessage("project not found"))
		case errors.Is(err, service.ErrProjectNotOwned):
			response.Fail(c, errcode.ErrForbidden.WithMessage("project does not belong to you"))
		default:
			logger.ErrorCtxf(ctx, "failed to get project", "error", err, "projectID", id)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.Success(c, project)
}

// Create godoc
// @Summary      Create project
// @Description  Create a new project
// @Tags         Projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      CreateProjectRequest  true  "Project info"
// @Success      200      {object}  response.Response{data=model.Project}
// @Failure      401      {object}  response.Response
// @Router       /projects [post]
func (h *ProjectHandler) Create(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserIDFromContext(c)

	var req CreateProjectRequest
	if err := c.BindJSON(&req); err != nil {
		// If no body, use default name
		req.Name = ""
	}

	project, err := h.projectService.Create(ctx, userID, req.Name)
	if err != nil {
		logger.ErrorCtxf(ctx, "failed to create project", "error", err, "userID", userID)
		response.Fail(c, errcode.ErrDatabase)
		return
	}

	response.Success(c, project)
}

// Update godoc
// @Summary      Update project
// @Description  Update an existing project
// @Tags         Projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                   true  "Project ID"
// @Param        request  body      UpdateProjectRequest  true  "Project data"
// @Success      200      {object}  response.Response{data=model.Project}
// @Failure      401      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Router       /projects/{id} [put]
func (h *ProjectHandler) Update(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserIDFromContext(c)
	projectID, _ := c.Params.Get("id")

	var id uint64
	if _, err := parseUint64(projectID, &id); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	var req UpdateProjectRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	project, err := h.projectService.Update(ctx, id, userID, req.Name, req.HTML, req.CSS, req.Messages)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNotFound):
			response.Fail(c, errcode.ErrNotFound.WithMessage("project not found"))
		case errors.Is(err, service.ErrProjectNotOwned):
			response.Fail(c, errcode.ErrForbidden.WithMessage("project does not belong to you"))
		default:
			logger.ErrorCtxf(ctx, "failed to update project", "error", err, "projectID", id)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.Success(c, project)
}

// Delete godoc
// @Summary      Delete project
// @Description  Delete a project
// @Tags         Projects
// @Security     BearerAuth
// @Param        id   path      int  true  "Project ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /projects/{id} [delete]
func (h *ProjectHandler) Delete(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserIDFromContext(c)
	projectID, _ := c.Params.Get("id")

	var id uint64
	if _, err := parseUint64(projectID, &id); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	err := h.projectService.Delete(ctx, id, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNotFound):
			response.Fail(c, errcode.ErrNotFound.WithMessage("project not found"))
		case errors.Is(err, service.ErrProjectNotOwned):
			response.Fail(c, errcode.ErrForbidden.WithMessage("project does not belong to you"))
		default:
			logger.ErrorCtxf(ctx, "failed to delete project", "error", err, "projectID", id)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.Success(c, nil)
}

// Helper function to parse uint64
func parseUint64(s string, result *uint64) (bool, error) {
	var id uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, errors.New("invalid number")
		}
		id = id*10 + uint64(c-'0')
	}
	*result = id
	return true, nil
}
