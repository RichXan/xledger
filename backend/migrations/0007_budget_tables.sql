-- Budget tables for per-category monthly budgets

CREATE TABLE budgets (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    category_id TEXT,
    amount NUMERIC(12, 2) NOT NULL,
    period TEXT NOT NULL DEFAULT 'monthly',
    alert_at NUMERIC(5, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);
CREATE INDEX idx_budgets_category_id ON budgets(category_id);

CREATE TABLE budget_alerts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    budget_id TEXT NOT NULL,
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    alert_type TEXT NOT NULL,
    spent_amount NUMERIC(12, 2) NOT NULL,
    budget_amount NUMERIC(12, 2) NOT NULL,
    message TEXT
);

CREATE INDEX idx_budget_alerts_user_id ON budget_alerts(user_id);
CREATE INDEX idx_budget_alerts_budget_id ON budget_alerts(budget_id);
CREATE INDEX idx_budget_alerts_triggered_at ON budget_alerts(triggered_at DESC);

CREATE TABLE user_notification_prefs (
    user_id TEXT PRIMARY KEY,
    realtime_alert BOOLEAN DEFAULT TRUE,
    daily_digest BOOLEAN DEFAULT FALSE,
    weekly_digest BOOLEAN DEFAULT FALSE,
    push_endpoint TEXT,
    push_key TEXT
);
