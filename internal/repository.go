package internal

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *Repository) FindSubscriptionByEmailAndRepo(ctx context.Context, email, repo string) (*Subscription, error) {
	var sub Subscription
	err := r.db.WithContext(ctx).
		Where("email = ? AND repo = ?", email, repo).
		First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sub, err
}

func (r *Repository) FindSubscriptionsByEmail(ctx context.Context, email string) ([]Subscription, error) {
	var subs []Subscription
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		Find(&subs).Error
	return subs, err
}

func (r *Repository) FindSubscriptionByUnsubscribeToken(ctx context.Context, token string) (*Subscription, error) {
	var sub Subscription
	err := r.db.WithContext(ctx).
		Where("unsubscribe_token = ?", token).
		First(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sub, err
}

func (r *Repository) ConfirmSubscription(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&Subscription{}).
		Where("id = ?", id).
		Update("confirmed", true).Error
}

func (r *Repository) DeleteSubscription(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Subscription{}, id).Error
}

// --- Token queries ---

func (r *Repository) CreateToken(ctx context.Context, token *ConfirmationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *Repository) FindTokenByValue(ctx context.Context, tokenValue string) (*ConfirmationToken, error) {
	var token ConfirmationToken
	err := r.db.WithContext(ctx).
		Preload("Subscription").
		Where("token = ?", tokenValue).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}

func (r *Repository) DeleteToken(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&ConfirmationToken{}, id).Error
}

func (r *Repository) DeleteTokensBySubscriptionID(ctx context.Context, subscriptionID uint) error {
	return r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Delete(&ConfirmationToken{}).Error
}

// --- Scanner queries ---

func (r *Repository) FindDistinctConfirmedRepos(ctx context.Context) ([]string, error) {
	var repos []string
	err := r.db.WithContext(ctx).
		Model(&Subscription{}).
		Where("confirmed = ?", true).
		Distinct("repo").
		Pluck("repo", &repos).Error
	return repos, err
}

func (r *Repository) FindConfirmedSubscriptionsByRepo(ctx context.Context, repo string) ([]Subscription, error) {
	var subs []Subscription
	err := r.db.WithContext(ctx).
		Where("repo = ? AND confirmed = ?", repo, true).
		Find(&subs).Error
	return subs, err
}

func (r *Repository) UpdateLastSeenTag(ctx context.Context, id uint, tag string) error {
	return r.db.WithContext(ctx).
		Model(&Subscription{}).
		Where("id = ?", id).
		Update("last_seen_tag", tag).Error
}
