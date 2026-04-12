package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type GitHubClient struct {
	httpClient *http.Client
	token      string
}

func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		token:      token,
	}
}

func (g *GitHubClient) ValidateRepo(ctx context.Context, owner, repo string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	g.setHeaders(req)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach GitHub API: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrRepoNotFound
	case http.StatusForbidden, http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return fmt.Errorf("unexpected GitHub API status: %d", resp.StatusCode)
	}
}

func (g *GitHubClient) GetLatestRelease(ctx context.Context, owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	g.setHeaders(req)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach GitHub API: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := decodeJSON(resp.Body, &release); err != nil {
			return "", fmt.Errorf("failed to decode release response: %w", err)
		}
		return release.TagName, nil
	case http.StatusNotFound:
		return "", nil
	case http.StatusForbidden, http.StatusTooManyRequests:
		return "", ErrRateLimited
	default:
		return "", fmt.Errorf("unexpected GitHub API status: %d", resp.StatusCode)
	}
}

func (g *GitHubClient) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "repo-release-notifier")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}
}
