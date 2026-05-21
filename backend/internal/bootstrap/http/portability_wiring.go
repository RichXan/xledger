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
	var categoryService *classification.CategoryService
	if deps != nil {
		patService = deps.patService
		categoryService = deps.categoryService
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
		newImportConfirmService(repo, categoryService),
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
		newImportConfirmService(portability.NewPostgresRepository(db), categoryService),
		portability.NewExportService(exportRepo),
		patService,
		shortcutHandler,
	)
}

func newImportConfirmService(repo portability.ImportConfirmRepository, categoryService *classification.CategoryService) *portability.ImportConfirmService {
	if categoryService == nil {
		return portability.NewImportConfirmService(repo)
	}
	return portability.NewImportConfirmService(repo, categoryService)
}
