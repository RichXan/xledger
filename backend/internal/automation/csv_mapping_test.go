package automation

import (
	"bytes"
	"testing"

	"xledger/backend/internal/portability"
)

func TestCSVMappingSuggestion_ReturnsSuggestedMappingWhenEnabled(t *testing.T) {
	service := portability.NewImportPreviewService()
	adapter := NewCSVMappingAdapter(true, func(columns []string) (map[string]string, error) {
		return map[string]string{"amount": "金额", "date": "日期"}, nil
	})
	result, err := service.PreviewCSVWithSuggestions(bytes.NewBufferString("日期,金额,备注\n2026-03-01,12.5,午饭\n"), adapter)
	if err != nil {
		t.Fatalf("preview with suggestions: %v", err)
	}
	if result.SuggestedMapping["amount"] != "金额" || result.SuggestedMapping["date"] != "日期" {
		t.Fatalf("expected suggested mapping, got %#v", result.SuggestedMapping)
	}
}

func TestCSVMappingSuggestion_DisabledOrUnavailable_FallsBackToManualOnly(t *testing.T) {
	service := portability.NewImportPreviewService()
	disabled := NewCSVMappingAdapter(false, nil)
	result, err := service.PreviewCSVWithSuggestions(bytes.NewBufferString("date,amount\n2026-03-01,12.5\n"), disabled)
	if err != nil {
		t.Fatalf("preview disabled mapping: %v", err)
	}
	if len(result.SuggestedMapping) != 0 {
		t.Fatalf("expected empty suggested mapping when disabled, got %#v", result.SuggestedMapping)
	}

	unavailable := NewCSVMappingAdapter(true, func(columns []string) (map[string]string, error) {
		return nil, ErrCSVMappingUnavailable
	})
	result, err = service.PreviewCSVWithSuggestions(bytes.NewBufferString("date,amount\n2026-03-01,12.5\n"), unavailable)
	if err != nil {
		t.Fatalf("preview unavailable mapping should still succeed: %v", err)
	}
	if len(result.SuggestedMapping) != 0 {
		t.Fatalf("expected empty suggested mapping when unavailable, got %#v", result.SuggestedMapping)
	}
}

func TestCSVMappingSuggestion_Timeout_NoImportBlock(t *testing.T) {
	service := portability.NewImportPreviewService()
	adapter := NewCSVMappingAdapter(true, func(columns []string) (map[string]string, error) {
		return nil, ErrCSVMappingTimeout
	})
	result, err := service.PreviewCSVWithSuggestions(bytes.NewBufferString("date,amount\n2026-03-01,12.5\n"), adapter)
	if err != nil {
		t.Fatalf("preview timeout should not fail import preview: %v", err)
	}
	if len(result.SuggestedMapping) != 0 {
		t.Fatalf("expected timeout fallback to empty suggested mapping, got %#v", result.SuggestedMapping)
	}
}
