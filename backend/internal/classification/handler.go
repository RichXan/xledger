package classification

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/common/httpx"
)

type Handler struct {
	categoryService *CategoryService
	tagService      *TagService
}

type createCategoryRequest struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

type updateCategoryRequest struct {
	Name     *string `json:"name"`
	ParentID *string `json:"parent_id"`
	NoParent bool    `json:"no_parent"`
	Archived *bool   `json:"archived"`
}

type createTagRequest struct {
	Name string `json:"name"`
}

type updateTagRequest struct {
	Name string `json:"name"`
}

func NewHandler(categoryService *CategoryService, tagService *TagService) *Handler {
	return &Handler{categoryService: categoryService, tagService: tagService}
}

func (h *Handler) ListCategories(c *gin.Context) {
	if h.categoryService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	categories, err := h.categoryService.ListCategories(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": categories, "pagination": gin.H{"page": 1, "page_size": len(categories), "total": len(categories), "total_pages": 1}})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	if h.categoryService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	category, err := h.categoryService.CreateCategory(c.Request.Context(), userID, CategoryCreateInput{Name: req.Name, ParentID: req.ParentID})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, "OK", "成功", category)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	if h.categoryService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req updateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	category, err := h.categoryService.UpdateCategory(c.Request.Context(), userID, c.Param("id"), CategoryUpdateInput{Name: req.Name, ParentID: req.ParentID, ClearParent: req.NoParent, Archive: req.Archived})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	if h.categoryService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	result, err := h.categoryService.DeleteCategory(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		if ErrorCode(err) == CAT_IN_USE_ARCHIVED {
			httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"deleted": true, "archived": true, "category": result.Category})
			return
		}
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"deleted": result.Deleted, "archived": result.Archived, "category": result.Category})
}

func (h *Handler) ListTags(c *gin.Context) {
	if h.tagService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	tags, err := h.tagService.ListTags(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": tags, "pagination": gin.H{"page": 1, "page_size": len(tags), "total": len(tags), "total_pages": 1}})
}

func (h *Handler) CreateTag(c *gin.Context) {
	if h.tagService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	tag, err := h.tagService.CreateTag(c.Request.Context(), userID, TagCreateInput{Name: req.Name})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, "OK", "成功", tag)
}

func (h *Handler) UpdateTag(c *gin.Context) {
	if h.tagService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req updateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	tag, err := h.tagService.UpdateTag(c.Request.Context(), userID, c.Param("id"), TagUpdateInput{Name: req.Name})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", tag)
}

func (h *Handler) DeleteTag(c *gin.Context) {
	if h.tagService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if err := h.tagService.DeleteTag(c.Request.Context(), userID, c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"deleted": true})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case CAT_INVALID, TAG_INVALID, CAT_INVALID_PARENT:
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
	case CAT_NOT_FOUND, TAG_NOT_FOUND:
		httpx.JSON(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "资源不存在", nil)
	case TAG_DUPLICATED:
		httpx.JSON(c, http.StatusConflict, "BUSINESS_RULE_VIOLATION", "业务规则不满足", nil)
	case CAT_IN_USE_ARCHIVED:
		httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"archived": true})
	default:
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
	}
}

func userIDFromContext(c *gin.Context) (string, bool) {
	return httpx.UserIDFromContext(c)
}
