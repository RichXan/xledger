package http

import "xledger/backend/internal/accounting"

func newDefaultAccountingHandler() *accounting.Handler {
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	transactionRepo := accounting.NewInMemoryTransactionRepository()
	ledgerService := accounting.NewLedgerService(ledgerRepo)
	accountService := accounting.NewAccountService(accountRepo)
	transactionService := accounting.NewTransactionService(transactionRepo, ledgerRepo, accountRepo)
	return accounting.NewHandler(ledgerService, accountService, transactionService)
}
