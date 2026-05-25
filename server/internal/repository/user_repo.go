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
	Phone    string
	Email    string
	Password string
	Nickname string
}

// CreateOAuthUserParams 使用 Google / Apple 的 sub 创建无密码用户（password_hash 为 NULL）。
type CreateOAuthUserParams struct {
	GoogleSub *string
	AppleSub  *string
	Email     *string
	Nickname  *string
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
func (r *UserRepo) CreateSmsUser(ctx context.Context, phone, nickname string) (*model.User, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, errors.New("phone required")
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
		ID:       uuid.New().String(),
		Phone:    &phone,
		Nickname: nickPtr,
		TierID:   r.freeTierID(ctx),
		Status:   "active",
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

func (r *UserRepo) GetByGoogleSub(ctx context.Context, sub string) (*model.User, error) {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return nil, errors.New("empty google_sub")
	}
	var row entity.User
	err := r.activeUserQuery().WithContext(ctx).Where("google_sub = ?", sub).First(&row).Error
	if err != nil {
		return nil, err
	}
	return userToModel(&row), nil
}

func (r *UserRepo) GetByAppleSub(ctx context.Context, sub string) (*model.User, error) {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return nil, errors.New("empty apple_sub")
	}
	var row entity.User
	err := r.activeUserQuery().WithContext(ctx).Where("apple_sub = ?", sub).First(&row).Error
	if err != nil {
		return nil, err
	}
	return userToModel(&row), nil
}

func (r *UserRepo) CreateOAuthUser(ctx context.Context, p CreateOAuthUserParams) (*model.User, error) {
	hasGoogle := p.GoogleSub != nil && strings.TrimSpace(*p.GoogleSub) != ""
	hasApple := p.AppleSub != nil && strings.TrimSpace(*p.AppleSub) != ""
	if !hasGoogle && !hasApple {
		return nil, errors.New("google_sub or apple_sub required")
	}
	var gPtr, aPtr *string
	if hasGoogle {
		v := strings.TrimSpace(*p.GoogleSub)
		gPtr = &v
	}
	if hasApple {
		v := strings.TrimSpace(*p.AppleSub)
		aPtr = &v
	}
	var emPtr *string
	if p.Email != nil {
		e := strings.TrimSpace(strings.ToLower(*p.Email))
		if e != "" {
			emPtr = &e
		}
	}
	var nickPtr *string
	if p.Nickname != nil {
		n := strings.TrimSpace(*p.Nickname)
		if n != "" {
			if len([]rune(n)) > 32 {
				n = string([]rune(n)[:32])
			}
			nickPtr = &n
		}
	}
	if nickPtr == nil {
		base := "用户"
		if emPtr != nil {
			local := *emPtr
			if i := strings.IndexByte(local, '@'); i > 0 {
				local = local[:i]
			}
			if local != "" {
				base = local
			}
		}
		if len([]rune(base)) > 32 {
			base = string([]rune(base)[:32])
		}
		nickPtr = &base
	}
	row := entity.User{
		ID:        uuid.New().String(),
		Email:     emPtr,
		Nickname:  nickPtr,
		TierID:    r.freeTierID(ctx),
		GoogleSub: gPtr,
		AppleSub:  aPtr,
		Status:    "active",
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return userToModel(&row), nil
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
