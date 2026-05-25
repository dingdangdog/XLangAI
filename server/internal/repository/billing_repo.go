package repository

import (
	"context"
	"errors"
	"strings"

	"wlltalk/server/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BillingProduct 商店商品目录行（sys_billing_products）。
type BillingProduct struct {
	ID               string  `json:"id"`
	Code             string  `json:"code"`
	Kind             string  `json:"kind"`
	IOSProductID     string  `json:"ios_product_id"`
	AndroidProductID string  `json:"android_product_id"`
	TierCode         *string `json:"tier_code,omitempty"`
	TokenGrant       int64   `json:"token_grant"`
	SortOrder        int     `json:"sort_order"`
	Status           string  `json:"status"`
}

type BillingProductPublic struct {
	Code             string  `json:"code"`
	Kind             string  `json:"kind"`
	IOSProductID     string  `json:"ios_product_id"`
	AndroidProductID string  `json:"android_product_id"`
	TierCode         *string `json:"tier_code,omitempty"`
	TokenGrant       int64   `json:"token_grant"`
	SortOrder        int     `json:"sort_order"`
}

type BillingRepo struct {
	db *gorm.DB
}

func NewBillingRepo(db *gorm.DB) *BillingRepo {
	return &BillingRepo{db: db}
}

func (r *BillingRepo) ListCatalog(ctx context.Context) ([]BillingProductPublic, error) {
	var rows []entity.BillingProduct
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort_order ASC, code ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]BillingProductPublic, 0, len(rows))
	for _, row := range rows {
		out = append(out, BillingProductPublic{
			Code:             row.Code,
			Kind:             row.Kind,
			IOSProductID:     row.IOSProductID,
			AndroidProductID: row.AndroidProductID,
			TierCode:         row.TierCode,
			TokenGrant:       row.TokenGrant,
			SortOrder:        row.SortOrder,
		})
	}
	return out, nil
}

func (r *BillingRepo) GetProductByStoreSKU(ctx context.Context, productID string) (*BillingProduct, error) {
	pid := strings.TrimSpace(productID)
	if pid == "" {
		return nil, nil
	}
	var row entity.BillingProduct
	err := r.db.WithContext(ctx).
		Where("status = ? AND (ios_product_id = ? OR android_product_id = ?)", "active", pid, pid).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &BillingProduct{
		ID:               row.ID,
		Code:             row.Code,
		Kind:             row.Kind,
		IOSProductID:     row.IOSProductID,
		AndroidProductID: row.AndroidProductID,
		TierCode:         row.TierCode,
		TokenGrant:       row.TokenGrant,
		SortOrder:        row.SortOrder,
		Status:           row.Status,
	}, nil
}

// TryInsertStoreTransaction 幂等插入；若已存在则返回 applied=false。
func (r *BillingRepo) TryInsertStoreTransaction(ctx context.Context, userID, platform, storeTxnID, productCode, raw string) (applied bool, err error) {
	row := entity.StoreTransaction{
		ID:                 uuid.New().String(),
		UserID:             userID,
		Platform:           platform,
		StoreTransactionID: storeTxnID,
		ProductCode:        productCode,
	}
	if strings.TrimSpace(raw) != "" {
		row.RawPayload = &raw
	}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "platform"}, {Name: "store_transaction_id"}},
		DoNothing: true,
	}).Create(&row)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// DeleteStoreTransaction 用于发放失败时回滚幂等行，便于客户端重试。
func (r *BillingRepo) DeleteStoreTransaction(ctx context.Context, platform, storeTxnID string) error {
	return r.db.WithContext(ctx).
		Where("platform = ? AND store_transaction_id = ?", platform, storeTxnID).
		Delete(&entity.StoreTransaction{}).Error
}
