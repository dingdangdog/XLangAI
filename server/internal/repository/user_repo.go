package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"xlangai/server/internal/entity"
	"xlangai/server/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateUserParams struct {
	Phone       string
	Email       string
	Password    string
	Nickname    string
	TurnBalance int // 注册赠送的永久对话次数
}

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) activeUserQuery() *gorm.DB {
	return r.db.Model(&entity.User{}).Where("status = ? AND deleted_at IS NULL", "active")
}

func (r *UserRepo) freeTierID(ctx context.Context) *string {
	var tier entity.MembershipTier
	if err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", "free", "active").
		First(&tier).Error; err != nil {
		return nil
	}
	return &tier.ID
}

// CreateSmsUser 短信验证码注册：仅绑定手机号，无密码（password_hash 为 NULL）。
func (r *UserRepo) CreateSmsUser(ctx context.Context, phone, nickname string, turnBalance int) (*model.User, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, errors.New("phone required")
	}
	if turnBalance < 0 {
		turnBalance = 0
	}
	var nickPtr *string
	n := strings.TrimSpace(nickname)
	if n != "" {
		if len([]rune(n)) > 32 {
			n = string([]rune(n)[:32])
		}
		nickPtr = &n
	}
	row := entity.User{
		ID:          uuid.New().String(),
		Phone:       &phone,
		Nickname:    nickPtr,
		TierID:      r.freeTierID(ctx),
		TurnBalance: turnBalance,
		Status:      "active",
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return userToModel(&row), nil
}

func (r *UserRepo) Create(ctx context.Context, p CreateUserParams) (*model.User, error) {
	if p.Phone == "" && p.Email == "" {
		return nil, errors.New("phone or email required")
	}
	if p.TurnBalance < 0 {
		p.TurnBalance = 0
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hash)
	row := entity.User{
		ID:           uuid.New().String(),
		Phone:        strPtr(p.Phone),
		Email:        strPtr(p.Email),
		PasswordHash: &hashStr,
		Nickname:     strPtr(p.Nickname),
		TierID:       r.freeTierID(ctx),
		TurnBalance:  p.TurnBalance,
		Status:       "active",
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return userToModel(&row), nil
}

func (r *UserRepo) CheckPassword(hash, password string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashStr := string(hash)
	res := r.activeUserQuery().WithContext(ctx).Where("id = ?", id).
		Updates(map[string]any{"password_hash": hashStr})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ResolveActiveDefaultLlmConfigID 返回用户指定的、且仍为 active 的默认 LLM 配置 ID；未设置或已失效时返回 nil。
func (r *UserRepo) ResolveActiveDefaultLlmConfigID(ctx context.Context, userID string) (*string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil
	}
	var row entity.User
	err := r.db.WithContext(ctx).
		Select("default_llm_config_id").
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if row.DefaultLlmConfigID == nil {
		return nil, nil
	}
	pref := strings.TrimSpace(*row.DefaultLlmConfigID)
	if pref == "" {
		return nil, nil
	}
	var llm entity.SysLlmServiceConfig
	err = r.db.WithContext(ctx).
		Where("id = ? AND status = ?", pref, "active").
		First(&llm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	id := llm.ID
	return &id, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var row entity.User
	err := r.activeUserQuery().WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return userToModel(&row), nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	var row entity.User
	err := r.activeUserQuery().WithContext(ctx).Where("phone = ?", phone).First(&row).Error
	if err != nil {
		return nil, err
	}
	return userToModel(&row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var row entity.User
	err := r.activeUserQuery().WithContext(ctx).Where("email = ?", email).First(&row).Error
	if err != nil {
		return nil, err
	}
	return userToModel(&row), nil
}

func (r *UserRepo) GetByPhoneOrEmail(ctx context.Context, phone, email string) (*model.User, error) {
	if phone != "" {
		return r.GetByPhone(ctx, phone)
	}
	if email != "" {
		return r.GetByEmail(ctx, email)
	}
	return nil, errors.New("phone or email required")
}

// UpdateProfile 按非 nil 指针更新字段；nickname / avatarURL / languageID（母语）为 "" 时写入 NULL。
func (r *UserRepo) UpdateProfile(ctx context.Context, id string, nickname, avatarURL, languageID *string) error {
	if nickname == nil && avatarURL == nil && languageID == nil {
		return nil
	}
	updates := map[string]any{}
	if nickname != nil {
		v := strings.TrimSpace(*nickname)
		if v == "" {
			updates["nickname"] = nil
		} else {
			updates["nickname"] = v
		}
	}
	if avatarURL != nil {
		v := strings.TrimSpace(*avatarURL)
		if v == "" {
			updates["avatar_url"] = nil
		} else {
			updates["avatar_url"] = v
		}
	}
	if languageID != nil {
		v := strings.TrimSpace(*languageID)
		if v == "" {
			updates["language_id"] = nil
		} else {
			updates["language_id"] = v
		}
	}
	res := r.activeUserQuery().WithContext(ctx).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ApplySubscriptionEntitlement 写入会员档与订阅到期时间（UTC）。
func (r *UserRepo) ApplySubscriptionEntitlement(ctx context.Context, userID string, tierID *string, subExpires *time.Time) error {
	return r.activeUserQuery().WithContext(ctx).Where("id = ?", userID).
		Updates(map[string]any{
			"tier_id":                 tierID,
			"subscription_expires_at": subExpires,
		}).Error
}

// AddTokenBalance 增加 token 余额（消耗型内购）。
func (r *UserRepo) AddTokenBalance(ctx context.Context, userID string, delta int64) error {
	if delta == 0 {
		return nil
	}
	return r.activeUserQuery().WithContext(ctx).Where("id = ?", userID).
		Update("token_balance", gorm.Expr("COALESCE(token_balance, 0) + ?", delta)).Error
}

// DeductTokenBalance 扣减 token；余额不足时扣到 0。返回扣减后的余额。
func (r *UserRepo) DeductTokenBalance(ctx context.Context, userID string, n int64) (int64, error) {
	q := r.activeUserQuery().WithContext(ctx).Where("id = ?", userID)
	if n > 0 {
		if err := q.Update("token_balance", gorm.Expr("GREATEST(0, COALESCE(token_balance, 0) - ?)", n)).Error; err != nil {
			return 0, err
		}
	}
	var row entity.User
	if err := r.db.WithContext(ctx).Select("token_balance").Where("id = ?", userID).First(&row).Error; err != nil {
		return 0, err
	}
	return row.TokenBalance, nil
}

// AddTurnBalance 增加永久对话次数（运营加次）。
func (r *UserRepo) AddTurnBalance(ctx context.Context, userID string, delta int) error {
	if delta == 0 {
		return nil
	}
	return r.activeUserQuery().WithContext(ctx).Where("id = ?", userID).
		Update("turn_balance", gorm.Expr("GREATEST(0, COALESCE(turn_balance, 0) + ?)", delta)).Error
}

// TurnBalance 返回当前永久对话次数余额，用于管理员发放后刷新缓存中的空余额。
func (r *UserRepo) TurnBalance(ctx context.Context, userID string) (int, error) {
	var row entity.User
	if err := r.activeUserQuery().WithContext(ctx).
		Select("turn_balance").
		Where("id = ?", userID).
		First(&row).Error; err != nil {
		return 0, err
	}
	return row.TurnBalance, nil
}

// DeductTurnBalance 原子扣减永久次数；余额不足时不扣减，返回 ok=false。
func (r *UserRepo) DeductTurnBalance(ctx context.Context, userID string, n int) (ok bool, err error) {
	if n <= 0 {
		return true, nil
	}
	res := r.activeUserQuery().WithContext(ctx).
		Where("id = ? AND COALESCE(turn_balance, 0) >= ?", userID, n).
		Update("turn_balance", gorm.Expr("turn_balance - ?", n))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
