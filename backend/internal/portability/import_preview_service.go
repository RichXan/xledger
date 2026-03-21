package portability

import (
	"encoding/csv"
	"io"
	"sort"
	"strings"
)

const IMPORT_INVALID_FILE = "IMPORT_INVALID_FILE"

type contractError struct{ code string }

func (e *contractError) Error() string { return e.code }

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if typed, ok := err.(*contractError); ok {
		return typed.code
	}
	return err.Error()
}

type PreviewResponse struct {
	Columns           []string            `json:"columns"`
	SampleRows        [][]string          `json:"sample_rows"`
	MappingSlots      []string            `json:"mappingSlots"`
	MappingCandidates map[string][]string `json:"mappingCandidates"`
	SuggestedMapping  map[string]string   `json:"suggested_mapping,omitempty"`
}

type ImportPreviewService struct{}

func NewImportPreviewService() *ImportPreviewService { return &ImportPreviewService{} }

func (s *ImportPreviewService) PreviewCSV(reader io.Reader) (PreviewResponse, error) {
	rows, err := csv.NewReader(reader).ReadAll()
	if err != nil || len(rows) < 2 {
		return PreviewResponse{}, &contractError{code: IMPORT_INVALID_FILE}
	}
	columns := normalizeColumns(rows[0])
	if len(columns) == 0 {
		return PreviewResponse{}, &contractError{code: IMPORT_INVALID_FILE}
	}
	sampleRows := make([][]string, 0, min(5, len(rows)-1))

	for _, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}
		sampleRows = append(sampleRows, row)
		if len(sampleRows) == 5 {
			break
		}
	}
	if len(sampleRows) == 0 {
		return PreviewResponse{}, &contractError{code: IMPORT_INVALID_FILE}
	}
	mappingSlots := []string{"amount", "date", "description", "category", "account", "tag"}
	return PreviewResponse{
		Columns:           columns,
		SampleRows:        sampleRows,
		MappingSlots:      mappingSlots,
		MappingCandidates: buildMappingCandidates(columns, mappingSlots),
	}, nil
}

func normalizeColumns(columns []string) []string {
	normalized := make([]string, 0, len(columns))
	for _, column := range columns {
		trimmed := strings.TrimSpace(column)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func buildMappingCandidates(columns []string, slots []string) map[string][]string {
	candidates := make(map[string][]string, len(slots))
	for _, slot := range slots {
		matches := make([]string, 0, len(columns))
		for _, column := range columns {
			lowerColumn := strings.ToLower(column)
			lowerSlot := strings.ToLower(slot)
			if strings.Contains(lowerColumn, lowerSlot) || strings.Contains(lowerSlot, lowerColumn) {
				matches = append(matches, column)
			}
		}
		if len(matches) == 0 {
			matches = append(matches, columns...)
		}
		sort.Strings(matches)
		candidates[slot] = matches
	}
	return candidates
}

func (s *ImportPreviewService) PreviewCSVWithSuggestions(reader io.Reader, suggester interface {
	Suggest(columns []string) (map[string]string, error)
}) (PreviewResponse, error) {
	result, err := s.PreviewCSV(reader)
	if err != nil {
		return PreviewResponse{}, err
	}
	if suggester == nil {
		result.SuggestedMapping = map[string]string{}
		return result, nil
	}
	mapping, suggestErr := suggester.Suggest(result.Columns)
	if suggestErr != nil {
		result.SuggestedMapping = map[string]string{}
		return result, nil
	}
	if mapping == nil {
		mapping = map[string]string{}
	}
	result.SuggestedMapping = mapping
	return result, nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
