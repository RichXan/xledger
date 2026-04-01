package portability

import (
	"encoding/csv"
	"io"
	"math"
	"strconv"
	"strings"
)

// ParseImportRowsFromCSV parses a CSV file and returns an ImportConfirmRequest.
// It handles various column name formats (Chinese/English) and normalizes amounts.
func ParseImportRowsFromCSV(reader io.Reader) (ImportConfirmRequest, error) {
	records, err := csv.NewReader(reader).ReadAll()
	if err != nil || len(records) < 2 {
		return ImportConfirmRequest{}, err
	}
	headerMap := map[string]int{}
	for idx, name := range records[0] {
		trimmed := strings.TrimSpace(name)
		trimmed = strings.TrimPrefix(trimmed, "\uFEFF")
		headerMap[trimmed] = idx
	}

	getIndex := func(candidates ...string) int {
		for _, candidate := range candidates {
			if idx, ok := headerMap[candidate]; ok {
				return idx
			}
		}
		return -1
	}

	dateIdx := getIndex("时间", "date", "Date", "time", "occurred_at")
	typeIdx := getIndex("类型", "type", "Type")
	purposeIdx := getIndex("用途/来源", "category", "Category")
	amountIdx := getIndex("金额", "amount", "Amount")
	noteIdx := getIndex("备注", "description", "memo", "note")
	signAmountIdx := getIndex("金额正负处理")

	if dateIdx < 0 || amountIdx < 0 {
		return ImportConfirmRequest{}, io.EOF
	}

	rows := make([]ImportRow, 0, len(records)-1)
	for _, raw := range records[1:] {
		date := csvCell(raw, dateIdx)
		amountRaw := csvCell(raw, amountIdx)
		if strings.TrimSpace(date) == "" || strings.TrimSpace(amountRaw) == "" {
			continue
		}

		amount, parseErr := parseImportAmount(amountRaw)
		if parseErr != nil {
			continue
		}

		rowType := normalizeImportType(csvCell(raw, typeIdx))
		signAmount := csvCell(raw, signAmountIdx)
		if rowType == "" && strings.Contains(signAmount, "-") {
			rowType = "expense"
		}
		if rowType == "" && strings.Contains(signAmount, "+") {
			rowType = "income"
		}
		if rowType == "" {
			rowType = "expense"
		}

		category := strings.TrimSpace(csvCell(raw, purposeIdx))
		description := strings.TrimSpace(csvCell(raw, noteIdx))
		if description == "" {
			description = category
		}

		rows = append(rows, ImportRow{
			Date:        strings.TrimSpace(date),
			Amount:      math.Abs(amount),
			Description: description,
			Type:        rowType,
			Category:    category,
		})
	}
	if len(rows) == 0 {
		return ImportConfirmRequest{}, io.EOF
	}
	return ImportConfirmRequest{Rows: rows}, nil
}

func csvCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func parseImportAmount(raw string) (float64, error) {
	replacer := strings.NewReplacer("¥", "", "￥", "", ",", "", " ", "")
	normalized := replacer.Replace(strings.TrimSpace(raw))
	return strconv.ParseFloat(normalized, 64)
}

func normalizeImportType(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(lower, "income"), strings.Contains(lower, "收入"):
		return "income"
	case strings.Contains(lower, "expense"), strings.Contains(lower, "支出"):
		return "expense"
	default:
		return ""
	}
}
