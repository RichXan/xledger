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

const (
	SubCategoryID = "sub_category_id"
)

type SubCategoryHandler struct {
	logger          *xlog.Logger
	subCategoryService service.SubCategoryService
}

func NewSubCategoryHandler(logger *xlog.Logger, subCategoryService service.SubCategoryService) *SubCategoryHandler {
	return &SubCategoryHandler{logger: logger, subCategoryService: subCategoryService}
}

func (h *SubCategoryHandler) Create(c *gin.Context) {
	var req dto.SubCategoryCreate
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

	subCategory, err := h.subCategoryService.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("创建子类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, subCategory)
}

func (h *SubCategoryHandler) Delete(c *gin.Context) {
	id := c.Param(SubCategoryID)
	if id == "" {
		xhttp.Error(c, xerror.ParamError)
		return
	}

	err := h.subCategoryService.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("删除子类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, nil)
}

// HandleUpdate 更新子类目信息
func (h *SubCategoryHandler) Update(c *gin.Context) {
	var req dto.SubCategoryUpdate
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

	id := c.Param(SubCategoryID)
	if id == "" {
		xhttp.Error(c, xerror.Wrap(xerror.ParamError, xerror.CodeParamError, "id is required"))
		return
	}
	req.ID = id

	subCategory, err := h.subCategoryService.Update(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("更新子类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, subCategory)
}

// HandleGet 获取子类目信息
func (h *SubCategoryHandler) Get(c *gin.Context) {
	id := c.Param(SubCategoryID)
	if id == "" {
		xhttp.Error(c, xerror.ParamError)
		return
	}

	subCategory, err := h.subCategoryService.Get(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("获取子类目失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, subCategory)
}

func (h *SubCategoryHandler) List(c *gin.Context) {
	var req dto.SubCategoryList
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error().Err(err).Msg("参数错误")
		xhttp.Error(c, xerror.ParamError)
		return
	}

	subCategorys, total, err := h.subCategoryService.List(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("获取子类目列表失败")
		xhttp.Error(c, err)
		return
	}

	xhttp.Success(c, xhttp.NewResponseData(xerror.Success, subCategorys).WithTotal(total))
}
