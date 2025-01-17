package handler

import (
	"xledger/internal/http/handler/dto"
	"xledger/internal/http/service"

	"github.com/RichXan/xcommon/xerror"
	"github.com/RichXan/xcommon/xhttp"
	"github.com/RichXan/xcommon/xlog"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	logger      *xlog.Logger
	userService service.UserService
}

func NewUserHandler(logger *xlog.Logger, userService service.UserService) *UserHandler {
	return &UserHandler{logger: logger, userService: userService}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.ParamError)
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.Wrap(err, xerror.CodeParamError, err.Error()))
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建用户失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xhttp.Error(c, xerror.ParamError)
		return
	}

	err := h.userService.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("删除用户失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, nil)
}

// HandleUpdate 更新用户信息
func (h *UserHandler) Update(c *gin.Context) {
	var req dto.UserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.Wrap(err, xerror.CodeParamError, err.Error()))
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.Wrap(err, xerror.CodeParamError, err.Error()))
		return
	}

	id := c.Param("id")
	if id == "" {
		xhttp.Error(c, xerror.Wrap(xerror.ParamError, xerror.CodeParamError, "id is required"))
		return
	}
	req.ID = id

	user, err := h.userService.Update(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("更新用户失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, user)
}

// HandleGet 获取用户信息
func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xhttp.Error(c, xerror.ParamError)
		return
	}

	user, err := h.userService.Get(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("获取用户失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, user)
}

func (h *UserHandler) List(c *gin.Context) {
	var req dto.UserList
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.ParamError)
		return
	}

	users, total, err := h.userService.List(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("获取用户列表失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, xhttp.NewResponseData(xerror.Success, users).WithTotal(total))
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.ParamError)
		return
	}

	token, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("登录失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, token)
}
