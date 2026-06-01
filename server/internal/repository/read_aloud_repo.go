package repository

import (
	"context"
	"strings"
	"time"

	"xlangai/server/internal/entity"
	"xlangai/server/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReadAloudRepo struct {
	db *gorm.DB
}

func NewReadAloudRepo(db *gorm.DB) *ReadAloudRepo {
	return &ReadAloudRepo{db: db}
}

func categoryToModel(c *entity.ReadAloudCategory) *model.ReadAloudCategory {
	if c == nil {
		return nil
	}
	out := &model.ReadAloudCategory{
		ID:        c.ID,
		Code:      c.Code,
		Name:      c.Name,
		SortOrder: c.SortOrder,
	}
	if c.NameEn != nil {
		out.NameEn = strings.TrimSpace(*c.NameEn)
	}
	if c.Icon != nil {
		out.Icon = strings.TrimSpace(*c.Icon)
	}
	if c.Description != nil {
		out.Description = strings.TrimSpace(*c.Description)
	}
	if c.DescriptionEn != nil {
		out.DescriptionEn = strings.TrimSpace(*c.DescriptionEn)
	}
	return out
}

func vocabToModel(v *entity.ReadAloudVocabulary) *model.ReadAloudVocabulary {
	if v == nil {
		return nil
	}
	return &model.ReadAloudVocabulary{
		ID:               v.ID,
		CategoryID:       v.CategoryID,
		LanguageID:       v.LanguageID,
		Word:             v.Word,
		ExampleSentence:  v.ExampleSentence,
		VoiceRoleID:      v.VoiceRoleID,
		WordAudioURL:     v.WordAudioURL,
		SentenceAudioURL: v.SentenceAudioURL,
		SortOrder:        v.SortOrder,
	}
}

func sessionToModel(s *entity.ReadAloudSession) *model.ReadAloudSession {
	if s == nil {
		return nil
	}
	return &model.ReadAloudSession{
		ID:             s.ID,
		UserID:         s.UserID,
		CategoryID:     s.CategoryID,
		LanguageID:     s.LanguageID,
		Status:         s.Status,
		TotalItems:     s.TotalItems,
		CompletedItems: s.CompletedItems,
		AverageScore:   s.AverageScore,
		StartedAt:      s.StartedAt,
		CompletedAt:    s.CompletedAt,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

// SessionEntityToModel exports session entity for handlers.
func SessionEntityToModel(s *entity.ReadAloudSession) *model.ReadAloudSession {
	return sessionToModel(s)
}

func attemptToModel(a *entity.ReadAloudAttempt) *model.ReadAloudAttempt {
	if a == nil {
		return nil
	}
	return &model.ReadAloudAttempt{
		ID:            a.ID,
		SessionID:     a.SessionID,
		VocabularyID:  a.VocabularyID,
		Part:          a.Part,
		ReferenceText: a.ReferenceText,
		Transcript:    a.Transcript,
		Score:         a.Score,
		MatchDetail:   a.MatchDetail,
		DurationMs:    a.DurationMs,
		CreatedAt:     a.CreatedAt,
	}
}

func (r *ReadAloudRepo) ListActiveCategories(ctx context.Context) ([]*model.ReadAloudCategory, error) {
	var rows []entity.ReadAloudCategory
	if err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.ReadAloudCategory, 0, len(rows))
	for i := range rows {
		out = append(out, categoryToModel(&rows[i]))
	}
	return out, nil
}

func (r *ReadAloudRepo) GetCategoryLocale(
	ctx context.Context,
	categoryID, languageID string,
) (*entity.ReadAloudCategoryLocale, error) {
	categoryID = strings.TrimSpace(categoryID)
	languageID = strings.TrimSpace(languageID)
	if categoryID == "" || languageID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var row entity.ReadAloudCategoryLocale
	err := r.db.WithContext(ctx).
		Where("category_id = ? AND language_id = ?", categoryID, languageID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReadAloudRepo) GetCategoryByID(ctx context.Context, id string) (*entity.ReadAloudCategory, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var row entity.ReadAloudCategory
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReadAloudRepo) CountVocabularies(ctx context.Context, categoryID, languageID string) (int, error) {
	var n int64
	q := r.db.WithContext(ctx).Model(&entity.ReadAloudVocabulary{}).
		Where("status = ?", "active")
	if categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}
	if languageID != "" {
		q = q.Where("language_id = ?", languageID)
	}
	if err := q.Count(&n).Error; err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *ReadAloudRepo) ListVocabularies(ctx context.Context, categoryID, languageID string) ([]*model.ReadAloudVocabulary, error) {
	categoryID = strings.TrimSpace(categoryID)
	languageID = strings.TrimSpace(languageID)
	if categoryID == "" || languageID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var rows []entity.ReadAloudVocabulary
	if err := r.db.WithContext(ctx).
		Where("category_id = ? AND language_id = ? AND status = ?", categoryID, languageID, "active").
		Order("sort_order ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.ReadAloudVocabulary, 0, len(rows))
	for i := range rows {
		out = append(out, vocabToModel(&rows[i]))
	}
	return out, nil
}

type CreateReadAloudSessionInput struct {
	UserID     string
	CategoryID string
	LanguageID string
	TotalItems int
}

func (r *ReadAloudRepo) CreateSession(ctx context.Context, in CreateReadAloudSessionInput) (*model.ReadAloudSession, error) {
	now := time.Now().UTC()
	row := entity.ReadAloudSession{
		ID:             uuid.NewString(),
		UserID:         in.UserID,
		CategoryID:     in.CategoryID,
		LanguageID:     in.LanguageID,
		Status:         "in_progress",
		TotalItems:     in.TotalItems,
		CompletedItems: 0,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return sessionToModel(&row), nil
}

func (r *ReadAloudRepo) GetSession(ctx context.Context, sessionID, userID string) (*entity.ReadAloudSession, error) {
	sessionID = strings.TrimSpace(sessionID)
	userID = strings.TrimSpace(userID)
	if sessionID == "" || userID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var row entity.ReadAloudSession
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReadAloudRepo) ListSessions(ctx context.Context, userID string, limit int) ([]*model.ReadAloudSession, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []entity.ReadAloudSession
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.ReadAloudSession, 0, len(rows))
	for i := range rows {
		out = append(out, sessionToModel(&rows[i]))
	}
	return out, nil
}

type CreateReadAloudAttemptInput struct {
	SessionID     string
	VocabularyID  string
	Part          string
	ReferenceText string
	Transcript    string
	Score         int
	MatchDetail   *string
	DurationMs    *int
}

func (r *ReadAloudRepo) CreateAttempt(ctx context.Context, in CreateReadAloudAttemptInput) (*model.ReadAloudAttempt, error) {
	if in.Score < 0 {
		in.Score = 0
	}
	if in.Score > 100 {
		in.Score = 100
	}
	row := entity.ReadAloudAttempt{
		ID:            uuid.NewString(),
		SessionID:     in.SessionID,
		VocabularyID:  in.VocabularyID,
		Part:          in.Part,
		ReferenceText: in.ReferenceText,
		Transcript:    in.Transcript,
		Score:         in.Score,
		MatchDetail:   in.MatchDetail,
		DurationMs:    in.DurationMs,
		CreatedAt:     time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return attemptToModel(&row), nil
}

func (r *ReadAloudRepo) ListAttempts(ctx context.Context, sessionID string) ([]*model.ReadAloudAttempt, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, nil
	}
	var rows []entity.ReadAloudAttempt
	if err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*model.ReadAloudAttempt, 0, len(rows))
	for i := range rows {
		out = append(out, attemptToModel(&rows[i]))
	}
	return out, nil
}

func (r *ReadAloudRepo) RefreshSessionProgress(ctx context.Context, sessionID string) (*model.ReadAloudSession, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var session entity.ReadAloudSession
	if err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}

	type aggRow struct {
		AttemptCount int
		AvgScore     float64
	}
	var agg aggRow
	if err := r.db.WithContext(ctx).Model(&entity.ReadAloudAttempt{}).
		Select("COUNT(*) AS attempt_count, COALESCE(AVG(score), 0) AS avg_score").
		Where("session_id = ?", sessionID).
		Scan(&agg).Error; err != nil {
		return nil, err
	}

	// 每个词汇 2 步（word + sentence）算一次完成
	completedItems := agg.AttemptCount / 2
	if completedItems > session.TotalItems {
		completedItems = session.TotalItems
	}

	avgScore := int(agg.AvgScore + 0.5)
	updates := map[string]interface{}{
		"completed_items": completedItems,
		"average_score":   avgScore,
		"updated_at":      time.Now().UTC(),
	}
	if completedItems >= session.TotalItems && session.TotalItems > 0 {
		now := time.Now().UTC()
		updates["status"] = "completed"
		updates["completed_at"] = now
	}

	if err := r.db.WithContext(ctx).Model(&entity.ReadAloudSession{}).
		Where("id = ?", sessionID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}
	return sessionToModel(&session), nil
}
