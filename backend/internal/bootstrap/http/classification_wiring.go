package http

import "xledger/backend/internal/classification"

func newDefaultClassificationHandler(deps *defaultBusinessDeps) *classification.Handler {
	return classification.NewHandler(deps.categoryService, deps.tagService)
}
