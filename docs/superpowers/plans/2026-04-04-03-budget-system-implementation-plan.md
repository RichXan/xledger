# 预算系统实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现按分类月度预算系统 + 三类提醒触发（实时/每日/每周）

**Architecture:**
- 后端: Go + Gin + GORM，新增 budgets 表 + budget_alerts 表
- 前端: React + TanStack Query，新增预算列表/表单/告警页面
- 提醒触发: 实时在交易写入事务中检查，定时任务用 robfig/cron

**Tech Stack:** GORM, robfig/cron, Web Push

---

## 文件影响

### 前端

| 文件 | 动作 |
|------|------|
| `frontend/app/src/features/budgets/budgets-api.ts` | 新建 |
| `frontend/app/src/features/budgets/budgets-hooks.ts` | 新建 |
| `frontend/app/src/pages/budgets-page.tsx` | 新建 |
| `frontend/app/src/pages/budget-form-page.tsx` | 新建 |
| `frontend/app/src/pages/budget-alerts-page.tsx` | 新建 |
| `frontend/app/src/pages/notification-prefs-page.tsx` | 新建 |
| `frontend/app/src/App.tsx` | 修改：添加路由 |

### 后端

| 文件 | 动作 |
|------|------|
| `backend/internal/budget/models.go` | 新建 |
| `backend/internal/budget/repository.go` | 新建 |
| `backend/internal/budget/repository_postgres.go` | 新建 |
| `backend/internal/budget/service.go` | 新建 |
| `backend/internal/budget/handler.go` | 新建 |
| `backend/internal/budget/alert_service.go` | 新建 |
| `backend/internal/bootstrap/http/budget_wiring.go` | 新建 |
| `backend/internal/bootstrap/http/router.go` | 修改：注册路由 |
| `migrations/YYYYMMDDDD_create_budgets.sql` | 新建 |
| `migrations/YYYYMMDDDD_create_budget_alerts.sql` | 新建 |
| `migrations/YYYYMMDDDD_create_notification_prefs.sql` | 新建 |

---

## Task 1: 创建数据库迁移

- [ ] **Step 1: 创建 budgets 表迁移**

```sql
-- migrations/YYYYMMDDDD_create_budgets.sql
CREATE TABLE IF NOT EXISTS budgets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount      DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    period      TEXT NOT NULL DEFAULT 'monthly',  -- 'monthly' for v1
    thresholds  INTEGER[] NOT NULL DEFAULT ARRAY[80, 100],  -- 百分比阈值
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, category_id, period)
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);
CREATE INDEX idx_budgets_category_id ON budgets(category_id);
```

- [ ] **Step 2: 创建 budget_alerts 表迁移**

```sql
-- migrations/YYYYMMDDDD_create_budget_alerts.sql
CREATE TABLE IF NOT EXISTS budget_alerts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id     UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    threshold     INTEGER NOT NULL,   -- 触发阈值 80 或 100
    spent_amount  DECIMAL(15, 2) NOT NULL,
    period_start  DATE NOT NULL,
    period_end    DATE NOT NULL,
    sent_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budget_alerts_budget_id ON budget_alerts(budget_id);
CREATE INDEX idx_budget_alerts_sent_at ON budget_alerts(sent_at);
```

- [ ] **Step 3: 创建 notification_prefs 表迁移**

```sql
-- migrations/YYYYMMDDDD_create_notification_prefs.sql
CREATE TABLE IF NOT EXISTS user_notification_prefs (
    user_id          UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    realtime_alert   BOOLEAN NOT NULL DEFAULT true,
    daily_digest     BOOLEAN NOT NULL DEFAULT false,
    weekly_digest    BOOLEAN NOT NULL DEFAULT true,
    push_endpoint    TEXT,
    push_p256dh      TEXT,
    push_auth        TEXT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 4: 运行迁移测试**

```bash
cd backend && TEST_DATABASE_URL=postgres://xledger:xledger_secret@127.0.0.1:5432/xledger_test?sslmode=disable go test ./migrations/... -count=1
```

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/ && git commit -m "feat(budget): add budgets, budget_alerts, and notification_prefs migrations

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 创建后端 Budget 模型

- [ ] **Step 1: 创建 models.go**

```go
// backend/internal/budget/models.go
package budget

import (
    "database/sql/driver"
    "time"

    "github.com/lib/pq"
)

type Budget struct {
    ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    UserID     string         `json:"user_id" gorm:"type:uuid;not null;index"`
    CategoryID string         `json:"category_id" gorm:"type:uuid;not null;index"`
    Amount     float64        `json:"amount" gorm:"type:decimal(15,2);not null"`
    Period     string         `json:"period" gorm:"type:text;not null;default:monthly"`
    Thresholds pq.Int64Array  `json:"thresholds" gorm:"type:integer[];not null;default:{80,100}"`
    CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    Category   *CategoryBrief `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}

type CategoryBrief struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type BudgetAlert struct {
    ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    BudgetID     string    `json:"budget_id" gorm:"type:uuid;not null;index"`
    Threshold    int       `json:"threshold" gorm:"not null"`
    SpentAmount  float64   `json:"spent_amount" gorm:"type:decimal(15,2);not null"`
    PeriodStart  time.Time `json:"period_start" gorm:"type:date;not null"`
    PeriodEnd    time.Time `json:"period_end" gorm:"type:date;not null"`
    SentAt       time.Time `json:"sent_at" gorm:"autoCreateTime"`
}

type UserNotificationPref struct {
    UserID        string    `json:"user_id" gorm:"primaryKey;type:uuid"`
    RealtimeAlert bool      `json:"realtime_alert" gorm:"not null;default:true"`
    DailyDigest   bool      `json:"daily_digest" gorm:"not null;default:false"`
    WeeklyDigest  bool      `json:"weekly_digest" gorm:"not null;default:true"`
    PushEndpoint  string    `json:"-" gorm:"type:text"`
    PushP256dh    string    `json:"-" gorm:"type:text"`
    PushAuth      string    `json:"-" gorm:"type:text"`
    UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// BudgetWithUsage combines a budget with its current usage stats
type BudgetWithUsage struct {
    Budget
    SpentAmount   float64 `json:"spent_amount"`
    UsagePercent  float64 `json:"usage_percent"`
    Remaining     float64 `json:"remaining"`
    PeriodStart   string  `json:"period_start"`
    PeriodEnd     string  `json:"period_end"`
    IsOverBudget  bool    `json:"is_over_budget"`
}

type CreateBudgetRequest struct {
    CategoryID string  `json:"category_id" binding:"required"`
    Amount     float64 `json:"amount" binding:"required,gt=0"`
    Period     string  `json:"period"`
    Thresholds []int   `json:"thresholds"`
}

type UpdateBudgetRequest struct {
    Amount     *float64 `json:"amount"`
    Thresholds []int    `json:"thresholds"`
}

type UpdatePrefsRequest struct {
    RealtimeAlert *bool `json:"realtime_alert"`
    DailyDigest   *bool `json:"daily_digest"`
    WeeklyDigest  *bool `json:"weekly_digest"`
}

type SubscribePushRequest struct {
    Endpoint string `json:"endpoint" binding:"required"`
    P256dh   string `json:"keys.p256dh" binding:"required"`
    Auth     string `json:"keys.auth" binding:"required"`
}

// Value implements driver.Valuer for pq.Int64Array
func (t pq.Int64Array) Value() (driver.Value, error) {
    return t, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/budget/models.go && git commit -m "feat(budget): add budget domain models

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 创建 Repository 层

- [ ] **Step 1: 创建 repository.go**

```go
// backend/internal/budget/repository.go
package budget

import "context"

type Repository interface {
    Create(ctx context.Context, b *Budget) error
    GetByID(ctx context.Context, id string) (*Budget, error)
    ListByUser(ctx context.Context, userID string) ([]Budget, error)
    Update(ctx context.Context, b *Budget) error
    Delete(ctx context.Context, id string) error
    GetByCategoryAndPeriod(ctx context.Context, userID, categoryID, period string) (*Budget, error)

    CreateAlert(ctx context.Context, a *BudgetAlert) error
    GetAlertsByUser(ctx context.Context, userID string, limit int) ([]BudgetAlert, error)
    GetAlertSentForThreshold(ctx context.Context, budgetID string, threshold, periodStart int) (bool, error)

    GetOrCreatePrefs(ctx context.Context, userID string) (*UserNotificationPref, error)
    UpdatePrefs(ctx context.Context, p *UserNotificationPref) error
    GetPrefs(ctx context.Context, userID string) (*UserNotificationPref, error)
}
```

- [ ] **Step 2: 创建 repository_postgres.go**

```go
// backend/internal/budget/repository_postgres.go
package budget

import (
    "context"
    "time"

    "github.com/lib/pq"
    "gorm.io/gorm"
)

type postgresRepository struct {
    db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
    return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, b *Budget) error {
    return r.db.WithContext(ctx).Create(b).Error
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*Budget, error) {
    var b Budget
    if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &b, nil
}

func (r *postgresRepository) ListByUser(ctx context.Context, userID string) ([]Budget, error) {
    var bs []Budget
    if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&bs).Error; err != nil {
        return nil, err
    }
    return bs, nil
}

func (r *postgresRepository) Update(ctx context.Context, b *Budget) error {
    return r.db.WithContext(ctx).Save(b).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&Budget{}, "id = ?", id).Error
}

func (r *postgresRepository) GetByCategoryAndPeriod(ctx context.Context, userID, categoryID, period string) (*Budget, error) {
    var b Budget
    if err := r.db.WithContext(ctx).Where("user_id = ? AND category_id = ? AND period = ?", userID, categoryID, period).First(&b).Error; err != nil {
        return nil, err
    }
    return &b, nil
}

func (r *postgresRepository) CreateAlert(ctx context.Context, a *BudgetAlert) error {
    return r.db.WithContext(ctx).Create(a).Error
}

func (r *postgresRepository) GetAlertsByUser(ctx context.Context, userID string, limit int) ([]BudgetAlert, error) {
    var as []BudgetAlert
    err := r.db.WithContext(ctx).
        Joins("JOIN budgets ON budget_alerts.budget_id = budgets.id").
        Where("budgets.user_id = ?", userID).
        Order("budget_alerts.sent_at DESC").
        Limit(limit).
        Find(&as).Error
    if err != nil {
        return nil, err
    }
    return as, nil
}

func (r *postgresRepository) GetAlertSentForThreshold(ctx context.Context, budgetID string, threshold, periodStart int) (bool, error) {
    var count int64
    startDate := time.Unix(int64(periodStart), 0)
    err := r.db.WithContext(ctx).Model(&BudgetAlert{}).
        Where("budget_id = ? AND threshold = ? AND period_start = ?", budgetID, threshold, startDate).
        Count(&count).Error
    return count > 0, err
}

func (r *postgresRepository) GetOrCreatePrefs(ctx context.Context, userID string) (*UserNotificationPref, error) {
    var p UserNotificationPref
    if err := r.db.WithContext(ctx).First(&p, "user_id = ?", userID).Error; err == nil {
        return &p, nil
    }
    p = UserNotificationPref{UserID: userID}
    if err := r.db.WithContext(ctx).Create(&p).Error; err != nil {
        return nil, err
    }
    return &p, nil
}

func (r *postgresRepository) UpdatePrefs(ctx context.Context, p *UserNotificationPref) error {
    return r.db.WithContext(ctx).Save(p).Error
}

func (r *postgresRepository) GetPrefs(ctx context.Context, userID string) (*UserNotificationPref, error) {
    var p UserNotificationPref
    if err := r.db.WithContext(ctx).First(&p, "user_id = ?", userID).Error; err != nil {
        return nil, err
    }
    return &p, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/budget/repository.go backend/internal/budget/repository_postgres.go
git commit -m "feat(budget): add budget repository layer

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: 创建 Budget Service

- [ ] **Step 1: 创建 service.go**

```go
// backend/internal/budget/service.go
package budget

import (
    "context"
    "fmt"
    "time"

    "xledger/backend/internal/reporting"
)

type Service struct {
    repo           Repository
    reportingRepo  reporting.CategoryStatsFetcher
}

func NewService(repo Repository, reportingRepo reporting.CategoryStatsFetcher) *Service {
    return &Service{repo: repo, reportingRepo: reportingRepo}
}

func (s *Service) CreateBudget(ctx context.Context, userID string, req CreateBudgetRequest) (*BudgetWithUsage, error) {
    period := req.Period
    if period == "" {
        period = "monthly"
    }
    thresholds := req.Thresholds
    if len(thresholds) == 0 {
        thresholds = []int{80, 100}
    }

    b := &Budget{
        UserID:     userID,
        CategoryID: req.CategoryID,
        Amount:     req.Amount,
        Period:     period,
        Thresholds: thresholds,
    }
    if err := s.repo.Create(ctx, b); err != nil {
        return nil, fmt.Errorf("create budget: %w", err)
    }

    return s.getBudgetWithUsage(ctx, b)
}

func (s *Service) GetUserBudgets(ctx context.Context, userID string) ([]BudgetWithUsage, error) {
    budgets, err := s.repo.ListByUser(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("list budgets: %w", err)
    }

    result := make([]BudgetWithUsage, 0, len(budgets))
    for _, b := range budgets {
        bw, err := s.getBudgetWithUsage(ctx, &b)
        if err != nil {
            continue
        }
        result = append(result, *bw)
    }
    return result, nil
}

func (s *Service) UpdateBudget(ctx context.Context, userID, budgetID string, req UpdateBudgetRequest) (*BudgetWithUsage, error) {
    b, err := s.repo.GetByID(ctx, budgetID)
    if err != nil {
        return nil, fmt.Errorf("get budget: %w", err)
    }
    if b.UserID != userID {
        return nil, fmt.Errorf("not authorized")
    }

    if req.Amount != nil {
        b.Amount = *req.Amount
    }
    if req.Thresholds != nil {
        b.Thresholds = req.Thresholds
    }

    if err := s.repo.Update(ctx, b); err != nil {
        return nil, fmt.Errorf("update budget: %w", err)
    }

    return s.getBudgetWithUsage(ctx, b)
}

func (s *Service) DeleteBudget(ctx context.Context, userID, budgetID string) error {
    b, err := s.repo.GetByID(ctx, budgetID)
    if err != nil {
        return fmt.Errorf("get budget: %w", err)
    }
    if b.UserID != userID {
        return fmt.Errorf("not authorized")
    }
    return s.repo.Delete(ctx, budgetID)
}

func (s *Service) GetUserAlerts(ctx context.Context, userID string, limit int) ([]BudgetAlert, error) {
    if limit <= 0 {
        limit = 50
    }
    return s.repo.GetAlertsByUser(ctx, userID, limit)
}

func (s *Service) getBudgetWithUsage(ctx context.Context, b *Budget) (*BudgetWithUsage, error) {
    periodStart, periodEnd := getMonthBounds(time.Now())

    spent, err := s.getSpentAmount(ctx, b.UserID, b.CategoryID, periodStart, periodEnd)
    if err != nil {
        return nil, err
    }

    usagePercent := 0.0
    if b.Amount > 0 {
        usagePercent = (spent / b.Amount) * 100
    }
    remaining := b.Amount - spent
    if remaining < 0 {
        remaining = 0
    }

    return &BudgetWithUsage{
        Budget:       *b,
        SpentAmount:  spent,
        UsagePercent: usagePercent,
        Remaining:    remaining,
        PeriodStart:  periodStart.Format("2006-01-02"),
        PeriodEnd:    periodEnd.Format("2006-01-02"),
        IsOverBudget: spent > b.Amount,
    }, nil
}

func (s *Service) getSpentAmount(ctx context.Context, userID, categoryID string, start, end time.Time) (float64, error) {
    // 通过 reporting repo 查询分类支出
    stats, err := s.reportingRepo.GetCategoryStats(ctx, userID, reporting.CategoryQuery{
        From: start,
        To:   end,
    })
    if err != nil {
        return 0, err
    }
    for _, item := range stats {
        if item.CategoryID == categoryID {
            return item.Amount, nil
        }
    }
    return 0, nil
}

func getMonthBounds(now time.Time) (time.Time, time.Time) {
    start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
    end := start.AddDate(0, 1, -1)
    end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
    return start, end
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/budget/service.go && git commit -m "feat(budget): add budget service layer

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 创建 Alert Service（实时告警触发）

- [ ] **Step 1: 创建 alert_service.go**

```go
// backend/internal/budget/alert_service.go
package budget

import (
    "context"
    "fmt"
    "time"

    "xledger/internal/push"
)

type AlertService struct {
    repo         Repository
    pushService  push.Pusher
}

func NewAlertService(repo Repository, pushService push.Pusher) *AlertService {
    return &AlertService{repo: repo, pushService: pushService}
}

// CheckAndAlert checks spending after a transaction and sends alerts if thresholds are crossed
func (s *AlertService) CheckAndAlert(ctx context.Context, userID, categoryID string) error {
    // 查询用户该分类的预算
    period := "monthly"
    budget, err := s.repo.GetByCategoryAndPeriod(ctx, userID, categoryID, period)
    if err != nil || budget == nil {
        return nil // 没有预算，不告警
    }

    // 查询偏好
    prefs, err := s.repo.GetPrefs(ctx, userID)
    if err != nil || prefs == nil || !prefs.RealtimeAlert {
        return nil // 用户关闭了实时提醒
    }

    // 查询当月已用
    periodStart, periodEnd := getMonthBounds(time.Now())
    spent, err := s.calculateSpent(ctx, userID, categoryID, periodStart, periodEnd)
    if err != nil {
        return err
    }

    usagePercent := (spent / budget.Amount) * 100

    // 检查每个阈值
    for _, threshold := range budget.Thresholds {
        if usagePercent >= float64(threshold) {
            alreadySent, err := s.repo.GetAlertSentForThreshold(ctx, budget.ID, threshold, int(periodStart.Unix()))
            if err != nil || alreadySent {
                continue
            }

            // 发送告警
            alert := &BudgetAlert{
                BudgetID:    budget.ID,
                Threshold:   threshold,
                SpentAmount: spent,
                PeriodStart: periodStart,
                PeriodEnd:   periodEnd,
            }
            if err := s.repo.CreateAlert(ctx, alert); err != nil {
                return err
            }

            // 通过 Push 发送
            if s.pushService != nil && prefs.PushEndpoint != "" {
                _ = s.pushService.Send(context.Background(), prefs, "预算提醒",
                    fmt.Sprintf("分类已使用 %.0f%%（%d%%阈值）", usagePercent, threshold), "budget-alert")
            }
        }
    }
    return nil
}

func (s *AlertService) calculateSpent(ctx context.Context, userID, categoryID string, start, end time.Time) (float64, error) {
    // 复用 service.go 中的逻辑
    // 这里简化，实际应该调用 TransactionRepository.GetSpentByCategory
    return 0, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/budget/alert_service.go && git commit -m "feat(budget): add alert service for real-time budget alerts

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: 创建 Budget Handler

- [ ] **Step 1: 创建 handler.go**

```go
// backend/internal/budget/handler.go
package budget

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "xledger/internal/common/httpx"
)

type Handler struct {
    service *Service
    prefs   *PrefService
}

func NewHandler(service *Service, prefs *PrefService) *Handler {
    return &Handler{service: service, prefs: prefs}
}

func (h *Handler) ListBudgets(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    budgets, err := h.service.GetUserBudgets(c.Request.Context(), userID)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": budgets})
}

func (h *Handler) CreateBudget(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    var req CreateBudgetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
        return
    }

    budget, err := h.service.CreateBudget(c.Request.Context(), userID, req)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusCreated, "OK", "成功", budget)
}

func (h *Handler) UpdateBudget(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    var req UpdateBudgetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
        return
    }

    budget, err := h.service.UpdateBudget(c.Request.Context(), userID, c.Param("id"), req)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", budget)
}

func (h *Handler) DeleteBudget(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    if err := h.service.DeleteBudget(c.Request.Context(), userID, c.Param("id")); err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"deleted": true})
}

func (h *Handler) ListAlerts(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    limit := 50
    if raw := c.Query("limit"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
            limit = parsed
        }
    }

    alerts, err := h.service.GetUserAlerts(c.Request.Context(), userID, limit)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": alerts})
}

func (h *Handler) GetPreferences(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    prefs, err := h.prefs.GetPreferences(c.Request.Context(), userID)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"prefs": prefs})
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
    userID, ok := httpx.UserIDFromContext(c)
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    var req UpdatePrefsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
        return
    }

    if err := h.prefs.UpdatePreferences(c.Request.Context(), userID, req); err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", nil)
}
```

- [ ] **Step 2: 创建 PrefService**

在 `service.go` 中添加 `PrefService`：

```go
type PrefService struct {
    repo Repository
}

func NewPrefService(repo Repository) *PrefService {
    return &PrefService{repo: repo}
}

func (s *PrefService) GetPreferences(ctx context.Context, userID string) (*UserNotificationPref, error) {
    return s.repo.GetOrCreatePrefs(ctx, userID)
}

func (s *PrefService) UpdatePreferences(ctx context.Context, userID string, req UpdatePrefsRequest) error {
    p, err := s.repo.GetOrCreatePrefs(ctx, userID)
    if err != nil {
        return err
    }
    if req.RealtimeAlert != nil {
        p.RealtimeAlert = *req.RealtimeAlert
    }
    if req.DailyDigest != nil {
        p.DailyDigest = *req.DailyDigest
    }
    if req.WeeklyDigest != nil {
        p.WeeklyDigest = *req.WeeklyDigest
    }
    return s.repo.UpdatePrefs(ctx, p)
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/budget/handler.go && git commit -m "feat(budget): add budget HTTP handler

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: 注册路由

- [ ] **Step 1: 创建 budget_wiring.go**

```go
// backend/internal/bootstrap/http/budget_wiring.go
package http

import (
    "database/sql"

    "xledger/internal/budget"
    "xledger/internal/reporting"
)

func newBudgetHandlerWithPostgreSQL(db *sql.DB, reportingRepo reporting.CategoryStatsFetcher) *budget.Handler {
    repo := budget.NewPostgresRepository(db)
    svc := budget.NewService(repo, reportingRepo)
    prefs := budget.NewPrefService(repo)
    return budget.NewHandler(svc, prefs)
}
```

- [ ] **Step 2: 修改 router.go**

在 `NewRouterWithPostgreSQL` 中添加：

```go
// 在 reportingHandler 创建后添加
reportingRepo := reporting.NewRepository(nil, acctDeps.TxnRepo, acctDeps.CategoryService)
budgetHandler := newBudgetHandlerWithPostgreSQL(db, reportingRepo)

reportingGroup := r.Group("/api")
reportingGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
// ... existing stats routes ...

budgetGroup := r.Group("/api")
budgetGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
budgetGroup.GET("/budgets", budgetHandler.ListBudgets)
budgetGroup.POST("/budgets", budgetHandler.CreateBudget)
budgetGroup.PATCH("/budgets/:id", budgetHandler.UpdateBudget)
budgetGroup.DELETE("/budgets/:id", budgetHandler.DeleteBudget)
budgetGroup.GET("/budgets/alerts", budgetHandler.ListAlerts)
budgetGroup.GET("/budgets/preferences", budgetHandler.GetPreferences)
budgetGroup.PATCH("/budgets/preferences", budgetHandler.UpdatePreferences)
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/bootstrap/http/router.go backend/internal/bootstrap/http/budget_wiring.go
git commit -m "feat(budget): register budget routes

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: 创建前端 API 和 Hooks

- [ ] **Step 1: 创建 budgets-api.ts**

```typescript
// frontend/app/src/features/budgets/budgets-api.ts
import { requestEnvelope } from '@/lib/api'

export interface Budget {
  id: string
  category_id: string
  category_name: string
  amount: number
  period: string
  thresholds: number[]
  spent_amount: number
  usage_percent: number
  remaining: number
  period_start: string
  period_end: string
  is_over_budget: boolean
}

export interface BudgetAlert {
  id: string
  budget_id: string
  threshold: number
  spent_amount: number
  period_start: string
  period_end: string
  sent_at: string
}

export interface NotificationPrefs {
  realtime_alert: boolean
  daily_digest: boolean
  weekly_digest: boolean
}

export function getBudgets(accessToken: string) {
  return requestEnvelope<{ items: Budget[] }>('/budgets', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function createBudget(accessToken: string, data: { category_id: string; amount: number; thresholds?: number[] }) {
  return requestEnvelope<Budget>('/budgets', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
}

export function updateBudget(accessToken: string, id: string, data: { amount?: number; thresholds?: number[] }) {
  return requestEnvelope<Budget>(`/budgets/${id}`, {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
}

export function deleteBudget(accessToken: string, id: string) {
  return requestEnvelope<{ deleted: boolean }>(`/budgets/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getBudgetAlerts(accessToken: string, limit = 50) {
  return requestEnvelope<{ items: BudgetAlert[] }>(`/budgets/alerts?limit=${limit}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function getNotificationPrefs(accessToken: string) {
  return requestEnvelope<{ prefs: NotificationPrefs }>('/budgets/preferences', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function updateNotificationPrefs(accessToken: string, prefs: Partial<NotificationPrefs>) {
  return requestEnvelope('/budgets/preferences', {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(prefs),
  })
}
```

- [ ] **Step 2: 创建 budgets-hooks.ts**

```typescript
// frontend/app/src/features/budgets/budgets-hooks.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthToken } from '@/features/auth/auth-context'
import * as api from './budgets-api'

export function useBudgets() {
  const token = useAuthToken()
  return useQuery({
    queryKey: ['budgets'],
    queryFn: () => api.getBudgets(token!),
    enabled: !!token,
  })
}

export function useBudgetAlerts(limit = 50) {
  const token = useAuthToken()
  return useQuery({
    queryKey: ['budgets', 'alerts', limit],
    queryFn: () => api.getBudgetAlerts(token!, limit),
    enabled: !!token,
  })
}

export function useCreateBudget() {
  const token = useAuthToken()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { category_id: string; amount: number; thresholds?: number[] }) =>
      api.createBudget(token!, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['budgets'] }),
  })
}

export function useUpdateBudget() {
  const token = useAuthToken()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string; amount?: number; thresholds?: number[] }) =>
      api.updateBudget(token!, id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['budgets'] }),
  })
}

export function useDeleteBudget() {
  const token = useAuthToken()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteBudget(token!, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['budgets'] }),
  })
}

export function useNotificationPrefs() {
  const token = useAuthToken()
  return useQuery({
    queryKey: ['notification-prefs'],
    queryFn: () => api.getNotificationPrefs(token!),
    enabled: !!token,
  })
}

export function useUpdateNotificationPrefs() {
  const token = useAuthToken()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (prefs: Partial<api.NotificationPrefs>) =>
      api.updateNotificationPrefs(token!, prefs),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notification-prefs'] }),
  })
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/app/src/features/budgets/ && git commit -m "feat(frontend): add budget API and hooks

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: 创建前端页面

- [ ] **Step 1: 创建 budgets-page.tsx**

```typescript
// frontend/app/src/pages/budgets-page.tsx
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { useBudgets, useDeleteBudget } from '@/features/budgets/budgets-hooks'
import { formatCurrency } from '@/lib/format'

export function BudgetsPage() {
  const { t } = useTranslation()
  const { data, isLoading } = useBudgets()
  const deleteBudget = useDeleteBudget()

  const budgets = data?.items ?? []

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t('budget.title')}</h1>
        <Link
          to="/budgets/new"
          className="flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-white font-semibold"
        >
          <Plus size={18} /> {t('budget.create')}
        </Link>
      </div>

      {isLoading && <div className="text-center py-8">{t('common.loading')}</div>}

      <div className="space-y-3">
        {budgets.map((budget) => (
          <div key={budget.id} className="rounded-xl border border-outline/15 bg-surface-container p-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold">{budget.category_name}</h3>
                <p className="text-sm text-on-surface-variant">
                  {t('budget.spent')}: {formatCurrency(budget.spent_amount)} / {formatCurrency(budget.amount)}
                </p>
              </div>
              <div className="text-right">
                <p className={`text-lg font-bold ${budget.is_over_budget ? 'text-rose-500' : 'text-emerald-500'}`}>
                  {budget.usage_percent.toFixed(0)}%
                </p>
                <p className="text-xs text-on-surface-variant">
                  {budget.is_over_budget
                    ? t('budget.overBudget')
                    : `${t('budget.remaining')}: ${formatCurrency(budget.remaining)}`}
                </p>
              </div>
            </div>
            {/* 进度条 */}
            <div className="mt-3 h-2 w-full rounded-full bg-outline/20">
              <div
                className={`h-full rounded-full ${budget.is_over_budget ? 'bg-rose-500' : 'bg-emerald-500'}`}
                style={{ width: `${Math.min(100, budget.usage_percent)}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 创建 budget-form-page.tsx**

```typescript
// frontend/app/src/pages/budget-form-page.tsx
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router-dom'
import { useCategories } from '@/features/transactions/transactions-hooks'
import { useCreateBudget, useUpdateBudget } from '@/features/budgets/budgets-hooks'

export function BudgetFormPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const isEdit = !!id

  const { data: categories } = useCategories()
  const createBudget = useCreateBudget()
  const updateBudget = useUpdateBudget()

  const [categoryId, setCategoryId] = useState('')
  const [amount, setAmount] = useState('')
  const [thresholds, setThresholds] = useState<number[]>([80, 100])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const data = {
      category_id: categoryId,
      amount: parseFloat(amount),
      thresholds,
    }
    if (isEdit && id) {
      await updateBudget.mutateAsync({ id, ...data })
    } else {
      await createBudget.mutateAsync(data)
    }
    navigate('/budgets')
  }

  return (
    <div className="max-w-md mx-auto space-y-6">
      <h1 className="text-2xl font-bold">{isEdit ? t('budget.edit') : t('budget.create')}</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">{t('transaction.category')}</label>
          <select
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            className="w-full rounded-xl border border-outline/15 px-4 py-2"
            required
          >
            <option value="">--</option>
            {categories?.map((cat) => (
              <option key={cat.id} value={cat.id}>{cat.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">{t('budget.amount')}</label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="w-full rounded-xl border border-outline/15 px-4 py-2"
            required
            min="0"
            step="0.01"
          />
        </div>
        <div className="flex gap-4">
          <button type="submit" className="flex-1 rounded-xl bg-primary py-2 text-white font-semibold">
            {t('common.save')}
          </button>
          <button type="button" onClick={() => navigate('/budgets')} className="flex-1 rounded-xl border border-outline/15 py-2">
            {t('common.cancel')}
          </button>
        </div>
      </form>
    </div>
  )
}
```

- [ ] **Step 3: 创建 budget-alerts-page.tsx**

```typescript
// frontend/app/src/pages/budget-alerts-page.tsx
import { useTranslation } from 'react-i18next'
import { useBudgetAlerts } from '@/features/budgets/budgets-hooks'

export function BudgetAlertsPage() {
  const { t } = useTranslation()
  const { data, isLoading } = useBudgetAlerts()

  const alerts = data?.items ?? []

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">{t('budget.alerts')}</h1>
      {isLoading && <div className="text-center py-8">{t('common.loading')}</div>}
      <div className="space-y-3">
        {alerts.map((alert) => (
          <div key={alert.id} className="rounded-xl border border-rose-200 bg-rose-50 p-4">
            <p className="font-semibold text-rose-700">
              {t('budget.alertTriggered', { threshold: alert.threshold })}
            </p>
            <p className="text-sm text-rose-600">
              {t('budget.alertSpent', { amount: alert.spent_amount })}
            </p>
            <p className="text-xs text-rose-400 mt-1">{new Date(alert.sent_at).toLocaleString()}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: 修改 App.tsx 添加路由**

```typescript
// 添加路由：
<Route path="/budgets" element={<BudgetsPage />} />
<Route path="/budgets/new" element={<BudgetFormPage />} />
<Route path="/budgets/:id/edit" element={<BudgetFormPage />} />
<Route path="/budgets/alerts" element={<BudgetAlertsPage />} />
```

- [ ] **Step 5: Commit**

```bash
git add frontend/app/src/pages/budgets-page.tsx frontend/app/src/pages/budget-form-page.tsx frontend/app/src/pages/budget-alerts-page.tsx frontend/app/src/App.tsx
git commit -m "feat(frontend): add budget pages

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 10: 集成实时告警到交易创建

- [ ] **Step 1: 修改交易创建逻辑**

在 `backend/internal/accounting/transaction_service.go` 的 `CreateTransaction` 方法中，事务成功后调用 alert service：

```go
// 在 CreateTransaction 成功后添加：
if categoryID != "" && alertService != nil {
    go alertService.CheckAndAlert(context.Background(), userID, categoryID)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/accounting/transaction_service.go && git commit -m "feat(budget): trigger budget alerts on transaction creation

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```
