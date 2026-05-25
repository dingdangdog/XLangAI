package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"xlangai/server/config"
	"xlangai/server/internal/authz"
	"xlangai/server/internal/billing"
	"xlangai/server/internal/repository"

	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	cfg   *config.Config
	apple *billing.AppleConfig
	repo  *repository.BillingRepo
	users *repository.UserRepo
	tiers *repository.MembershipRepo
	az    *authz.Service
}

func NewBillingHandler(cfg *config.Config, apple *billing.AppleConfig, repo *repository.BillingRepo, users *repository.UserRepo, tiers *repository.MembershipRepo, az *authz.Service) *BillingHandler {
	return &BillingHandler{cfg: cfg, apple: apple, repo: repo, users: users, tiers: tiers, az: az}
}

// Catalog 返回 sys_billing_products 中的上架目录（不含密钥）。
func (h *BillingHandler) Catalog(c *gin.Context) {
	list, err := h.repo.ListCatalog(c.Request.Context())
	if err != nil {
		// 表未迁移时给出可读错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": list})
}

type verifyBody struct {
	Platform      string `json:"platform" binding:"required"`
	ProductID     string `json:"product_id" binding:"required"`
	TransactionID string `json:"transaction_id"`
	PurchaseToken string `json:"purchase_token"`
}

// Verify 校验商店单据并发放权益（幂等）。
func (h *BillingHandler) Verify(c *gin.Context) {
	uid := c.GetString("user_id")
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req verifyBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	productID := strings.TrimSpace(req.ProductID)
	ctx := c.Request.Context()

	prod, err := h.repo.GetProductByStoreSKU(ctx, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if prod == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown product_id"})
		return
	}

	switch platform {
	case "ios":
		h.verifyIOS(c, ctx, uid, prod, strings.TrimSpace(req.TransactionID))
	case "android":
		h.verifyAndroid(c, ctx, uid, prod, strings.TrimSpace(req.PurchaseToken))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform must be ios or android"})
	}
}

func (h *BillingHandler) verifyIOS(c *gin.Context, ctx context.Context, uid string, prod *repository.BillingProduct, transactionID string) {
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction_id required for ios"})
		return
	}
	if h.apple == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Apple IAP not configured on server"})
		return
	}
	payload, err := billing.FetchIOSTransaction(ctx, h.apple, transactionID)
	if err != nil {
		log.Printf("billing ios verify: user=%s txn=%s: %v", uid, transactionID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "apple verification failed", "detail": err.Error()})
		return
	}
	if strings.TrimSpace(payload.BundleID) != "" {
		want := strings.TrimSpace(h.cfg.AppleBundleID)
		if want != "" && strings.TrimSpace(payload.BundleID) != want {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bundle id mismatch"})
			return
		}
	}
	if strings.TrimSpace(payload.ProductID) != strings.TrimSpace(prod.IOSProductID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product id mismatch"})
		return
	}
	storeTxnID := strings.TrimSpace(payload.TransactionID)
	if storeTxnID == "" {
		storeTxnID = transactionID
	}
	raw := billing.MustStringJSON(payload)
	applied, err := h.applyPurchase(ctx, uid, "ios", storeTxnID, prod, raw, func() error {
		return h.applyIOSProductEffects(ctx, uid, prod, payload)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.respondVerifySuccess(c, ctx, uid, applied)
}

func (h *BillingHandler) respondVerifySuccess(c *gin.Context, ctx context.Context, uid string, applied bool) {
	if applied && h.az != nil {
		h.az.InvalidatePrincipal(ctx, uid)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "applied": applied})
}

func (h *BillingHandler) applyIOSProductEffects(ctx context.Context, uid string, prod *repository.BillingProduct, payload *billing.IOSignedTransactionPayload) error {
	switch strings.ToLower(strings.TrimSpace(prod.Kind)) {
	case "consumable":
		if prod.TokenGrant <= 0 {
			return nil
		}
		return h.users.AddTokenBalance(ctx, uid, prod.TokenGrant)
	case "subscription":
		if prod.TierCode == nil || strings.TrimSpace(*prod.TierCode) == "" {
			return nil
		}
		tid, err := h.tiers.GetTierIDByCode(ctx, strings.TrimSpace(*prod.TierCode))
		if err != nil {
			return err
		}
		if tid == nil {
			return nil
		}
		var exp *time.Time
		if payload.ExpiresDate > 0 {
			t := time.UnixMilli(payload.ExpiresDate).UTC()
			exp = &t
		}
		return h.users.ApplySubscriptionEntitlement(ctx, uid, tid, exp)
	default:
		return nil
	}
}

func (h *BillingHandler) verifyAndroid(c *gin.Context, ctx context.Context, uid string, prod *repository.BillingProduct, purchaseToken string) {
	if purchaseToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "purchase_token required for android"})
		return
	}
	sa := strings.TrimSpace(h.cfg.GooglePlayServiceAccountJSON)
	if sa == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Play not configured on server"})
		return
	}
	jsonBytes, err := billing.LoadGoogleServiceAccountJSON(sa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	saStr := string(jsonBytes)
	pkg := strings.TrimSpace(h.cfg.GooglePlayPackageName)
	if pkg == "" {
		pkg = "com.xlangai.android"
	}

	switch strings.ToLower(strings.TrimSpace(prod.Kind)) {
	case "consumable":
		gp, err := billing.VerifyAndroidProductPurchase(ctx, pkg, saStr, prod.AndroidProductID, purchaseToken)
		if err != nil {
			log.Printf("billing android product: user=%s: %v", uid, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "google play verification failed", "detail": err.Error()})
			return
		}
		if gp.PurchaseState != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "purchase not in purchased state"})
			return
		}
		storeTxnID := strings.TrimSpace(gp.OrderId)
		if storeTxnID == "" {
			storeTxnID = purchaseToken + ":" + prod.AndroidProductID
		}
		raw := billing.MustStringJSON(gp)
		applied, err := h.applyPurchase(ctx, uid, "android", storeTxnID, prod, raw, func() error {
			if prod.TokenGrant <= 0 {
				return nil
			}
			return h.users.AddTokenBalance(ctx, uid, prod.TokenGrant)
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.respondVerifySuccess(c, ctx, uid, applied)
	case "subscription":
		sv2, err := billing.VerifyAndroidSubscriptionV2(ctx, pkg, saStr, purchaseToken)
		if err != nil {
			log.Printf("billing android sub: user=%s: %v", uid, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "google play subscription verification failed", "detail": err.Error()})
			return
		}
		if st := strings.TrimSpace(sv2.SubscriptionState); st != "" && st != "SUBSCRIPTION_STATE_ACTIVE" && st != "SUBSCRIPTION_STATE_IN_GRACE_PERIOD" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "subscription not active", "state": st})
			return
		}
		// 校验 line item 的 productId
		match := false
		for _, li := range sv2.LineItems {
			if strings.TrimSpace(li.ProductId) == strings.TrimSpace(prod.AndroidProductID) {
				match = true
				break
			}
		}
		if !match && len(sv2.LineItems) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product id mismatch in subscription"})
			return
		}
		storeTxnID := strings.TrimSpace(sv2.LatestOrderId)
		if storeTxnID == "" {
			storeTxnID = purchaseToken
		}
		raw := billing.MustStringJSON(sv2)
		applied, err := h.applyPurchase(ctx, uid, "android", storeTxnID, prod, raw, func() error {
			if prod.TierCode == nil {
				return nil
			}
			tid, err := h.tiers.GetTierIDByCode(ctx, strings.TrimSpace(*prod.TierCode))
			if err != nil {
				return err
			}
			if tid == nil {
				return nil
			}
			expTime, err := billing.SubscriptionExpiryFromV2(sv2)
			if err != nil {
				return err
			}
			exp := expTime.UTC()
			return h.users.ApplySubscriptionEntitlement(ctx, uid, tid, &exp)
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.respondVerifySuccess(c, ctx, uid, applied)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown product kind"})
	}
}

func (h *BillingHandler) applyPurchase(ctx context.Context, uid, platform, storeTxnID string, prod *repository.BillingProduct, raw string, grant func() error) (applied bool, err error) {
	applied, err = h.repo.TryInsertStoreTransaction(ctx, uid, platform, storeTxnID, prod.Code, raw)
	if err != nil {
		return false, err
	}
	if !applied {
		return false, nil
	}
	if err := grant(); err != nil {
		_ = h.repo.DeleteStoreTransaction(ctx, platform, storeTxnID)
		return false, err
	}
	return true, nil
}
