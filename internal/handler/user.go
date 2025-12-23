package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/internal/middleware"
	"github.com/test-tt/internal/model"
	"github.com/test-tt/internal/service"
	"github.com/test-tt/pkg/errcode"
	"github.com/test-tt/pkg/logger"
	"github.com/test-tt/pkg/pagination"
	"github.com/test-tt/pkg/response"
	"github.com/test-tt/pkg/validate"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

// GetUserByID godoc
// @Summary      获取用户详情
// @Description  根据用户ID获取用户详细信息
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  response.Response{data=model.User}
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidUserID)
		return
	}

	user, err := h.userService.GetByID(ctx, id)
	if err != nil {
		logger.ErrorCtxf(ctx, "failed to get user", "id", id, "error", err)
		response.Fail(c, errcode.ErrUserNotFound)
		return
	}

	response.Success(c, user)
}

// GetUsers godoc
// @Summary      获取用户列表
// @Description  分页获取用户列表，支持按名称或邮箱搜索
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "页码"      default(1)
// @Param        page_size  query     int     false  "每页数量"  default(10)
// @Param        keyword    query     string  false  "搜索关键词（名称或邮箱）"
// @Success      200  {object}  response.Response{data=pagination.PageResult}
// @Failure      500  {object}  response.Response
// @Router       /users [get]
func (h *UserHandler) GetUsers(ctx context.Context, c *app.RequestContext) {
	// 获取分页参数
	page := pagination.GetFromQuery(c)
	keyword := strings.TrimSpace(c.Query("keyword"))

	var users []model.User
	var total int64
	var err error

	if keyword != "" {
		users, total, err = h.userService.Search(ctx, keyword, page.Offset(), page.PageSize)
	} else {
		users, total, err = h.userService.GetPage(ctx, page.Offset(), page.PageSize)
	}

	if err != nil {
		logger.ErrorCtxf(ctx, "failed to get users", "error", err)
		response.Fail(c, errcode.ErrDatabase)
		return
	}

	result := pagination.NewPageResult(users, total, page.Page, page.PageSize)
	response.Success(c, result)
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=50" example:"张三"`
	Age   int    `json:"age" validate:"gte=0,lte=150" example:"25"`
	Email string `json:"email" validate:"omitempty,email" example:"zhangsan@example.com"`
}

// CreateUser godoc
// @Summary      创建用户
// @Description  创建新用户
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        request  body      CreateUserRequest  true  "用户信息"
// @Success      200      {object}  response.Response{data=model.User}
// @Failure      400      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Security     Bearer
// @Router       /users [post]
func (h *UserHandler) CreateUser(ctx context.Context, c *app.RequestContext) {
	var req CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	// 参数校验
	if err := validate.Struct(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
		return
	}

	user := &model.User{
		Name:  req.Name,
		Age:   req.Age,
		Email: req.Email,
	}

	if err := h.userService.Create(ctx, user); err != nil {
		logger.ErrorCtxf(ctx, "failed to create user", "error", err)
		response.Fail(c, errcode.ErrDatabase)
		return
	}

	response.Success(c, user)
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=50" example:"李四"`
	Age   int    `json:"age" validate:"gte=0,lte=150" example:"30"`
	Email string `json:"email" validate:"omitempty,email" example:"lisi@example.com"`
}

// UpdateUser godoc
// @Summary      更新用户
// @Description  更新用户信息（只能更新自己的信息）
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                true  "用户ID"
// @Param        request  body      UpdateUserRequest  true  "用户信息"
// @Success      200      {object}  response.Response{data=model.User}
// @Failure      400      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Security     Bearer
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidUserID)
		return
	}

	// 安全检查：只能更新自己的信息
	currentUserID := middleware.GetUserIDFromContext(c)
	if currentUserID == 0 {
		response.Fail(c, errcode.ErrUnauthorized)
		return
	}
	if id != currentUserID {
		response.Fail(c, errcode.ErrForbidden.WithMessage("can only update your own profile"))
		return
	}

	var req UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	// 参数校验
	if err := validate.Struct(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
		return
	}

	user := &model.User{
		ID:    id,
		Name:  req.Name,
		Age:   req.Age,
		Email: req.Email,
	}

	if err := h.userService.Update(ctx, user); err != nil {
		logger.ErrorCtxf(ctx, "failed to update user", "id", id, "error", err)
		response.Fail(c, errcode.ErrDatabase)
		return
	}

	response.Success(c, user)
}

// DeleteUser godoc
// @Summary      删除用户
// @Description  根据用户ID删除用户（只能删除自己的账号）
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     Bearer
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidUserID)
		return
	}

	// 安全检查：只能删除自己的账号
	currentUserID := middleware.GetUserIDFromContext(c)
	if currentUserID == 0 {
		response.Fail(c, errcode.ErrUnauthorized)
		return
	}
	if id != currentUserID {
		response.Fail(c, errcode.ErrForbidden.WithMessage("can only delete your own account"))
		return
	}

	if err := h.userService.Delete(ctx, id); err != nil {
		logger.ErrorCtxf(ctx, "failed to delete user", "id", id, "error", err)
		response.Fail(c, errcode.ErrDatabase)
		return
	}

	response.SuccessWithMessage(c, "user deleted", nil)
}
