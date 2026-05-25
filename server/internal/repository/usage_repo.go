package repository

import (
	"context"
	"errors"
	"time"

	"xlangai/server/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ServiceUsageLLM       = "llm"
	ServiceUsageTTS       = "tts"
	ServiceUsageTranslate = "translate"
	ServiceUsageSTT       = "stt"
)

type UsageRepo struct {
	db *gorm.DB
}

func NewUsageRepo(db *gorm.DB) *UsageRepo {
	return &UsageRepo{db: db}
}

func utcToday() time.Time {
	now := time.Now().UTC()
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// UsageRecord 单次成功调用的用户日聚合与服务配置日聚合增量。
type UsageRecord struct {
	UserID string

	ChatTurns       int
	LLMTokens       int
	TranslateCalls  int
	TranslateChars  int
	TTSCalls        int
	TTSChars        int
	STTCalls        int
	STTAudioBytes   int64
	ServiceType     string
	ServiceConfigID string
	ServiceRequests int
	ServiceUnits    int64
}

func (r *UsageRepo) TodayUsageCount(ctx context.Context, userID string) (int, error) {
	var row entity.UserUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date = ?", userID, utcToday()).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.UsageCount, nil
}

func (r *UsageRepo) MonthUsageCount(ctx context.Context, userID string) (int, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var total int
	err := r.db.WithContext(ctx).Model(&entity.UserUsage{}).
		Select("COALESCE(SUM(usage_count), 0)").
		Where("user_id = ? AND date >= ?", userID, monthStart).
		Scan(&total).Error
	return total, err
}

// DailyCountsInMonth 返回指定 UTC 年月内每日对话轮次（usage_count），键为 YYYY-MM-DD。
func (r *UsageRepo) DailyCountsInMonth(ctx context.Context, userID string, year, month int) (map[string]int, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	var rows []entity.UserUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date >= ? AND date < ?", userID, start, end).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string]int{}
	for _, row := range rows {
		if row.UsageCount > 0 {
			out[row.Date.Format("2006-01-02")] = row.UsageCount
		}
	}
	return out, nil
}

// AddChatTurn 将当日 usage_count +1（对话轮次计一次）。
func (r *UsageRepo) AddChatTurn(ctx context.Context, userID string) error {
	return r.Record(ctx, UsageRecord{
		UserID:    userID,
		ChatTurns: 1,
	})
}

// Record 写入用户日用量与服务配置日用量（仅累加大于 0 的字段）。
func (r *UsageRepo) Record(ctx context.Context, rec UsageRecord) error {
	if rec.UserID == "" {
		return nil
	}
	hasUser := rec.ChatTurns > 0 || rec.LLMTokens > 0 || rec.TranslateCalls > 0 ||
		rec.TranslateChars > 0 || rec.TTSCalls > 0 || rec.TTSChars > 0 ||
		rec.STTCalls > 0 || rec.STTAudioBytes > 0
	hasService := rec.ServiceType != "" && rec.ServiceConfigID != "" &&
		(rec.ServiceRequests > 0 || rec.ServiceUnits > 0)
	if !hasUser && !hasService {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if hasUser {
			if err := r.upsertUserUsage(ctx, tx, rec); err != nil {
				return err
			}
		}
		if hasService {
			if err := r.upsertServiceUsage(ctx, tx, rec); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UsageRepo) upsertUserUsage(ctx context.Context, tx *gorm.DB, rec UsageRecord) error {
	today := utcToday()
	row := entity.UserUsage{
		ID:             uuid.New().String(),
		UserID:         rec.UserID,
		Date:           today,
		UsageCount:     rec.ChatTurns,
		TokenCount:     rec.LLMTokens,
		TranslateCount: rec.TranslateCalls,
		TranslateChars: rec.TranslateChars,
		TtsCount:       rec.TTSCalls,
		TtsChars:       rec.TTSChars,
		SttCount:       rec.STTCalls,
		SttAudioBytes:  rec.STTAudioBytes,
	}
	updates := map[string]interface{}{
		"updated_at": gorm.Expr("now()"),
	}
	if rec.ChatTurns > 0 {
		updates["usage_count"] = gorm.Expr("usr_user_usage.usage_count + ?", rec.ChatTurns)
	}
	if rec.LLMTokens > 0 {
		updates["token_count"] = gorm.Expr("usr_user_usage.token_count + ?", rec.LLMTokens)
	}
	if rec.TranslateCalls > 0 {
		updates["translate_count"] = gorm.Expr("usr_user_usage.translate_count + ?", rec.TranslateCalls)
	}
	if rec.TranslateChars > 0 {
		updates["translate_chars"] = gorm.Expr("usr_user_usage.translate_chars + ?", rec.TranslateChars)
	}
	if rec.TTSCalls > 0 {
		updates["tts_count"] = gorm.Expr("usr_user_usage.tts_count + ?", rec.TTSCalls)
	}
	if rec.TTSChars > 0 {
		updates["tts_chars"] = gorm.Expr("usr_user_usage.tts_chars + ?", rec.TTSChars)
	}
	if rec.STTCalls > 0 {
		updates["stt_count"] = gorm.Expr("usr_user_usage.stt_count + ?", rec.STTCalls)
	}
	if rec.STTAudioBytes > 0 {
		updates["stt_audio_bytes"] = gorm.Expr("usr_user_usage.stt_audio_bytes + ?", rec.STTAudioBytes)
	}
	return tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(&row).Error
}

func (r *UsageRepo) upsertServiceUsage(ctx context.Context, tx *gorm.DB, rec UsageRecord) error {
	today := utcToday()
	row := entity.ServiceUsageDaily{
		ID:           uuid.New().String(),
		Date:         today,
		ServiceType:  rec.ServiceType,
		ConfigID:     rec.ServiceConfigID,
		RequestCount: rec.ServiceRequests,
		UnitCount:    rec.ServiceUnits,
	}
	updates := map[string]interface{}{
		"updated_at": gorm.Expr("now()"),
	}
	if rec.ServiceRequests > 0 {
		updates["request_count"] = gorm.Expr("sys_service_usage_daily.request_count + ?", rec.ServiceRequests)
	}
	if rec.ServiceUnits > 0 {
		updates["unit_count"] = gorm.Expr("sys_service_usage_daily.unit_count + ?", rec.ServiceUnits)
	}
	return tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "date"},
			{Name: "service_type"},
			{Name: "config_id"},
		},
		DoUpdates: clause.Assignments(updates),
	}).Create(&row).Error
}
