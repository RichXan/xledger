package handler

import (
	"context"
	"net/http"
	"xledger/database/model"
	"xledger/internal/http/handler/dto"

	"github.com/RichXan/xcommon/xerror"
	"github.com/RichXan/xcommon/xhttp"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go-micro.dev/v4/logger"
)

type Repo[
	Tmodel model.Tmodel,
	TDto dto.TDto,
	TUpdate dto.TUpdate,
	TQueryParam dto.TQueryParam,
] interface {
	Create(ctx context.Context, Param TDto) (string, error)
	Delete(ctx context.Context, key string) error
	Update(ctx context.Context, key string, Param TUpdate) error
	Get(ctx context.Context, key string) (Tmodel, error)
	List(ctx context.Context, Query TQueryParam) (*xhttp.APIResponse, error)
}

type IPublicLogic interface {
	Create(ctx *gin.Context) string
	Delete(ctx *gin.Context)
	Update(ctx *gin.Context)
	Get(ctx *gin.Context)
	List(ctx *gin.Context)
}

type PublicLogic[
	Tmodel model.Tmodel,
	TDto dto.TDto,
	TUpdate dto.TUpdate,
	TQueryParam dto.TQueryParam,
] struct {
	repo Repo[Tmodel, TDto, TUpdate, TQueryParam]
}

func NewPublicLogic[Tmodel model.Tmodel,
	TDto dto.TDto,
	TUpdate dto.TUpdate,
	TQueryParam dto.TQueryParam](repo Repo[Tmodel, TDto, TUpdate, TQueryParam]) *PublicLogic[Tmodel, TDto, TUpdate, TQueryParam] {
	return &PublicLogic[Tmodel, TDto, TUpdate, TQueryParam]{
		repo: repo,
	}
}

// 创建
func (l *PublicLogic[Tmodel, TDto, Tupdate, TQueryParam]) Create(ctx *gin.Context) string {
	// 绑定参数
	var dto TDto

	// a := dto.(basicFrom.UserForm)
	if err := ctx.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		logger.Debug("ShouldBindJSON: ", err)
		ctx.JSON(http.StatusOK,
			xhttp.NewResponse(xerror.ParamError))
		return ""
	}
	// log
	logger.Debugf("Create Param: %+v", dto)

	// 创建
	id, err := l.repo.Create(ctx, dto)
	if err != nil {
		logger.Debug("Create Error: ", err)
		ctx.JSON(http.StatusOK, xhttp.NewResponseMessage(
			xerror.CreateError,
			err.Error(),
		))
		return ""
	}

	ctx.JSON(http.StatusOK,
		xhttp.NewResponse(xerror.Success))

	return id
}

// 删除
func (l *PublicLogic[Tmodel, TDto, Tupdate, TQueryParam]) Delete(ctx *gin.Context) {
	// 删除
	if err := l.repo.Delete(ctx, ctx.Param("queryKey")); err != nil {
		logger.Debug("Delete Error: ", err)
		ctx.JSON(http.StatusOK, xhttp.NewResponseMessage(
			xerror.DeleteError,
			err.Error(),
		))
		return
	}

	ctx.JSON(http.StatusOK,
		xhttp.NewResponse(xerror.Success))
}

// 修改
func (l *PublicLogic[Tmodel, TDto, Tupdate, TQueryParam]) Update(ctx *gin.Context) {
	var dto Tupdate

	// 绑定参数
	if err := ctx.ShouldBindBodyWith(&dto, binding.JSON); err != nil {
		logger.Debug("ShouldBindJSON: ", err)
		ctx.JSON(http.StatusOK,
			xhttp.NewResponse(xerror.ParamError))
		return
	}

	// 修改
	if err := l.repo.Update(ctx, ctx.Param("queryKey"), dto); err != nil {
		logger.Debug("Update Error: ", err)
		ctx.JSON(http.StatusOK, xhttp.NewResponseMessage(
			xerror.UpdateError,
			err.Error(),
		))
		return
	}

	ctx.JSON(http.StatusOK,
		xhttp.NewResponse(xerror.Success))
}

// 获取单个信息
func (l *PublicLogic[Tmodel, TDto, Tupdate, TQueryParam]) Get(ctx *gin.Context) {
	// 获取信息
	one, err := l.repo.Get(ctx, ctx.Param("queryKey"))
	if err != nil {
		logger.Debug("Get Error: ", err)
		ctx.JSON(http.StatusOK, xhttp.NewResponseMessage(
			xerror.GetError,
			err.Error(),
		))
		return
	}

	ctx.JSON(http.StatusOK, xhttp.NewResponseData(
		xerror.Success, one))
}

func (l *PublicLogic[Tmodel, TDto, Tupdate, TQueryParam]) List(ctx *gin.Context) {
	var dto TQueryParam

	// 绑定参数
	if err := ctx.ShouldBindQuery(&dto); err != nil {
		logger.Debug("ShouldBindQuery: ", err)
		ctx.JSON(http.StatusOK,
			xhttp.NewResponse(xerror.ParamError))
		return
	}

	// 获取列表信息
	resp, err := l.repo.List(ctx, dto)
	if err != nil {
		ctx.JSON(http.StatusOK, xhttp.NewResponseMessage(
			xerror.GetError,
			err.Error(),
		))
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
