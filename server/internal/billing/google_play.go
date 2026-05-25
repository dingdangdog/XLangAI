package billing

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// VerifyAndroidProductPurchase 校验消耗型商品购买（inapp product）。
func VerifyAndroidProductPurchase(ctx context.Context, packageName, serviceAccountJSON, productID, purchaseToken string) (*androidpublisher.ProductPurchase, error) {
	if strings.TrimSpace(packageName) == "" || strings.TrimSpace(serviceAccountJSON) == "" {
		return nil, errors.New("google play: missing package or service account")
	}
	svc, err := androidpublisher.NewService(ctx, option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	if err != nil {
		return nil, err
	}
	return svc.Purchases.Products.Get(packageName, productID, purchaseToken).Context(ctx).Do()
}

// VerifyAndroidSubscriptionV2 校验订阅（Billing Library 5+ / subscriptionsv2）。
func VerifyAndroidSubscriptionV2(ctx context.Context, packageName, serviceAccountJSON, purchaseToken string) (*androidpublisher.SubscriptionPurchaseV2, error) {
	if strings.TrimSpace(packageName) == "" || strings.TrimSpace(serviceAccountJSON) == "" {
		return nil, errors.New("google play: missing package or service account")
	}
	svc, err := androidpublisher.NewService(ctx, option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	if err != nil {
		return nil, err
	}
	return svc.Purchases.Subscriptionsv2.Get(packageName, purchaseToken).Context(ctx).Do()
}

// LoadGoogleServiceAccountJSON 从路径或内联 JSON 读取服务账号。
func LoadGoogleServiceAccountJSON(pathOrJSON string) ([]byte, error) {
	s := strings.TrimSpace(pathOrJSON)
	if s == "" {
		return nil, errors.New("empty google service account config")
	}
	if strings.HasPrefix(s, "{") {
		return []byte(s), nil
	}
	return os.ReadFile(s)
}

// SubscriptionExpiryFromV2 从 SubscriptionPurchaseV2 解析最近一条 line item 的到期时间（RFC3339）。
func SubscriptionExpiryFromV2(v *androidpublisher.SubscriptionPurchaseV2) (time.Time, error) {
	if v == nil || len(v.LineItems) == 0 {
		return time.Time{}, errors.New("google play: no subscription line items")
	}
	exp := strings.TrimSpace(v.LineItems[0].ExpiryTime)
	if exp == "" {
		return time.Time{}, errors.New("google play: missing expiryTime")
	}
	return time.Parse(time.RFC3339, exp)
}

// MustStringJSON 用于日志落库时截断。
func MustStringJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	if len(b) > 8000 {
		return string(b[:8000]) + "...(truncated)"
	}
	return string(b)
}
