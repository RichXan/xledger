package http

import (
	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

type defaultBusinessDeps struct {
	ledgerRepo         *accounting.InMemoryLedgerRepository
	accountRepo        *accounting.InMemoryAccountRepository
	transactionRepo    *accounting.InMemoryTransactionRepository
	classificationRepo classification.Repository
	categoryService    *classification.CategoryService
	tagService         *classification.TagService
}

func newDefaultBusinessDeps() *defaultBusinessDeps {
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	return &defaultBusinessDeps{
		ledgerRepo:         accounting.NewInMemoryLedgerRepository(),
		accountRepo:        accounting.NewInMemoryAccountRepository(),
		transactionRepo:    accounting.NewInMemoryTransactionRepository(),
		classificationRepo: classificationRepo,
		categoryService:    categoryService,
		tagService:         tagService,
	}
}

func newDefaultAccountingHandler(deps *defaultBusinessDeps) *accounting.Handler {
	ledgerService := accounting.NewLedgerService(deps.ledgerRepo)
	accountService := accounting.NewAccountService(deps.accountRepo)
	transactionService := accounting.NewTransactionService(deps.transactionRepo, deps.ledgerRepo, deps.accountRepo, deps.categoryService, deps.tagService)
	return accounting.NewHandler(ledgerService, accountService, transactionService)
}
