package http

import "xledger/backend/internal/portability"

func newDefaultPortabilityHandler() *portability.Handler {
	repo := portability.NewRepository(nil)
	return portability.NewHandler(portability.NewImportPreviewService(), portability.NewImportConfirmService(repo))
}
