package internal

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	findExisting        *Subscription
	findExistingErr     error
	findByEmail         []Subscription
	findByEmailErr      error
	findByUnsubToken    *Subscription
	findByUnsubTokenErr error
	createErr           error
	createTokenErr      error
	findToken           *ConfirmationToken
	findTokenErr        error
	confirmErr          error
	deleteSubErr        error
	deleteTokenErr      error

	createdSub     *Subscription
	createdToken   *ConfirmationToken
	confirmedID    uint
	deletedSubID   uint
	deletedTokenID uint
}

func (m *mockRepo) CreateSubscription(_ context.Context, sub *Subscription) error {
	m.createdSub = sub
	if m.createErr == nil {
		sub.ID = 1
	}
	return m.createErr
}

func (m *mockRepo) FindSubscriptionByEmailAndRepo(_ context.Context, _, _ string) (*Subscription, error) {
	return m.findExisting, m.findExistingErr
}

func (m *mockRepo) FindSubscriptionsByEmail(_ context.Context, _ string) ([]Subscription, error) {
	return m.findByEmail, m.findByEmailErr
}

func (m *mockRepo) FindSubscriptionByUnsubscribeToken(_ context.Context, _ string) (*Subscription, error) {
	return m.findByUnsubToken, m.findByUnsubTokenErr
}

func (m *mockRepo) DeleteSubscription(_ context.Context, id uint) error {
	m.deletedSubID = id
	return m.deleteSubErr
}

func (m *mockRepo) ConfirmSubscription(_ context.Context, id uint) error {
	m.confirmedID = id
	return m.confirmErr
}

func (m *mockRepo) CreateToken(_ context.Context, token *ConfirmationToken) error {
	m.createdToken = token
	return m.createTokenErr
}

func (m *mockRepo) FindTokenByValue(_ context.Context, _ string) (*ConfirmationToken, error) {
	return m.findToken, m.findTokenErr
}

func (m *mockRepo) DeleteToken(_ context.Context, id uint) error {
	m.deletedTokenID = id
	return m.deleteTokenErr
}

type mockGitHub struct {
	err error
}

func (m *mockGitHub) ValidateRepo(_ context.Context, _, _ string) error {
	return m.err
}

type mockNotifier struct {
	err            error
	sentEmail      string
	sentRepo       string
	sentToken      string
	sentUnsubToken string
	callCount      int
}

func (m *mockNotifier) SendConfirmation(email, repo, token, unsubscribeToken string) error {
	m.callCount++
	m.sentEmail = email
	m.sentRepo = repo
	m.sentToken = token
	m.sentUnsubToken = unsubscribeToken
	return m.err
}

func TestService_Subscribe(t *testing.T) {
	tests := []struct {
		name        string
		req         SubscribeRequest
		repo        *mockRepo
		github      *mockGitHub
		notifier    *mockNotifier
		wantErr     error
		wantCreated bool
	}{
		{
			name:        "valid subscription",
			req:         SubscribeRequest{Email: "a@b.com", Repo: "golang/go"},
			repo:        &mockRepo{},
			github:      &mockGitHub{},
			notifier:    &mockNotifier{},
			wantErr:     nil,
			wantCreated: true,
		},
		{
			name:     "invalid repo format - no slash",
			req:      SubscribeRequest{Email: "a@b.com", Repo: "invalid"},
			repo:     &mockRepo{},
			github:   &mockGitHub{},
			notifier: &mockNotifier{},
			wantErr:  ErrInvalidRepoFormat,
		},
		{
			name:     "invalid repo format - too many slashes",
			req:      SubscribeRequest{Email: "a@b.com", Repo: "a/b/c"},
			repo:     &mockRepo{},
			github:   &mockGitHub{},
			notifier: &mockNotifier{},
			wantErr:  ErrInvalidRepoFormat,
		},
		{
			name:     "duplicate subscription",
			req:      SubscribeRequest{Email: "a@b.com", Repo: "golang/go"},
			repo:     &mockRepo{findExisting: &Subscription{ID: 1}},
			github:   &mockGitHub{},
			notifier: &mockNotifier{},
			wantErr:  ErrAlreadySubscribed,
		},
		{
			name:     "repo not on github",
			req:      SubscribeRequest{Email: "a@b.com", Repo: "ghost/ghost"},
			repo:     &mockRepo{},
			github:   &mockGitHub{err: ErrRepoNotFound},
			notifier: &mockNotifier{},
			wantErr:  ErrRepoNotFound,
		},
		{
			name:     "github rate limited",
			req:      SubscribeRequest{Email: "a@b.com", Repo: "golang/go"},
			repo:     &mockRepo{},
			github:   &mockGitHub{err: ErrRateLimited},
			notifier: &mockNotifier{},
			wantErr:  ErrRateLimited,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(tc.repo, tc.github, tc.notifier)
			err := s.Subscribe(context.Background(), tc.req)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}

			if tc.wantCreated {
				assertSubscribeSideEffects(t, tc.repo, tc.notifier, tc.req.Email)
			}
		})
	}
}

func assertSubscribeSideEffects(t *testing.T, repo *mockRepo, notifier *mockNotifier, wantEmail string) {
	t.Helper()
	if repo.createdSub == nil {
		t.Fatal("expected subscription created")
	}
	if repo.createdToken == nil {
		t.Fatal("expected confirmation token created")
	}
	if notifier.callCount != 1 {
		t.Fatalf("expected 1 confirmation email, got %d", notifier.callCount)
	}
	if notifier.sentEmail != wantEmail {
		t.Fatalf("notifier got email %q, want %q", notifier.sentEmail, wantEmail)
	}
	if repo.createdSub.UnsubscribeToken == "" {
		t.Fatal("expected unsubscribe token generated on subscription")
	}
	if notifier.sentUnsubToken != repo.createdSub.UnsubscribeToken {
		t.Fatalf("notifier got unsub token %q, want %q",
			notifier.sentUnsubToken, repo.createdSub.UnsubscribeToken)
	}
}

func TestService_Subscribe_NotifierFailureDoesNotFailRequest(t *testing.T) {
	repo := &mockRepo{}
	s := NewService(repo, &mockGitHub{}, &mockNotifier{err: errors.New("smtp down")})

	err := s.Subscribe(context.Background(), SubscribeRequest{Email: "a@b.com", Repo: "golang/go"})
	if err != nil {
		t.Fatalf("expected nil error when notifier fails, got %v", err)
	}
	if repo.createdSub == nil {
		t.Fatal("subscription should still be created")
	}
}

func TestService_ConfirmSubscription(t *testing.T) {
	tests := []struct {
		name           string
		tokenValue     string
		repo           *mockRepo
		wantErr        error
		wantConfirmed  uint
		wantDeletedTok uint
	}{
		{
			name:           "valid token",
			tokenValue:     "abc123",
			repo:           &mockRepo{findToken: &ConfirmationToken{ID: 7, SubscriptionID: 42}},
			wantErr:        nil,
			wantConfirmed:  42,
			wantDeletedTok: 7,
		},
		{
			name:       "empty token",
			tokenValue: "",
			repo:       &mockRepo{},
			wantErr:    ErrTokenNotFound,
		},
		{
			name:       "token not found",
			tokenValue: "missing",
			repo:       &mockRepo{findToken: nil},
			wantErr:    ErrTokenNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(tc.repo, &mockGitHub{}, &mockNotifier{})
			err := s.ConfirmSubscription(context.Background(), tc.tokenValue)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}
			if tc.wantConfirmed != 0 && tc.repo.confirmedID != tc.wantConfirmed {
				t.Fatalf("confirmed id=%d, want %d", tc.repo.confirmedID, tc.wantConfirmed)
			}
			if tc.wantDeletedTok != 0 && tc.repo.deletedTokenID != tc.wantDeletedTok {
				t.Fatalf("deleted token id=%d, want %d", tc.repo.deletedTokenID, tc.wantDeletedTok)
			}
		})
	}
}

func TestService_Unsubscribe(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		repo        *mockRepo
		wantErr     error
		wantDeleted uint
	}{
		{
			name:        "valid token",
			token:       "abc",
			repo:        &mockRepo{findByUnsubToken: &Subscription{ID: 42}},
			wantErr:     nil,
			wantDeleted: 42,
		},
		{
			name:    "empty token",
			token:   "",
			repo:    &mockRepo{},
			wantErr: ErrTokenNotFound,
		},
		{
			name:    "token not found",
			token:   "missing",
			repo:    &mockRepo{findByUnsubToken: nil},
			wantErr: ErrTokenNotFound,
		},
		{
			name:    "lookup error bubbles up",
			token:   "x",
			repo:    &mockRepo{findByUnsubTokenErr: errors.New("db down")},
			wantErr: nil, // checked separately
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(tc.repo, &mockGitHub{}, &mockNotifier{})
			err := s.Unsubscribe(context.Background(), tc.token)

			if tc.name == "lookup error bubbles up" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}
			if tc.wantDeleted != 0 && tc.repo.deletedSubID != tc.wantDeleted {
				t.Fatalf("deleted sub id=%d, want %d", tc.repo.deletedSubID, tc.wantDeleted)
			}
		})
	}
}

func TestService_GetSubscriptions(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		repo    *mockRepo
		wantErr error
		wantLen int
	}{
		{
			name:  "returns subscriptions for valid email",
			email: "a@b.com",
			repo: &mockRepo{findByEmail: []Subscription{
				{Email: "a@b.com", Repo: "golang/go", Confirmed: true, LastSeenTag: "v1"},
				{Email: "a@b.com", Repo: "gin-gonic/gin"},
			}},
			wantLen: 2,
		},
		{
			name:    "empty email",
			email:   "",
			repo:    &mockRepo{},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "whitespace only",
			email:   "   ",
			repo:    &mockRepo{},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "malformed email",
			email:   "not-an-email",
			repo:    &mockRepo{},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "valid email, no results",
			email:   "nobody@b.com",
			repo:    &mockRepo{findByEmail: nil},
			wantLen: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(tc.repo, &mockGitHub{}, &mockNotifier{})
			got, err := s.GetSubscriptions(context.Background(), tc.email)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}
			if err == nil && len(got) != tc.wantLen {
				t.Fatalf("got %d subs, want %d", len(got), tc.wantLen)
			}
		})
	}
}

func TestService_GetSubscriptions_DTOMapping(t *testing.T) {
	repo := &mockRepo{findByEmail: []Subscription{
		{Email: "a@b.com", Repo: "golang/go", Confirmed: true, LastSeenTag: "v1.22.0",
			UnsubscribeToken: "secret-should-not-leak"},
	}}
	s := NewService(repo, &mockGitHub{}, &mockNotifier{})

	got, err := s.GetSubscriptions(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 sub, got %d", len(got))
	}
	if got[0].Email != "a@b.com" || got[0].Repo != "golang/go" ||
		!got[0].Confirmed || got[0].LastSeenTag != "v1.22.0" {
		t.Fatalf("unexpected DTO: %+v", got[0])
	}
}

func TestRepoFormatRegex(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"golang/go", true},
		{"a/b", true},
		{"owner-name/repo.name_1", true},
		{"invalid", false},
		{"a/b/c", false},
		{"", false},
		{"/repo", false},
		{"owner/", false},
		{"own er/repo", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := repoFormatRegex.MatchString(tc.input); got != tc.valid {
				t.Fatalf("repoFormatRegex(%q) = %v, want %v", tc.input, got, tc.valid)
			}
		})
	}
}
