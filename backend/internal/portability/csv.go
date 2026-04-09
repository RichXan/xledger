package portability

import (
	"encoding/csv"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type CSVFormat string

const (
	CSVFormatAlipay        CSVFormat = "alipay"
	CSVFormatEZBookkeeping CSVFormat = "ezbookkeeping"
	CSVFormatGeneric       CSVFormat = "generic"
)

const DATE_LAYOUT = "2006-01-02 15:04:05"

// DetectCSVFormat detects the CSV format based on headers
func DetectCSVFormat(headers []string) CSVFormat {
	headerStr := strings.Join(headers, ",")

	// Alipay format features
	if strings.Contains(headerStr, "交易号") &&
		strings.Contains(headerStr, "收/支") &&
		strings.Contains(headerStr, "金额(元)") &&
		strings.Contains(headerStr, "商品说明") {
		return CSVFormatAlipay
	}

	// ezbookkeeping format features
	lowerHeader := strings.ToLower(headerStr)
	if strings.Contains(lowerHeader, "create_time") ||
		strings.Contains(lowerHeader, "transaction_date") ||
		strings.Contains(lowerHeader, "ezbookkeeping") {
		return CSVFormatEZBookkeeping
	}

	return CSVFormatGeneric
}

// GetAlipayColumnMapping returns the column index mapping for Alipay CSV
func GetAlipayColumnMapping(headers []string) map[string]int {
	mapping := make(map[string]int)
	for i, h := range headers {
		switch h {
		case "交易号":
			mapping["transaction_id"] = i
		case "交易对方":
			mapping["counterparty"] = i
		case "商品说明":
			mapping["description"] = i
		case "金额(元)":
			mapping["amount"] = i
		case "收/支":
			mapping["direction"] = i
		case "创建时间":
			mapping["created_at"] = i
		case "备注":
			mapping["memo"] = i
		}
	}
	return mapping
}

// ParseAlipayRow parses a single Alipay CSV row into ImportRow
func ParseAlipayRow(row []string, mapping map[string]int) *ImportRow {
	input := &ImportRow{}

	// Parse amount
	if idx, ok := mapping["amount"]; ok && idx < len(row) {
		amountStr := strings.TrimSpace(row[idx])
		amountStr = strings.ReplaceAll(amountStr, ",", "")
		if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
			input.Amount = math.Abs(amount)
		}
	}

	// Parse direction
	if idx, ok := mapping["direction"]; ok && idx < len(row) {
		dir := strings.TrimSpace(row[idx])
		if dir == "收入" {
			input.Type = "income"
		} else if dir == "支出" {
			input.Type = "expense"
			// Alipay expense amount is stored as positive, need to negate
			input.Amount = -input.Amount
		}
	}

	// Parse description
	if idx, ok := mapping["description"]; ok && idx < len(row) {
		input.Description = strings.TrimSpace(row[idx])
	}

	// Parse date
	if idx, ok := mapping["created_at"]; ok && idx < len(row) {
		dateStr := strings.TrimSpace(row[idx])
		if t, err := time.Parse(DATE_LAYOUT, dateStr); err == nil {
			input.Date = t.Format("2006-01-02")
		} else if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			input.Date = t.Format("2006-01-02")
		}
	}

	// Parse memo as description fallback
	if input.Description == "" {
		if idx, ok := mapping["memo"]; ok && idx < len(row) {
			input.Description = strings.TrimSpace(row[idx])
		}
	}

	// Use counterparty as description if still empty
	if input.Description == "" {
		if idx, ok := mapping["counterparty"]; ok && idx < len(row) {
			input.Description = strings.TrimSpace(row[idx])
		}
	}

	return input
}

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
