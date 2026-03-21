package automation

type CSVMappingAdapter struct {
	enabled bool
	client  CSVMappingClient
}

func NewCSVMappingAdapter(enabled bool, client CSVMappingClient) *CSVMappingAdapter {
	return &CSVMappingAdapter{enabled: enabled, client: client}
}

func (a *CSVMappingAdapter) Suggest(columns []string) (map[string]string, error) {
	if a == nil || !a.enabled || a.client == nil {
		return map[string]string{}, nil
	}
	mapping, err := a.client(columns)
	if err == ErrCSVMappingUnavailable || err == ErrCSVMappingTimeout {
		return map[string]string{}, nil
	}
	if err != nil {
		return map[string]string{}, nil
	}
	if mapping == nil {
		return map[string]string{}, nil
	}
	return mapping, nil
}
