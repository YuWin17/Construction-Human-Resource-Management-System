package service

import (
	"construction-hrms/backend/internal/domain"
	"context"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type ReminderItem struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	TalentID      string `json:"talent_id"`
	TalentName    string `json:"talent_name"`
	Subject       string `json:"subject"`
	DueDate       string `json:"due_date"`
	Status        string `json:"status"`
	Level         string `json:"level"`
	DaysRemaining int    `json:"days_remaining"`
}
type ReminderService struct{ db *gorm.DB }

func NewReminderService(db *gorm.DB) *ReminderService { return &ReminderService{db: db} }
func (s *ReminderService) DB() *gorm.DB               { return s.db }
func (s *ReminderService) settings(ctx context.Context) (int, int, error) {
	defaults := map[string]string{"contract_reminder_days": "30", "certificate_reminder_days": "30"}
	for k, v := range defaults {
		s.db.WithContext(ctx).Where(domain.SystemSetting{Key: k}).FirstOrCreate(&domain.SystemSetting{Key: k, Value: v})
	}
	var rows []domain.SystemSetting
	err := s.db.WithContext(ctx).Find(&rows).Error
	c, cert := 30, 30
	for _, r := range rows {
		n, _ := strconv.Atoi(r.Value)
		if n > 0 {
			if r.Key == "contract_reminder_days" {
				c = n
			}
			if r.Key == "certificate_reminder_days" {
				cert = n
			}
		}
	}
	return c, cert, err
}
func (s *ReminderService) Settings(ctx context.Context) (int, int, error) { return s.settings(ctx) }
func (s *ReminderService) UpdateSettings(ctx context.Context, contractDays, certificateDays int, adminID string) error {
	if contractDays < 1 || contractDays > 365 || certificateDays < 1 || certificateDays > 365 {
		return ErrValidation
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for k, v := range map[string]int{"contract_reminder_days": contractDays, "certificate_reminder_days": certificateDays} {
			if err := tx.Save(&domain.SystemSetting{Key: k, Value: strconv.Itoa(v), UpdatedByAdminID: adminID}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "settings.reminders_updated", ResourceType: "settings", DisplayName: "到期提醒设置", Summary: "更新提前提醒天数"}).Error
	})
}
func (s *ReminderService) List(ctx context.Context, status string) ([]ReminderItem, error) {
	cd, xd, err := s.settings(ctx)
	if err != nil {
		return nil, err
	}
	var contracts []domain.Contract
	var certs []domain.Certificate
	var talents []domain.Talent
	var deliveryOrders []domain.DeliveryOrder
	var companies []domain.Company
	s.db.WithContext(ctx).Where("status = ? AND end_date <> ''", domain.ContractStatusActive).Find(&contracts)
	s.db.WithContext(ctx).Where("expires_on <> ''").Find(&certs)
	s.db.WithContext(ctx).Find(&talents)
	s.db.WithContext(ctx).Where("status IN ? AND contract_expires_on <> ''", []string{"signed", "expired"}).Preload("Talents").Find(&deliveryOrders)
	s.db.WithContext(ctx).Find(&companies)
	tm := map[string]string{}
	cm := map[string]string{}
	for _, t := range talents {
		tm[t.ID] = t.Name
	}
	for _, company := range companies {
		cm[company.ID] = company.Name
	}
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	items := []ReminderItem{}
	add := func(kind, id, tid, talentName, subject, due string, days int) {
		d, e := time.Parse("2006-01-02", due)
		if e != nil {
			return
		}
		remain := int(d.Sub(now.Truncate(24*time.Hour)).Hours() / 24)
		if remain > days {
			return
		}
		var r domain.Reminder
		q := s.db.WithContext(ctx).Where("reminder_type = ? AND source_id = ?", kind, id).First(&r)
		if q.Error != nil {
			r = domain.Reminder{ReminderType: kind, SourceID: id, TalentID: tid, DueDate: due, Status: "pending"}
			s.db.WithContext(ctx).Create(&r)
		} else if r.DueDate != due {
			// A changed signing date requires a fresh handling decision.
			r.DueDate = due
			r.Status = "pending"
			r.HandledAt = nil
			r.HandledByAdminID = ""
			s.db.WithContext(ctx).Save(&r)
		}
		if status != "" && r.Status != status {
			return
		}
		level := "normal"
		if remain <= 0 {
			level = "expired"
		} else if remain <= 7 {
			level = "urgent"
		}
		if talentName == "" {
			talentName = tm[tid]
		}
		items = append(items, ReminderItem{ID: r.ID, Type: kind, TalentID: tid, TalentName: talentName, Subject: subject, DueDate: due, Status: r.Status, Level: level, DaysRemaining: remain})
	}
	for _, c := range contracts {
		add("contract_expiry", c.ID, c.TalentID, "", c.CompanyName+" / "+c.ContractType, c.EndDate, cd)
	}
	for _, c := range certs {
		add("certificate_expiry", c.ID, c.TalentID, "", c.CertificateNameSnapshot, c.ExpiresOn, xd)
	}
	for _, order := range deliveryOrders {
		names := make([]string, 0, len(order.Talents))
		for _, item := range order.Talents {
			if name := tm[item.TalentID]; name != "" {
				names = append(names, name)
			}
		}
		subject := order.Code
		if companyName := cm[order.CompanyID]; companyName != "" {
			subject += " / " + companyName
		} else if order.RegistrationUnitName != "" {
			subject += " / " + order.RegistrationUnitName
		}
		add("delivery_order_expiry", order.ID, "", strings.Join(names, "、"), subject, order.ContractExpiresOn, cd)
	}
	return items, nil
}
func (s *ReminderService) Handle(ctx context.Context, id, status, adminID string) error {
	if status != "handled" && status != "ignored" {
		return ErrValidation
	}
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Reminder{}).Where("id = ?", id).Updates(map[string]any{"status": status, "handled_at": now, "handled_by_admin_id": adminID}).Error; err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "reminder." + status, ResourceType: "reminder", ResourceID: id, DisplayName: "到期提醒", Summary: fmt.Sprintf("标记为%s", status)}).Error
	})
}
