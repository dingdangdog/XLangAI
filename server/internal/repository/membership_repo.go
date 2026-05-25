package repository

import (
	"context"
	"errors"
	"strings"

	"wlltalk/server/internal/entity"

	"gorm.io/gorm"
)

type MembershipTier struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	DailyLimit   int     `json:"daily_limit"`
	MonthlyLimit int     `json:"monthly_limit"`
	FeaturesJSON *string `json:"-"`
	Status       string  `json:"status"`
}

type MembershipRepo struct {
	db *gorm.DB
}

func NewMembershipRepo(db *gorm.DB) *MembershipRepo {
	return &MembershipRepo{db: db}
}

func tierFromEntity(t *entity.MembershipTier) *MembershipTier {
	if t == nil {
		return nil
	}
	out := &MembershipTier{
		ID:           t.ID,
		Code:         t.Code,
		Name:         t.Name,
		FeaturesJSON: t.Features,
		Status:       t.Status,
	}
	if t.DailyLimit != nil {
		out.DailyLimit = *t.DailyLimit
	}
	if t.MonthlyLimit != nil {
		out.MonthlyLimit = *t.MonthlyLimit
	}
	return out
}

func (r *MembershipRepo) GetByID(ctx context.Context, tierID string) (*MembershipTier, error) {
	if tierID == "" {
		return nil, nil
	}
	var row entity.MembershipTier
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", tierID, "active").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return tierFromEntity(&row), nil
}

// TierPublic 对外展示的会员档位（不含 features 等内部字段）。
type TierPublic struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	DailyLimit   int    `json:"daily_limit"`
	MonthlyLimit int    `json:"monthly_limit"`
	SortOrder    int    `json:"sort_order"`
}

func (r *MembershipRepo) ListPublic(ctx context.Context) ([]TierPublic, error) {
	var rows []entity.MembershipTier
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, code ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]TierPublic, 0, len(rows))
	for _, row := range rows {
		t := TierPublic{
			ID:        row.ID,
			Code:      row.Code,
			Name:      row.Name,
			SortOrder: row.SortOrder,
		}
		if row.DailyLimit != nil {
			t.DailyLimit = *row.DailyLimit
		}
		if row.MonthlyLimit != nil {
			t.MonthlyLimit = *row.MonthlyLimit
		}
		out = append(out, t)
	}
	return out, nil
}

func (r *MembershipRepo) GetTierIDByCode(ctx context.Context, code string) (*string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}
	var row entity.MembershipTier
	err := r.db.WithContext(ctx).
		Select("id").
		Where("code = ? AND status = ?", code, "active").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row.ID, nil
}

func (r *MembershipRepo) GetFreeTierID(ctx context.Context) (*string, error) {
	return r.GetTierIDByCode(ctx, "free")
}
