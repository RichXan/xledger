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

type CategoryHandler struct {
	logger          *xlog.Logger
	categoryService service.CategoryService
}

func NewCategoryHandler(logger *xlog.Logger, categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{logger: logger, categoryService: categoryService}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CategoryCreate
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

	userID := c.MustGet("user_id").(string)
	req.UserID = userID
	category, err := h.categoryService.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xhttp.Error(c, xerror.ParamError)
		return
	}

	err := h.categoryService.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("删除类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, nil)
}

// HandleUpdate 更新类目信息
func (h *CategoryHandler) Update(c *gin.Context) {
	var req dto.CategoryUpdate
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

	// userID := c.MustGet("user_id").(string)

	category, err := h.categoryService.Update(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("更新类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, category)
}

// HandleGet 获取类目信息
func (h *CategoryHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		xhttp.Error(c, xerror.ParamError)
		return
	}

	category, err := h.categoryService.Get(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("获取类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, category)
}

func (h *CategoryHandler) List(c *gin.Context) {
	var req dto.CategoryList
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.ParamError)
		return
	}

	userID := c.MustGet("user_id").(string)
	req.UserID = userID

	categorys, total, err := h.categoryService.List(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("获取类目列表失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, xhttp.NewResponseData(xerror.Success, categorys).WithTotal(total))
}
