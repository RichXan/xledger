package http

import (
	"database/sql"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
	"xledger/backend/internal/portability"
	"xledger/backend/internal/reporting"
)

type defaultBusinessDeps struct {
	ledgerRepo         *accounting.InMemoryLedgerRepository
	accountRepo        *accounting.InMemoryAccountRepository
	transactionRepo    *accounting.InMemoryTransactionRepository
	classificationRepo classification.Repository
	categoryService    *classification.CategoryService
	tagService         *classification.TagService
	txnService         *accounting.TransactionService
	ledgerService      *accounting.LedgerService
	patService         *portability.PATService
	reportingCache     reporting.Cache // nil = no caching
}

func newDefaultBusinessDeps() *defaultBusinessDeps {
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	transactionRepo := accounting.NewInMemoryTransactionRepository()
	ledgerService := accounting.NewLedgerService(ledgerRepo)
	txnService := accounting.NewTransactionService(transactionRepo, ledgerRepo, accountRepo, categoryService, tagService)
	patService := portability.NewPATService(nil)
	return &defaultBusinessDeps{
		ledgerRepo:         ledgerRepo,
		accountRepo:        accountRepo,
		transactionRepo:    transactionRepo,
		classificationRepo: classificationRepo,
		categoryService:    categoryService,
		tagService:         tagService,
		txnService:         txnService,
		ledgerService:      ledgerService,
		patService:         patService,
	}
}

func newDefaultAccountingHandler(deps *defaultBusinessDeps) *accounting.Handler {
	return accounting.NewHandler(deps.ledgerService, deps.accountService(), deps.txnService)
}

func (d *defaultBusinessDeps) accountService() *accounting.AccountService {
	return accounting.NewAccountService(d.accountRepo)
}

// AccountingHandlerWithPostgreSQL holds the accounting domain components
// needed by both the accounting handler and other domains (e.g., reporting, portability).
type AccountingHandlerWithPostgreSQL struct {
	AccountingHandler *accounting.Handler
	AccountRepo       accounting.AccountRepository
	CategoryService   *classification.CategoryService
	TagService        *classification.TagService
	TxnService        *accounting.TransactionService
	LedgerService     *accounting.LedgerService
	TxnRepo           accounting.TransactionRepository
}

func newAccountingHandlerWithPostgreSQL(db *sql.DB) AccountingHandlerWithPostgreSQL {
	ledgerRepo := accounting.NewPostgresLedgerRepository(db)
	accountRepo := accounting.NewPostgresAccountRepository(db)
	txnRepo := accounting.NewPostgresTransactionRepository(db)
	classificationRepo := classification.NewPostgresRepository(db)
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	ledgerService := accounting.NewLedgerService(ledgerRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	accountingHandler := accounting.NewHandler(ledgerService, accounting.NewAccountService(accountRepo), txnService)
	return AccountingHandlerWithPostgreSQL{
		AccountingHandler: accountingHandler,
		AccountRepo:       accountRepo,
		CategoryService:   categoryService,
		TagService:        tagService,
		TxnService:        txnService,
		LedgerService:     ledgerService,
		TxnRepo:           txnRepo,
	}
}
