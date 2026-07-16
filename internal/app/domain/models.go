package domain

import "time"

type Subscription struct {
	ID               uint      `gorm:"primaryKey" json:"-"`
	Email            string    `gorm:"type:varchar(255);not null" json:"email"`
	Repo             string    `gorm:"type:varchar(255);not null" json:"repo"`
	Confirmed        bool      `gorm:"default:false;not null" json:"confirmed"`
	UnsubscribeToken string    `gorm:"type:varchar(64);uniqueIndex" json:"-"`
	UserID           *uint     `gorm:"index" json:"-"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

// WatchedRepo is the scanner's per-repo release cursor: the last release tag it
// has already notified subscribers about. One row per repo, regardless of how
// many subscriptions point at it.
type WatchedRepo struct {
	Repo         string    `gorm:"primaryKey;type:varchar(255)" json:"repo"`
	LastSeenTag  string    `gorm:"type:varchar(255);default:''" json:"last_seen_tag"`
	LastPolledAt time.Time `gorm:"not null;default:now()" json:"-"`
}

// IsNewRelease reports whether tag is a release this repo has not been notified
// about yet. Empty tags are ignored so an upstream "no releases yet" response is
// not treated as a regression.
func (w *WatchedRepo) IsNewRelease(tag string) bool {
	return tag != "" && tag != w.LastSeenTag
}

type ConfirmationToken struct {
	ID             uint         `gorm:"primaryKey" json:"-"`
	Token          string       `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	SubscriptionID uint         `gorm:"not null;index" json:"-"`
	Subscription   Subscription `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt      time.Time    `json:"-"`
}

type User struct {
	ID            uint      `gorm:"primaryKey" json:"-"`
	Email         string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	EmailVerified bool      `gorm:"not null;default:false" json:"email_verified"`
	PasswordHash  *string   `gorm:"type:varchar(255)" json:"-"`
	GoogleSub     *string   `gorm:"type:varchar(255)" json:"-"`
	GitHubID      *int64    `gorm:"column:github_id" json:"-"`
	GitHubLogin   *string   `gorm:"type:varchar(255)" json:"-"`
	TokenVersion  int       `gorm:"not null;default:0" json:"-"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

// CanPasswordLogin reports whether this account was set up with a password.
// OAuth-only accounts have a nil PasswordHash and must sign in via their provider.
func (u *User) CanPasswordLogin() bool { return u.PasswordHash != nil && *u.PasswordHash != "" }

type EmailVerificationToken struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	UserID    uint      `gorm:"not null;index" json:"-"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"-"`
}

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	UserID    uint      `gorm:"not null;index" json:"-"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"-"`
}

type APIKey struct {
	ID              uint       `gorm:"primaryKey" json:"-"`
	KeyHash         string     `gorm:"type:char(64);uniqueIndex;not null" json:"-"`
	KeyPrefix       string     `gorm:"type:varchar(12);not null" json:"key_prefix"`
	HolderEmail     string     `gorm:"type:varchar(255);not null" json:"holder_email"`
	Label           string     `gorm:"type:varchar(255)" json:"label,omitempty"`
	AccessRequestID *uint      `gorm:"column:access_request_id" json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
}

// Active reports whether the key can still authenticate (i.e. has not been revoked).
func (k *APIKey) Active() bool { return k.RevokedAt == nil }

type AccessRequest struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Email      string     `gorm:"type:varchar(255);not null" json:"email"`
	Reason     string     `gorm:"type:text" json:"reason"`
	Status     string     `gorm:"type:varchar(16);not null;default:pending" json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}
