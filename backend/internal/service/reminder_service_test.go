package service_test

import (
	"strings"
	"testing"
	"time"

	"construction-hrms/backend/internal/service"
)

func TestDailyReminderMessage(t *testing.T) {
	now := time.Date(2026, time.August, 18, 9, 0, 0, 0, time.FixedZone("CST", 8*3600))
	empty := service.DailyReminderMessage(nil, now)
	if empty != "【每日到期提醒】2026-08-18\n今日无待处理到期事项。" {
		t.Fatalf("unexpected empty message: %q", empty)
	}

	message := service.DailyReminderMessage([]service.ReminderItem{
		{Status: "pending", Type: "certificate_expiry", TalentName: "李四", Subject: "一级建造师", DueDate: "2026-08-20", Level: "urgent", DaysRemaining: 2},
		{Status: "pending", Type: "contract_expiry", TalentName: "张三", Subject: "年度合同", DueDate: "2026-08-16", Level: "expired", DaysRemaining: -2},
		{Status: "handled", Type: "contract_expiry", TalentName: "王五", Subject: "不应出现", DueDate: "2026-08-19", Level: "urgent", DaysRemaining: 1},
	}, now)

	if !strings.Contains(message, "待处理 2 项") || strings.Contains(message, "王五") {
		t.Fatalf("unexpected pending filtering: %q", message)
	}
	if strings.Index(message, "张三") > strings.Index(message, "李四") {
		t.Fatalf("expired reminder should be listed first: %q", message)
	}
	if !strings.Contains(message, "已逾期 2 天") || !strings.Contains(message, "剩余 2 天") {
		t.Fatalf("missing day labels: %q", message)
	}
}
