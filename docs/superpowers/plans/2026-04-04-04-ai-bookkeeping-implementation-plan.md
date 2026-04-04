# AI 记账实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现自然语言记账 + OCR 截图记账，通过 iOS Shortcuts + n8n 调用 Go API

**Architecture:**
- iOS Shortcuts → n8n Webhook → LLM 解析 → Go API POST /api/transactions
- Go API 通过 PAT 认证（n8n 持有），支持 category_hint 和 account_hint 模糊匹配

**Tech Stack:** Apple Shortcuts, n8n, OpenAI/Claude API, Go + Gin

---

## 概述

AI 记账不修改 Go 核心业务逻辑，只扩展 `POST /api/transactions` 的 hint 解析能力。

```
用户: "午饭25元微信"
    │
    ▼
┌─────────────────────┐
│  iOS Shortcuts      │
│  自然语言/OCR文本    │
└─────────┬───────────┘
          │ HTTP POST {text, pat}
          ▼
┌─────────────────────┐
│  n8n Workflow       │
│  1. 验证 PAT        │
│  2. 调用 LLM 解析    │
│  3. 映射 hint → id │
│  4. POST /api/tx    │
└─────────────────────┘
          │
          ▼
┌─────────────────────┐
│  Go API             │
│  POST /transactions │
│  (via PAT)          │
└─────────────────────┘
```

---

## 文件影响

| 文件 | 动作 |
|------|------|
| `docs/superpowers/n8n-workflow-ai-bookkeeping.json` | 新建：n8n workflow 导出 |
| `docs/superpowers/apple-shortcut-ai-bookkeeping.md` | 新建：快捷指令配置文档 |
| `backend/internal/portability/quick_entry_adapter.go` | 修改：增强 hint 解析 |
| `backend/internal/portability/shortcut_adapter.go` | 新建：PAT 验证 |

---

## Task 1: 确认现有 PAT + Quick Add 能力

- [ ] **Step 1: 读取现有 shortcut_adapter.go**

```bash
cat backend/internal/portability/shortcut_adapter.go
```

- [ ] **Step 2: 读取现有 PAT 验证逻辑**

```bash
cat backend/internal/portability/pat_service.go
```

- [ ] **Step 3: 读取现有 quick-add API handler**

```bash
cat backend/internal/accounting/quick_add_handler.go
```

Expected: 已有基本的 quick-add 端点，需要增强 category/account hint 解析能力

- [ ] **Step 4: Commit 发现**

```bash
git log --oneline -5
```

如果现有代码已有 n8n/shortcut 相关实现，先充分理解再修改。

---

## Task 2: 增强 Go API - Hint 解析

- [ ] **Step 1: 读取现有 quick_add_handler.go**

确认现有实现，理解 `QuickAdd` 和 `ListCategories` 的当前逻辑。

- [ ] **Step 2: 创建 hint resolver**

新建 `backend/internal/portability/hint_resolver.go`：

```go
package portability

import (
    "context"
    "strings"

    "xledger/internal/classification"
    "xledger/internal/accounting"
)

type HintResolver struct {
    categoryService *classification.CategoryService
    accountRepo     accounting.AccountRepository
}

func NewHintResolver(categoryService *classification.CategoryService, accountRepo accounting.AccountRepository) *HintResolver {
    return &HintResolver{
        categoryService: categoryService,
        accountRepo:     accountRepo,
    }
}

type ResolvedHints struct {
    CategoryID   *string
    AccountID   *string
    Description string
}

// ResolveCategoryHint fuzzy-matches a category name to a user category ID.
// Returns nil if no match or multiple matches (ambiguous).
func (r *HintResolver) ResolveCategoryHint(ctx context.Context, userID, hint string) (*string, error) {
    if hint == "" {
        return nil, nil
    }

    hint = strings.TrimSpace(hint)
    categories, err := r.categoryService.ListCategories(ctx, userID)
    if err != nil {
        return nil, err
    }

    var matches []string
    hintLower := strings.ToLower(hint)

    for _, cat := range categories {
        nameLower := strings.ToLower(cat.Name)
        // 精确包含匹配
        if strings.Contains(nameLower, hintLower) || strings.Contains(hintLower, nameLower) {
            matches = append(matches, cat.ID)
            continue
        }
        // 关键词匹配
        keywords := strings.Fields(hintLower)
        allMatch := true
        for _, kw := range keywords {
            if !strings.Contains(nameLower, kw) {
                allMatch = false
                break
            }
        }
        if allMatch {
            matches = append(matches, cat.ID)
        }
    }

    if len(matches) == 1 {
        return &matches[0], nil
    }
    // 多个匹配（歧义）或无匹配：返回 nil
    return nil, nil
}

// ResolveAccountHint fuzzy-matches an account name to a user account ID.
func (r *HintResolver) ResolveAccountHint(ctx context.Context, userID, hint string) (*string, error) {
    if hint == "" {
        return nil, nil
    }

    hint = strings.TrimSpace(hint)
    accounts, err := r.accountRepo.List(ctx, userID)
    if err != nil {
        return nil, err
    }

    var matches []string
    hintLower := strings.ToLower(hint)

    for _, acct := range accounts {
        nameLower := strings.ToLower(acct.Name)
        if strings.Contains(nameLower, hintLower) || strings.Contains(hintLower, nameLower) {
            matches = append(matches, acct.ID)
        }
    }

    if len(matches) == 1 {
        return &matches[0], nil
    }
    return nil, nil
}
```

- [ ] **Step 3: 修改 QuickAdd handler 使用 hint resolver**

修改 `quick_add_handler.go`，在 `QuickAdd` 方法中使用 HintResolver：

```go
func (h *QuickAddHandler) QuickAdd(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    var req struct {
        Type          string   `json:"type"`  // "expense", "income", "transfer"
        Amount        float64  `json:"amount"`
        CategoryHint  string   `json:"category_hint"`
        AccountHint   string   `json:"account_hint"`
        Description   string   `json:"description"`
        Date          string   `json:"date"`  // YYYY-MM-DD
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
        return
    }

    // 解析 category_hint
    categoryID, _ := h.hintResolver.ResolveCategoryHint(c.Request.Context(), userID, req.CategoryHint)

    // 解析 account_hint
    accountID, _ := h.hintResolver.ResolveAccountHint(c.Request.Context(), userID, req.AccountHint)

    // 构建交易输入
    input := accounting.TransactionCreateInput{
        LedgerID:   h.getDefaultLedgerID(c.Request.Context(), userID),
        Type:        req.Type,
        Amount:      req.Amount,
        CategoryID:  categoryID,
        AccountID:   accountID,
        Memo:        &req.Description,
    }

    if req.Date != "" {
        if t, err := time.Parse("2006-01-02", req.Date); err == nil {
            input.OccurredAt = t
        }
    }

    txn, err := h.txnService.CreateTransaction(c.Request.Context(), userID, input)
    if err != nil {
        h.writeError(c, err)
        return
    }
    httpx.JSON(c, http.StatusCreated, "OK", "成功", txn)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/portability/hint_resolver.go backend/internal/portability/quick_add_handler.go
git commit -m "feat(ai-bookkeeping): add hint resolver for NL parsing

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 创建 n8n Workflow

- [ ] **Step 1: 创建 workflow JSON 文件**

新建 `docs/superpowers/n8n-workflow-ai-bookkeeping.json`：

```json
{
  "name": "Xledger AI Bookkeeping",
  "nodes": [
    {
      "id": "webhook-entry",
      "name": "Webhook",
      "type": "n8n-nodes-base.webhook",
      "position": [250, 300],
      "parameters": {
        "httpMethod": "POST",
        "path": "quick-entry",
        "responseMode": "responseNode",
        "options": {}
      }
    },
    {
      "id": "pat-validation",
      "name": "Validate PAT",
      "type": "n8n-nodes-base.code",
      "position": [500, 300],
      "parameters": {
        "mode": "runOnceForAllItems",
        "jsCode": "const pat = $input.first().json.body.pat\nconst item = $input.first().json\n// Validate PAT against /api/auth/me\nconst response = await fetch('https://YOUR_XLEDGER_HOST/api/auth/me', {\n  headers: { 'Authorization': `Bearer ${pat}` }\n})\nif (!response.ok) {\n  throw new Error('Invalid PAT')\n}\nitem.parsed = true\nreturn [{json: item}]"
      }
    },
    {
      "id": "llm-parse",
      "name": "LLM Parse",
      "type": "@n8n/n8n-nodes-langchain.openAi",
      "position": [750, 300],
      "parameters": {
        "operation": "chat",
        "model": "gpt-4o-mini",
        "messages": {
          "values": [
            {
              "role": "system",
              "content": "你是一个家庭记账助手。用户输入一段自然语言记账内容，你需要提取结构化信息。\n\n输出格式（仅返回JSON）：\n{\n  \"type\": \"expense|income|transfer\",\n  \"amount\": 数字,\n  \"category_hint\": \"分类名称\",\n  \"account_hint\": \"账户名称\",\n  \"description\": \"交易描述\",\n  \"date\": \"YYYY-MM-DD\"\n}\n\n规则：\n1. type必须从内容判断\n2. amount必须是正数\n3. category_hint和account_hint不确定则设为null\n4. date默认为今天"
            },
            {
              "role": "user",
              "content": "={{ $json.body.text }}"
            }
          ]
        }
      }
    },
    {
      "id": "create-transaction",
      "name": "Create Transaction",
      "type": "n8n-nodes-base.httpRequest",
      "position": [1000, 300],
      "parameters": {
        "method": "POST",
        "url": "https://YOUR_XLEDGER_HOST/api/transactions",
        "authentication": "genericCredentialType",
        "sendHeaders": true,
        "headerParameters": {
          "parameters": [
            {
              "name": "Authorization",
              "value": "Bearer {{ $json.pat }}"
            },
            {
              "name": "Content-Type",
              "value": "application/json"
            }
          ]
        },
        "sendBody": true,
        "bodyParameters": {
          "parameters": [
            { "name": "type", "value": "={{ $json.choices[0].message.content }}" },
            { "name": "ledger_id", "value": "={{ $('Get Ledgers').item.json.default_ledger_id }}" }
          ]
        }
      }
    }
  ],
  "connections": {
    "Webhook": { "main": [[{ "node": "Validate PAT", "type": "main" }]] },
    "Validate PAT": { "main": [[{ "node": "LLM Parse", "type": "main" }]] },
    "LLM Parse": { "main": [[{ "node": "Create Transaction", "type": "main" }]] }
  }
}
```

**注意：** 实际 workflow 需要根据 n8n 版本调整，JSON 只是示意。

- [ ] **Step 2: Commit**

```bash
git add docs/superpowers/n8n-workflow-ai-bookkeeping.json && git commit -m "docs: add n8n AI bookkeeping workflow template

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 创建 Apple Shortcuts 配置文档

- [ ] **Step 1: 创建文档**

新建 `docs/superpowers/apple-shortcut-ai-bookkeeping.md`：

```markdown
# Xledger AI 记账 Apple Shortcuts 配置指南

## 前提条件

1. 已部署 Xledger 后端（支持 HTTPS）
2. 已有 Personal Access Token（PAT）
3. 有可用的 n8n 实例

## iOS 快捷指令配置

### 方式一：自然语言记账

1. 打开 iOS「快捷指令」App
2. 创建新快捷指令
3. 添加动作：
   - **「文本」**：输入你想记账的内容，如"午饭25元微信"
   - **「URL」**：填入你的 n8n webhook 地址
     ```
     https://your-n8n-domain.com/webhook/quick-entry
     ```
   - **「获取 URL 内容」**：
     - 方法：POST
     - 头部：Content-Type: application/json
     - 请求体：
       ```json
       {
         "text": "午饭25元微信",
         "pat": "你的PAT"
       }
       ```
   - **「显示结果」**：显示上一步的响应内容

4. 为快捷指令设置一个名称和图标
5. 可选：添加 Siri 短语，如"记账"

### 方式二：OCR 截图记账

1. 打开 iOS「快捷指令」App
2. 创建新快捷指令
3. 添加动作：
   - **「截图」** 或 **「从照片选择」**
   - **「识别文本」**（iOS 原生 OCR）
   - **「文本」**：将识别结果存入变量
   - **「URL」**：填入 n8n webhook 地址
   - **「获取 URL 内容」**：POST 方式，请求体同方式一
   - **「显示结果」**

### Siri 触发示例

- "Hey Siri，记账" → 打开快捷指令 → 输入文本

### 测试

1. 在 iOS 快捷指令中运行你的快捷指令
2. 应该收到来自 Xledger 的确认消息
3. 登录 Xledger Web 查看交易是否创建成功
```

- [ ] **Step 2: Commit**

```bash
git add docs/superpowers/apple-shortcut-ai-bookkeeping.md && git commit -m "docs: add Apple Shortcuts AI bookkeeping setup guide

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 验证端到端

- [ ] **Step 1: 生成测试 PAT**

```bash
curl -X POST https://your-xledger-host/api/personal-access-tokens \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "n8n-workflow", "expires_in_days": 90}'
```

Expected: 返回 PAT token

- [ ] **Step 2: 测试 n8n webhook**

直接 POST 到 n8n webhook URL：

```bash
curl -X POST https://your-n8n-domain/webhook/quick-entry \
  -H "Content-Type: application/json" \
  -d '{"text": "午饭25元微信", "pat": "YOUR_PAT"}'
```

Expected: 返回交易创建结果

- [ ] **Step 3: 在 Xledger Web 验证**

登录 Web，检查交易列表中是否出现了对应的交易记录。
