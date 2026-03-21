package http

import "xledger/backend/internal/portability"

func newDefaultPortabilityHandler() *portability.Handler {
	return portability.NewHandler(portability.NewImportPreviewService())
}
