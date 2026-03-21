package automation

import "errors"

var (
	ErrCSVMappingUnavailable = errors.New("CSV_MAPPING_UNAVAILABLE")
	ErrCSVMappingTimeout     = errors.New("CSV_MAPPING_TIMEOUT")
)

type CSVMappingClient func(columns []string) (map[string]string, error)
