package domain

import "errors"

var (
	ErrInvalidRepoFormat = errors.New("invalid repository format, expected owner/repo")
	ErrRepoNotFound      = errors.New("repository not found on GitHub or is private")
	ErrAlreadySubscribed = errors.New("email already subscribed to this repository")
	ErrTokenNotFound     = errors.New("token not found")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrRateLimited       = errors.New("GitHub API rate limit exceeded")

	ErrEmailTaken               = errors.New("email already registered")
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrPasswordLoginUnavailable = errors.New("account uses social sign-in")
	ErrEmailNotVerified         = errors.New("email not verified")
	ErrWeakPassword             = errors.New("password does not meet policy")
	ErrTokenExpired             = errors.New("token expired")
	ErrOAuthStateMismatch       = errors.New("oauth state mismatch")
	ErrIdentityConflict         = errors.New("provider identity bound to another account")
	ErrAccessRequestPending     = errors.New("an access request is already pending for this email")
)
