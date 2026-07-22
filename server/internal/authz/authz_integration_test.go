//go:build integration

package authz

import (
	"context"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"xlangai/server/internal/db"
	"xlangai/server/internal/entity"
	"xlangai/server/internal/repository"
)

// TestQuotaStrategyIntegration 在真实 PostgreSQL 事务中验证额度优先级。
// 全部测试数据在结束时回滚，不会写入或修改真实用户数据。
func TestQuotaStrategyIntegration(t *testing.T) {
	databaseURL := postgresDatabaseURL(os.Getenv("QUOTA_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("QUOTA_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	gdb, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("get sql database: %v", err)
	}
	defer sqlDB.Close()

	tx := gdb.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	dailyLimit := 100
	monthlyLimit := 2
	tierID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()
	tier := entity.MembershipTier{
		ID:           tierID,
		Code:         "quota-test-" + uuid.NewString()[:8],
		Name:         "Quota integration test",
		DailyLimit:   &dailyLimit,
		MonthlyLimit: &monthlyLimit,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := tx.Create(&tier).Error; err != nil {
		t.Fatalf("create test tier: %v", err)
	}
	user := entity.User{
		ID:           userID,
		Email:        stringPtr("quota-test-" + uuid.NewString() + "@example.invalid"),
		TierID:       &tierID,
		TokenBalance: 500,
		TurnBalance:  2,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := tx.Create(&user).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
	usage := entity.UserUsage{
		ID:         uuid.NewString(),
		UserID:     userID,
		Date:       utcDate(now),
		UsageCount: 1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := tx.Create(&usage).Error; err != nil {
		t.Fatalf("create test usage: %v", err)
	}

	users := repository.NewUserRepo(tx)
	usageRepo := repository.NewUsageRepo(tx)
	service := NewService(users, repository.NewMembershipRepo(tx), usageRepo, nil, time.Minute)
	principal := &Principal{
		UserID:       userID,
		Tier:         &repository.MembershipTier{DailyLimit: dailyLimit, MonthlyLimit: monthlyLimit},
		TurnBalance:  2,
		TokenBalance: 500,
	}

	t.Run("monthly quota is consumed first", func(t *testing.T) {
		if err := service.EnsureChatQuota(ctx, principal); err != nil {
			t.Fatalf("monthly quota should allow chat: %v", err)
		}
		usesTurns, err := service.UsesTurnWalletForNextTurn(ctx, principal)
		if err != nil {
			t.Fatal(err)
		}
		usesTokens, err := service.UsesTokenWalletForNextTurn(ctx, principal)
		if err != nil {
			t.Fatal(err)
		}
		if usesTurns || usesTokens {
			t.Fatalf("included monthly turn selected wrong wallet: turns=%v tokens=%v", usesTurns, usesTokens)
		}
	})

	if err := tx.Model(&entity.UserUsage{}).Where("id = ?", usage.ID).Update("usage_count", 2).Error; err != nil {
		t.Fatalf("exhaust monthly quota: %v", err)
	}

	t.Run("permanent turns are used after monthly quota", func(t *testing.T) {
		if err := service.EnsureChatQuota(ctx, principal); err != nil {
			t.Fatalf("permanent turns should allow chat: %v", err)
		}
		usesTurns, err := service.UsesTurnWalletForNextTurn(ctx, principal)
		if err != nil {
			t.Fatal(err)
		}
		usesTokens, err := service.UsesTokenWalletForNextTurn(ctx, principal)
		if err != nil {
			t.Fatal(err)
		}
		if !usesTurns || usesTokens {
			t.Fatalf("permanent turn selected wrong wallet: turns=%v tokens=%v", usesTurns, usesTokens)
		}
		if err := service.DeductChatTurn(ctx, userID); err != nil {
			t.Fatalf("deduct permanent turn: %v", err)
		}
		var balance int
		if err := tx.Model(&entity.User{}).Select("turn_balance").Where("id = ?", userID).Scan(&balance).Error; err != nil {
			t.Fatal(err)
		}
		if balance != 1 {
			t.Fatalf("turn_balance = %d, want 1", balance)
		}
	})

	principal.TurnBalance = 0
	if err := tx.Model(&entity.User{}).Where("id = ?", userID).Update("turn_balance", 0).Error; err != nil {
		t.Fatal(err)
	}
	t.Run("token wallet is the final fallback", func(t *testing.T) {
		if err := service.EnsureChatQuota(ctx, principal); err != nil {
			t.Fatalf("token wallet should allow chat: %v", err)
		}
		usesTokens, err := service.UsesTokenWalletForNextTurn(ctx, principal)
		if err != nil {
			t.Fatal(err)
		}
		if !usesTokens {
			t.Fatal("expected token wallet after monthly and permanent quotas are exhausted")
		}
	})

	if err := tx.Model(&entity.User{}).Where("id = ?", userID).Update("turn_balance", 3).Error; err != nil {
		t.Fatal(err)
	}
	principal.TurnBalance = 0
	t.Run("admin grant refreshes an empty cached balance", func(t *testing.T) {
		if err := service.EnsureChatQuota(ctx, principal); err != nil {
			t.Fatalf("freshly granted turns should allow chat: %v", err)
		}
		if principal.TurnBalance != 3 {
			t.Fatalf("cached turn balance = %d, want refreshed balance 3", principal.TurnBalance)
		}
		usesTurns, err := service.UsesTurnWalletForNextTurn(ctx, principal)
		if err != nil {
			t.Fatal(err)
		}
		if !usesTurns {
			t.Fatal("expected freshly granted permanent turns to be selected")
		}
	})

	if err := tx.Model(&entity.User{}).Where("id = ?", userID).Update("turn_balance", 0).Error; err != nil {
		t.Fatal(err)
	}
	principal.TurnBalance = 0
	principal.TokenBalance = 0
	t.Run("chat is rejected after every wallet is exhausted", func(t *testing.T) {
		if err := service.EnsureChatQuota(ctx, principal); !errors.Is(err, ErrQuotaTokens) {
			t.Fatalf("EnsureChatQuota() error = %v, want ErrQuotaTokens", err)
		}
	})

	principal.TurnBalance = 5
	principal.TokenBalance = 500
	if err := tx.Model(&entity.UserUsage{}).Where("id = ?", usage.ID).Update("usage_count", dailyLimit).Error; err != nil {
		t.Fatal(err)
	}
	t.Run("daily limit remains a hard limit", func(t *testing.T) {
		if err := service.EnsureChatQuota(ctx, principal); !errors.Is(err, ErrQuotaDaily) {
			t.Fatalf("EnsureChatQuota() error = %v, want ErrQuotaDaily", err)
		}
	})
}

func utcDate(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func stringPtr(value string) *string {
	return &value
}

func postgresDatabaseURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	query := parsed.Query()
	query.Del("schema")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
