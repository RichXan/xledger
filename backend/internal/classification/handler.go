package classification

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	categories, err := h.categoryService.ListCategories(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": categories})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	if h.categoryService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": CAT_INVALID})
		return
	}

	category, err := h.categoryService.CreateCategory(c.Request.Context(), userID, CategoryCreateInput{Name: req.Name, ParentID: req.ParentID})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	if h.categoryService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req updateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": CAT_INVALID})
		return
	}

	category, err := h.categoryService.UpdateCategory(c.Request.Context(), userID, c.Param("id"), CategoryUpdateInput{
		Name:        req.Name,
		ParentID:    req.ParentID,
		ClearParent: req.NoParent,
		Archive:     req.Archived,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	if h.categoryService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	result, err := h.categoryService.DeleteCategory(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		if ErrorCode(err) == CAT_IN_USE_ARCHIVED {
			c.JSON(http.StatusOK, gin.H{"deleted": true, "archived": true, "error_code": CAT_IN_USE_ARCHIVED, "category": result.Category})
			return
		}
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": result.Deleted, "archived": result.Archived, "category": result.Category})
}

func (h *Handler) ListTags(c *gin.Context) {
	if h.tagService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	tags, err := h.tagService.ListTags(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": tags})
}

func (h *Handler) CreateTag(c *gin.Context) {
	if h.tagService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": TAG_INVALID})
		return
	}

	tag, err := h.tagService.CreateTag(c.Request.Context(), userID, TagCreateInput{Name: req.Name})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, tag)
}

func (h *Handler) UpdateTag(c *gin.Context) {
	if h.tagService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req updateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": TAG_INVALID})
		return
	}

	tag, err := h.tagService.UpdateTag(c.Request.Context(), userID, c.Param("id"), TagUpdateInput{Name: req.Name})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, tag)
}

func (h *Handler) DeleteTag(c *gin.Context) {
	if h.tagService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	if err := h.tagService.DeleteTag(c.Request.Context(), userID, c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case CAT_INVALID, TAG_INVALID:
		c.JSON(http.StatusBadRequest, gin.H{"error_code": ErrorCode(err)})
	case CAT_INVALID_PARENT:
		c.JSON(http.StatusBadRequest, gin.H{"error_code": CAT_INVALID_PARENT})
	case CAT_NOT_FOUND, TAG_NOT_FOUND:
		c.JSON(http.StatusNotFound, gin.H{"error_code": ErrorCode(err)})
	case TAG_DUPLICATED:
		c.JSON(http.StatusConflict, gin.H{"error_code": TAG_DUPLICATED})
	case CAT_IN_USE_ARCHIVED:
		c.JSON(http.StatusOK, gin.H{"error_code": CAT_IN_USE_ARCHIVED, "archived": true})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "CLASSIFICATION_INTERNAL"})
	}
}

func userIDFromContext(c *gin.Context) (string, bool) {
	if value, exists := c.Get("user_id"); exists {
		if userID, ok := value.(string); ok && userID != "" {
			return userID, true
		}
	}
	return "", false
}
