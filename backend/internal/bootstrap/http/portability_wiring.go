package http

import "xledger/backend/internal/portability"

func newDefaultPortabilityHandler(deps *defaultBusinessDeps) *portability.Handler {
	repo := portability.NewRepository(nil)
	var exportService *portability.ExportService
	if deps != nil {
		exportService = portability.NewExportService(portability.NewExportRepository(deps.transactionRepo, deps.categoryService))
	}
	return portability.NewHandler(
		portability.NewImportPreviewService(),
		portability.NewImportConfirmService(repo),
		exportService,
	)
}
