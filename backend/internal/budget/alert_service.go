package budget

import (
    "context"

    "xledger/backend/internal/push"
)

type AlertService struct {
    pushService *push.Service
    repo        Repository
}

func NewAlertService(pushService *push.Service, repo Repository) *AlertService {
    return &AlertService{
        pushService: pushService,
        repo:        repo,
    }
}

func (s *AlertService) SendBudgetAlert(ctx context.Context, userID string, alert *BudgetAlert) {
    pref, err := s.repo.GetPreference(ctx, userID)
    if err != nil || !pref.RealtimeAlert {
        return
    }

    if s.pushService != nil {
        _ = s.pushService.SendPushNotification(userID, "Budget Alert", alert.Message, "budget-alert")
    }
}

func (s *AlertService) SendDailyDigest(ctx context.Context, userID string) {
    pref, err := s.repo.GetPreference(ctx, userID)
    if err != nil || !pref.DailyDigest {
        return
    }

    // Send daily summary push
    if s.pushService != nil {
        _ = s.pushService.SendPushNotification(userID, "Daily Budget Summary", "Your daily spending summary is ready", "daily-digest")
    }
}

func (s *AlertService) SendWeeklyDigest(ctx context.Context, userID string) {
    pref, err := s.repo.GetPreference(ctx, userID)
    if err != nil || !pref.WeeklyDigest {
        return
    }

    // Send weekly summary push
    if s.pushService != nil {
        _ = s.pushService.SendPushNotification(userID, "Weekly Budget Summary", "Your weekly spending summary is ready", "weekly-digest")
    }
}
