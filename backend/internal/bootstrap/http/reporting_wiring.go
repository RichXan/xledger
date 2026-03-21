package http

import "xledger/backend/internal/reporting"

func newDefaultReportingHandler(deps *defaultBusinessDeps) *reporting.Handler {
	if deps == nil {
		return nil
	}
	repo := reporting.NewRepository(deps.accountRepo, deps.transactionRepo, deps.categoryService)
	return reporting.NewHandler(
		reporting.NewOverviewService(repo),
		reporting.NewTrendService(repo),
		reporting.NewCategoryService(repo),
	)
}
