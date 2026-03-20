package http

import "xledger/backend/internal/accounting"

func newDefaultAccountingHandler() *accounting.Handler {
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	ledgerService := accounting.NewLedgerService(ledgerRepo)
	accountService := accounting.NewAccountService(accountRepo)
	return accounting.NewHandler(ledgerService, accountService)
}
