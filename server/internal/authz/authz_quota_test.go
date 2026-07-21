package authz

import (
	"testing"

	"xlangai/server/internal/repository"
)

func TestWalletForMonthUsagePriority(t *testing.T) {
	p := &Principal{
		Tier:         &repository.MembershipTier{MonthlyLimit: 10},
		TurnBalance:  3,
		TokenBalance: 100,
	}

	tests := []struct {
		name       string
		monthUsage int
		want       chatQuotaWallet
	}{
		{name: "monthly quota first", monthUsage: 9, want: chatQuotaIncluded},
		{name: "permanent turns after monthly quota", monthUsage: 10, want: chatQuotaTurnBalance},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.walletForMonthUsage(tt.monthUsage); got != tt.want {
				t.Fatalf("walletForMonthUsage(%d) = %v, want %v", tt.monthUsage, got, tt.want)
			}
		})
	}
}

func TestWalletForMonthUsageFallbacks(t *testing.T) {
	p := &Principal{
		Tier:         &repository.MembershipTier{MonthlyLimit: 10},
		TokenBalance: 100,
	}
	if got := p.walletForMonthUsage(10); got != chatQuotaTokenBalance {
		t.Fatalf("walletForMonthUsage() = %v, want token balance", got)
	}

	p.TokenBalance = 0
	if got := p.walletForMonthUsage(10); got != chatQuotaUnavailable {
		t.Fatalf("walletForMonthUsage() = %v, want unavailable", got)
	}
}

func TestWalletForMonthUsageFreeTierUsesPermanentTurns(t *testing.T) {
	p := &Principal{
		Tier:        &repository.MembershipTier{},
		TurnBalance: 2,
	}
	if got := p.walletForMonthUsage(0); got != chatQuotaTurnBalance {
		t.Fatalf("walletForMonthUsage() = %v, want permanent turns", got)
	}
}
