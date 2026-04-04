# 导入工具完善实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完善 CSV 导入工具，支持支付宝 CSV 优先 + ezbookkeeping CSV 其次，自动检测格式并映射字段

**Architecture:**
- 前端：上传 CSV → 预览映射 → 确认导入
- 后端：格式检测 → 规则映射 → 事务导入

**Tech Stack:** Go + Gin, React, TanStack Query

---

## 文件影响

| 文件 | 动作 |
|------|------|
| `backend/internal/portability/csv.go` | 修改：添加支付宝格式检测和映射 |
| `backend/internal/portability/import_preview_service.go` | 修改：增强预览服务 |
| `backend/internal/portability/import_confirm_service.go` | 修改：处理支付宝记录 |
| `frontend/app/src/features/management/management-api.ts` | 修改：添加导入 API |
| `frontend/app/src/pages/import-page.tsx` | 新建：导入页面 |
| `frontend/app/src/App.tsx` | 修改：添加导入路由 |

---

## Task 1: 理解现有导入架构

- [ ] **Step 1: 读取现有 csv.go**

```bash
cat backend/internal/portability/csv.go
```

理解现有的 CSV 解析和字段映射逻辑。

- [ ] **Step 2: 读取现有 import_preview_service.go**

```bash
cat backend/internal/portability/import_preview_service.go
```

- [ ] **Step 3: 读取现有 import_confirm_service.go**

```bash
cat backend/internal/portability/import_confirm_service.go
```

- [ ] **Step 4: 读取前端导入 API**

```bash
cat frontend/app/src/features/management/management-api.ts
```

理解现有的导入 API 结构。

- [ ] **Step 5: Commit 发现**

记录现有代码结构，理解扩展点。

---

## Task 2: 添加支付宝格式检测

- [ ] **Step 1: 修改 csv.go - 添加格式检测**

```go
// csv.go 新增

type CSVFormat string

const (
    CSVFormatAlipay       CSVFormat = "alipay"
    CSVFormatEZBookkeeping CSVFormat = "ezbookkeeping"
    CSVFormatGeneric      CSVFormat = "generic"
)

// DetectCSVFormat 检测 CSV 文件格式
func DetectCSVFormat(headers []string) CSVFormat {
    headerStr := strings.Join(headers, ",")

    // 支付宝格式特征
    if strings.Contains(headerStr, "交易号") &&
        strings.Contains(headerStr, "收/支") &&
        strings.Contains(headerStr, "金额(元)") &&
        strings.Contains(headerStr, "商品说明") {
        return CSVFormatAlipay
    }

    // ezbookkeeping 格式特征（需要根据实际字段调整）
    lowerHeader := strings.ToLower(headerStr)
    if strings.Contains(lowerHeader, "create_time") ||
        strings.Contains(lowerHeader, "transaction_date") ||
        strings.Contains(lowerHeader, "ezbookkeeping") {
        return CSVFormatEZBookkeeping
    }

    return CSVFormatGeneric
}

// GetAlipayColumnMapping 返回支付宝 CSV 字段索引
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

// ParseAlipayRow 将支付宝 CSV 行解析为 TransactionInput
func ParseAlipayRow(row []string, mapping map[string]int) *TransactionImportInput {
    input := &TransactionImportInput{}

    // 解析金额
    if idx, ok := mapping["amount"]; ok && idx < len(row) {
        amountStr := strings.TrimSpace(row[idx])
        amountStr = strings.ReplaceAll(amountStr, ",", "")
        if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
            input.Amount = amount
        }
    }

    // 解析方向
    if idx, ok := mapping["direction"]; ok && idx < len(row) {
        dir := strings.TrimSpace(row[idx])
        if dir == "收入" {
            input.Type = "income"
            // 支付宝收入金额为正
        } else if dir == "支出" {
            input.Type = "expense"
            // 支付宝支出金额为正，需要转为负数
            input.Amount = -input.Amount
        }
    }

    // 解析描述
    if idx, ok := mapping["description"]; ok && idx < len(row) {
        input.Description = strings.TrimSpace(row[idx])
    }

    // 解析日期
    if idx, ok := mapping["created_at"]; ok && idx < len(row) {
        dateStr := strings.TrimSpace(row[idx])
        // 支付宝格式：2024-01-15 12:30:45
        if t, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
            input.OccurredAt = t
        } else if t, err := time.Parse("2006-01-02", dateStr); err == nil {
            input.OccurredAt = t
        }
    }

    return input
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/portability/csv.go && git commit -m "feat(import): add Alipay CSV format detection and parsing

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 增强 Import Preview Service

- [ ] **Step 1: 修改 import_preview_service.go**

在 `ImportPreviewService` 中添加支付宝支持：

```go
// 找到 DetectFormat 调用，添加：
format := DetectCSVFormat(headers)
result.Format = string(format)

// 如果是支付宝，自动生成映射：
if format == CSVFormatAlipay {
    mapping := GetAlipayColumnMapping(headers)
    // 构建自动映射配置返回给前端
    result.AutoMapping = map[string]string{
        "description": "商品说明",
        "amount":      "金额(元)",
        "direction":   "收/支",
        "date":        "创建时间",
    }
    result.SuggestedCategories = s.suggestCategories(c, userID, previewRows)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/portability/import_preview_service.go && git commit -m "feat(import): enhance preview service with Alipay auto-mapping

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 增强 Import Confirm Service

- [ ] **Step 1: 修改 import_confirm_service.go**

在事务处理循环中添加支付宝格式处理：

```go
// 在事务循环中：
for _, row := range rows {
    var input *TransactionImportInput

    switch format {
    case CSVFormatAlipay:
        mapping := GetAlipayColumnMapping(headers)
        input = ParseAlipayRow(row, mapping)
    case CSVFormatEZBookkeeping:
        // ezbk 格式处理
        input = parseEZBookkeepingRow(row, mapping)
    default:
        input = parseGenericRow(row, mapping)
    }

    // 检查去重
    if s.isDuplicate(ctx, userID, ledgerID, input) {
        result.Skipped++
        continue
    }

    // 创建交易
    _, err := s.txnService.CreateTransaction(ctx, userID, toCreateInput(input, ledgerID))
    if err != nil {
        result.Failed++
        result.Errors = append(result.Errors, err.Error())
        continue
    }
    result.Success++
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/portability/import_confirm_service.go && git commit -m "feat(import): add Alipay row parsing to import confirm

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 创建前端导入页面

- [ ] **Step 1: 创建 import-page.tsx**

```typescript
// frontend/app/src/pages/import-page.tsx
import { useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Upload, AlertCircle, CheckCircle } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { importCSVPreview, importCSVConfirm } from '@/features/management/management-api'

type ImportStep = 'upload' | 'mapping' | 'result'

export function ImportPage() {
  const { t } = useTranslation()
  const [step, setStep] = useState<ImportStep>('upload')
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<any>(null)
  const [mapping, setMapping] = useState<Record<string, string>>({})
  const fileInputRef = useRef<HTMLInputElement>(null)

  const previewMutation = useMutation({
    mutationFn: async (formData: FormData) => {
      const response = await importCSVPreview(formData)
      return response
    },
    onSuccess: (data) => {
      setPreview(data)
      setMapping(data.auto_mapping || {})
      setStep('mapping')
    },
  })

  const confirmMutation = useMutation({
    mutationFn: async (mapping: Record<string, string>) => {
      const formData = new FormData()
      if (file) formData.append('file', file)
      formData.append('mapping', JSON.stringify(mapping))
      const response = await importCSVConfirm(formData)
      return response
    },
    onSuccess: (data) => {
      setStep('result')
    },
  })

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (selectedFile) {
      setFile(selectedFile)
      const formData = new FormData()
      formData.append('file', selectedFile)
      previewMutation.mutate(formData)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    const droppedFile = e.dataTransfer.files[0]
    if (droppedFile && droppedFile.name.endsWith('.csv')) {
      setFile(droppedFile)
      const formData = new FormData()
      formData.append('file', droppedFile)
      previewMutation.mutate(formData)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold">{t('import.title')}</h1>

      {step === 'upload' && (
        <div
          className="border-2 border-dashed border-outline/30 rounded-2xl p-12 text-center cursor-pointer hover:border-primary transition-colors"
          onDrop={handleDrop}
          onDragOver={(e) => e.preventDefault()}
          onClick={() => fileInputRef.current?.click()}
        >
          <Upload className="mx-auto mb-4 text-on-surface-variant" size={48} />
          <p className="text-lg font-medium">{t('import.selectFile')}</p>
          <p className="text-sm text-on-surface-variant mt-2">
            {t('import.supportedFormats')}
          </p>
          <input
            ref={fileInputRef}
            type="file"
            accept=".csv"
            className="hidden"
            onChange={handleFileChange}
          />
        </div>
      )}

      {step === 'mapping' && preview && (
        <div className="space-y-6">
          {/* 格式检测结果 */}
          <div className="rounded-xl bg-surface-container p-4">
            <p className="text-sm font-medium">
              {t('import.detectedFormat')}: <span className="text-primary">{preview.format}</span>
            </p>
            <p className="text-sm text-on-surface-variant mt-1">
              {preview.total_rows} {t('import.totalRows')}
            </p>
          </div>

          {/* 预览表格 */}
          <div className="overflow-x-auto rounded-xl border border-outline/15">
            <table className="w-full text-sm">
              <thead className="bg-surface-container">
                <tr>
                  <th className="px-4 py-2 text-left">{t('import.preview')}</th>
                  {preview.headers?.slice(0, 5).map((h: string) => (
                    <th key={h} className="px-4 py-2 text-left">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {preview.preview_rows?.slice(0, 3).map((row: string[], i: number) => (
                  <tr key={i} className="border-t border-outline/10">
                    <td className="px-4 py-2 text-on-surface-variant">#{i + 1}</td>
                    {row.slice(0, 5).map((cell: string, j: number) => (
                      <td key={j} className="px-4 py-2">{cell}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* 分类映射（如果是支付宝，自动建议） */}
          <div className="space-y-3">
            <h3 className="font-semibold">{t('import.fieldMapping')}</h3>
            {Object.entries(mapping).map(([field, csvCol]) => (
              <div key={field} className="flex items-center gap-4">
                <label className="w-32 text-sm font-medium">{field}</label>
                <select
                  value={csvCol}
                  onChange={(e) => setMapping({ ...mapping, [field]: e.target.value })}
                  className="flex-1 rounded-xl border border-outline/15 px-4 py-2"
                >
                  <option value="">--</option>
                  {preview.headers?.map((h: string) => (
                    <option key={h} value={h}>{h}</option>
                  ))}
                </select>
              </div>
            ))}
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-4">
            <button
              onClick={() => confirmMutation.mutate(mapping)}
              disabled={confirmMutation.isPending}
              className="flex-1 rounded-xl bg-primary py-3 text-white font-semibold disabled:opacity-50"
            >
              {confirmMutation.isPending ? t('common.loading') : t('import.confirm')}
            </button>
            <button
              onClick={() => { setStep('upload'); setFile(null); setPreview(null) }}
              className="flex-1 rounded-xl border border-outline/15 py-3"
            >
              {t('common.cancel')}
            </button>
          </div>
        </div>
      )}

      {step === 'result' && (
        <div className="space-y-6 text-center py-8">
          <CheckCircle className="mx-auto text-emerald-500" size={64} />
          <h2 className="text-xl font-bold">{t('import.success')}</h2>
          <p className="text-on-surface-variant">
            {confirmMutation.data?.success} {t('import.successRows')}
            {confirmMutation.data?.skipped > 0 && ` • ${confirmMutation.data.skipped} ${t('import.skipped')}`}
            {confirmMutation.data?.failed > 0 && ` • ${confirmMutation.data.failed} ${t('import.failed')}`}
          </p>
          <button
            onClick={() => { setStep('upload'); setFile(null); setPreview(null) }}
            className="rounded-xl bg-primary px-8 py-3 text-white font-semibold"
          >
            {t('import.importMore')}
          </button>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: 修改 management-api.ts 添加导入 API**

```typescript
export async function importCSVPreview(formData: FormData) {
  const response = await fetch('/api/import/csv', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${getAccessToken()}`,
    },
    body: formData,
  })
  const json = await response.json()
  if (!response.ok) throw new Error(json.message)
  return json.data
}

export async function importCSVConfirm(formData: FormData) {
  const response = await fetch('/api/import/csv/confirm', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${getAccessToken()}`,
    },
    body: formData,
  })
  const json = await response.json()
  if (!response.ok) throw new Error(json.message)
  return json.data
}
```

- [ ] **Step 3: 添加路由到 App.tsx**

```typescript
import { ImportPage } from '@/pages/import-page'

// 添加路由：
<Route path="/import" element={<ImportPage />} />
```

- [ ] **Step 4: Commit**

```bash
git add frontend/app/src/pages/import-page.tsx frontend/app/src/features/management/management-api.ts frontend/app/src/App.tsx
git commit -m "feat(import): add CSV import page with Alipay format support

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 最终验证

- [ ] **Step 1: 获取支付宝导出 CSV 示例**

找一份真实的支付宝导出 CSV 文件（可以去支付宝App导出），验证表头格式。

- [ ] **Step 2: 测试导入流程**

1. 前端上传支付宝 CSV
2. 验证格式自动检测为 "alipay"
3. 验证字段自动映射正确
4. 点击确认导入
5. 验证交易创建成功

- [ ] **Step 3: 测试 ezbk 导入**

如果有 ezbk 导出的 CSV 文件，同样测试导入流程。

- [ ] **Step 4: Commit 最终状态**

```bash
git status
```
