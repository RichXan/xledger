package http

import (
	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func newDefaultAccountingHandler(classificationRepo classification.Repository) *accounting.Handler {
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	transactionRepo := accounting.NewInMemoryTransactionRepository()
	ledgerService := accounting.NewLedgerService(ledgerRepo)
	accountService := accounting.NewAccountService(accountRepo)
	var categoryService *classification.CategoryService
	var tagService *classification.TagService
	if classificationRepo != nil {
		categoryService = classification.NewCategoryService(classificationRepo)
		tagService = classification.NewTagService(classificationRepo)
	}
	transactionService := accounting.NewTransactionService(transactionRepo, ledgerRepo, accountRepo, categoryService, tagService)
	return accounting.NewHandler(ledgerService, accountService, transactionService)
}
