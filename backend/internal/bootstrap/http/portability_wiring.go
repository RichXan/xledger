package http

import (
	"database/sql"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
	"xledger/backend/internal/portability"
)

func newDefaultPortabilityHandler(deps *defaultBusinessDeps) *portability.Handler {
	repo := portability.NewRepository(nil)
	var patService *portability.PATService
	if deps != nil {
		patService = deps.patService
	}
	if patService == nil {
		patService = portability.NewPATService(nil)
	}
	var exportService *portability.ExportService
	if deps != nil {
		exportService = portability.NewExportService(portability.NewExportRepository(deps.transactionRepo, deps.categoryService))
	}

	var shortcutHandler *portability.ShortcutHandler
	if deps != nil && deps.txnService != nil && deps.ledgerService != nil {
		adapter := portability.NewShortcutAdapter(deps.txnService, deps.ledgerService, deps.categoryService)
		callbackRecorder := portability.NewInMemoryCallbackRecorder()
		shortcutHandler = portability.NewShortcutHandler(patService, adapter, adapter, adapter, callbackRecorder)
	}

	return portability.NewHandler(
		portability.NewImportPreviewService(),
		portability.NewImportConfirmService(repo),
		exportService,
		patService,
		shortcutHandler,
	)
}

func newPortabilityHandlerWithPostgreSQL(db *sql.DB, txnRepo accounting.TransactionRepository, ledgerService *accounting.LedgerService, categoryService *classification.CategoryService) *portability.Handler {
	exportRepo := portability.NewExportRepository(txnRepo, categoryService)
	patService := portability.NewPATService(nil)
	var shortcutHandler *portability.ShortcutHandler
	if txnRepo != nil && ledgerService != nil && categoryService != nil {
		txnService := accounting.NewTransactionService(txnRepo, nil, nil, categoryService, nil)
		adapter := portability.NewShortcutAdapter(txnService, ledgerService, categoryService)
		callbackRecorder := portability.NewInMemoryCallbackRecorder()
		shortcutHandler = portability.NewShortcutHandler(patService, adapter, adapter, adapter, callbackRecorder)
	}
	return portability.NewHandler(
		portability.NewImportPreviewService(),
		portability.NewImportConfirmService(portability.NewPostgresRepository(db)),
		portability.NewExportService(exportRepo),
		patService,
		shortcutHandler,
	)
}
