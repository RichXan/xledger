package http

import "xledger/backend/internal/classification"

func newDefaultClassificationHandler(repo classification.Repository) *classification.Handler {
	if repo == nil {
		repo = classification.NewInMemoryRepository()
	}
	categoryService := classification.NewCategoryService(repo)
	tagService := classification.NewTagService(repo)
	return classification.NewHandler(categoryService, tagService)
}
