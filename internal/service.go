package internal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"strings"
)

var repoFormatRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub *Subscription) error
	FindSubscriptionByEmailAndRepo(ctx context.Context, email, repo string) (*Subscription, error)
	ConfirmSubscription(ctx context.Context, id uint) error
	CreateToken(ctx context.Context, token *ConfirmationToken) error
	FindTokenByValue(ctx context.Context, tokenValue string) (*ConfirmationToken, error)
	DeleteToken(ctx context.Context, id uint) error
}

type RepoValidator interface {
	ValidateRepo(ctx context.Context, owner, repo string) error
}

type ConfirmationSender interface {
	SendConfirmation(email, repo, token string) error
}

type Service struct {
	repo     SubscriptionRepository
	github   RepoValidator
	notifier ConfirmationSender
}

func NewService(repo SubscriptionRepository, github RepoValidator, notifier ConfirmationSender) *Service {
	return &Service{repo: repo, github: github, notifier: notifier}
}

func (s *Service) Subscribe(ctx context.Context, req SubscribeRequest) error {
	if !repoFormatRegex.MatchString(req.Repo) {
		return ErrInvalidRepoFormat
	}

	existing, err := s.repo.FindSubscriptionByEmailAndRepo(ctx, req.Email, req.Repo)
	if err != nil {
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}
	if existing != nil {
		return ErrAlreadySubscribed
	}

	parts := strings.SplitN(req.Repo, "/", 2)
	if err := s.github.ValidateRepo(ctx, parts[0], parts[1]); err != nil {
		return err
	}

	sub := &Subscription{
		Email: req.Email,
		Repo:  req.Repo,
	}
	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	tokenValue, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate confirmation token: %w", err)
	}

	token := &ConfirmationToken{
		Token:          tokenValue,
		SubscriptionID: sub.ID,
	}
	if err := s.repo.CreateToken(ctx, token); err != nil {
		return fmt.Errorf("failed to save confirmation token: %w", err)
	}

	if err := s.notifier.SendConfirmation(req.Email, req.Repo, tokenValue); err != nil {
		log.Printf("failed to send confirmation email for repo=%s: %v", req.Repo, err)
	}

	return nil
}

func (s *Service) ConfirmSubscription(ctx context.Context, tokenValue string) error {
	if tokenValue == "" {
		return ErrTokenNotFound
	}

	token, err := s.repo.FindTokenByValue(ctx, tokenValue)
	if err != nil {
		return fmt.Errorf("failed to look up token: %w", err)
	}
	if token == nil {
		return ErrTokenNotFound
	}

	if err := s.repo.ConfirmSubscription(ctx, token.SubscriptionID); err != nil {
		return fmt.Errorf("failed to confirm subscription id=%d: %w", token.SubscriptionID, err)
	}

	if err := s.repo.DeleteToken(ctx, token.ID); err != nil {
		log.Printf("failed to delete used confirmation token id=%d: %v", token.ID, err)
	}

	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand failed: %w", err)
	}
	return hex.EncodeToString(b), nil
}
