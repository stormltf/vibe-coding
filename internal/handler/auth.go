package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/internal/middleware"
	"github.com/test-tt/internal/service"
	"github.com/test-tt/pkg/errcode"
	"github.com/test-tt/pkg/logger"
	"github.com/test-tt/pkg/response"
	"github.com/test-tt/pkg/validate"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// RegisterRequest register request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

// Register godoc
// @Summary      User registration
// @Description  Register a new user account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Registration info"
// @Success      200      {object}  response.Response{data=object{user=model.User,token=string}}
// @Failure      400      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	if err := validate.Struct(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
		return
	}

	user, token, err := h.authService.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			response.Fail(c, errcode.ErrEmailAlreadyUsed)
		case errors.Is(err, service.ErrPasswordTooShort):
			response.Fail(c, errcode.ErrPasswordTooWeak.WithMessage("password must be at least 6 characters"))
		default:
			logger.ErrorCtxf(ctx, "failed to register user", "error", err)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.Success(c, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// LoginRequest login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login credentials"
// @Success      200      {object}  response.Response{data=object{user=model.User,token=string}}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	if err := validate.Struct(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
		return
	}

	user, token, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.Fail(c, errcode.ErrUserNotFound)
		case errors.Is(err, service.ErrInvalidPassword):
			response.Fail(c, errcode.ErrInvalidPassword)
		default:
			logger.ErrorCtxf(ctx, "failed to login", "error", err)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.Success(c, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// Logout godoc
// @Summary      User logout
// @Description  Invalidate current token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Security     Bearer
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(ctx context.Context, c *app.RequestContext) {
	authHeader := string(c.GetHeader("Authorization"))
	if authHeader == "" {
		response.Fail(c, errcode.ErrUnauthorized)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Fail(c, errcode.ErrUnauthorized)
		return
	}

	if err := h.authService.Logout(ctx, parts[1]); err != nil {
		logger.ErrorCtxf(ctx, "failed to logout", "error", err)
		response.Fail(c, errcode.ErrInternalServer)
		return
	}

	response.SuccessWithMessage(c, "logged out successfully", nil)
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Get the profile of currently authenticated user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response{data=model.User}
// @Failure      401  {object}  response.Response
// @Security     Bearer
// @Router       /auth/profile [get]
func (h *AuthHandler) GetProfile(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Fail(c, errcode.ErrLoginRequired)
		return
	}

	user, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Fail(c, errcode.ErrUserNotFound)
			return
		}
		logger.ErrorCtxf(ctx, "failed to get profile", "error", err)
		response.Fail(c, errcode.ErrDatabase)
		return
	}

	response.Success(c, user)
}

// UpdateProfileRequest update profile request
type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=50"`
	Age   int    `json:"age" validate:"omitempty,gte=0,lte=150"`
	Email string `json:"email" validate:"omitempty,email"`
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update the profile of currently authenticated user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateProfileRequest  true  "Profile info"
// @Success      200      {object}  response.Response{data=model.User}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Security     Bearer
// @Router       /auth/profile [put]
func (h *AuthHandler) UpdateProfile(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Fail(c, errcode.ErrLoginRequired)
		return
	}

	var req UpdateProfileRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	if err := validate.Struct(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
		return
	}

	user, err := h.authService.UpdateProfile(ctx, userID, req.Name, req.Age, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.Fail(c, errcode.ErrUserNotFound)
		case errors.Is(err, service.ErrEmailExists):
			response.Fail(c, errcode.ErrEmailAlreadyUsed)
		default:
			logger.ErrorCtxf(ctx, "failed to update profile", "userID", userID, "error", err)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.Success(c, user)
}

// ChangePasswordRequest change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=128"`
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Change the password of currently authenticated user
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      ChangePasswordRequest  true  "Password info"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Security     Bearer
// @Router       /auth/password [put]
func (h *AuthHandler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Fail(c, errcode.ErrLoginRequired)
		return
	}

	var req ChangePasswordRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	if err := validate.Struct(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
		return
	}

	if err := h.authService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.Fail(c, errcode.ErrUserNotFound)
		case errors.Is(err, service.ErrInvalidPassword):
			response.Fail(c, errcode.ErrInvalidPassword.WithMessage("current password is incorrect"))
		case errors.Is(err, service.ErrPasswordTooShort):
			response.Fail(c, errcode.ErrPasswordTooWeak.WithMessage("password must be at least 6 characters"))
		default:
			logger.ErrorCtxf(ctx, "failed to change password", "userID", userID, "error", err)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.SuccessWithMessage(c, "password changed successfully", nil)
}

// DeleteAccountRequest delete account request
type DeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

// DeleteAccount godoc
// @Summary      Delete account
// @Description  Delete the currently authenticated user's account
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      DeleteAccountRequest  true  "Password confirmation"
// @Success      200      {object}  response.Response
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Security     Bearer
// @Router       /auth/account [delete]
func (h *AuthHandler) DeleteAccount(ctx context.Context, c *app.RequestContext) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Fail(c, errcode.ErrLoginRequired)
		return
	}

	var req DeleteAccountRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	if err := validate.Struct(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams.WithMessage(validate.FirstError(err)))
		return
	}

	if err := h.authService.DeleteAccount(ctx, userID, req.Password); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			response.Fail(c, errcode.ErrUserNotFound)
		case errors.Is(err, service.ErrInvalidPassword):
			response.Fail(c, errcode.ErrInvalidPassword.WithMessage("password is incorrect"))
		default:
			logger.ErrorCtxf(ctx, "failed to delete account", "userID", userID, "error", err)
			response.Fail(c, errcode.ErrDatabase)
		}
		return
	}

	response.SuccessWithMessage(c, "account deleted successfully", nil)
}
