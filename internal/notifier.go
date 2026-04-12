package internal

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Zero-width space inserted after the URL scheme to prevent mail clients
// from auto-linkifying the "copy this link" text. See README.
const zwsp = "\u200B"

func breakAutoLink(url string) string {
	return strings.Replace(url, "://", zwsp+"://", 1)
}

type Notifier struct {
	host     string
	port     string
	username string
	password string
	baseURL  string
}

func NewNotifier(cfg *Config) *Notifier {
	return &Notifier{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		username: cfg.SMTPUserName,
		password: cfg.SMTPPassword,
		baseURL:  cfg.BaseURL,
	}
}

func (n *Notifier) SendConfirmation(email, repo, token string) error {
	subject := fmt.Sprintf("Confirm your subscription to %s releases", repo)
	confirmURL := fmt.Sprintf("%s/api/confirm/%s", n.baseURL, token)

	plain := fmt.Sprintf(
		"You have subscribed to release notifications for %s.\n\n"+
			"Please confirm your subscription by clicking the link below:\n%s\n\n"+
			"If you did not request this, you can safely ignore this email.",
		repo, confirmURL,
	)
	html, err := renderTemplate("confirmation.html", map[string]string{
		"Repo":              repo,
		"ConfirmURL":        confirmURL,
		"ConfirmURLDisplay": breakAutoLink(confirmURL),
	})
	if err != nil {
		return err
	}

	return n.send(email, subject, plain, html)
}

func (n *Notifier) SendReleaseNotification(email, repo, tag, unsubscribeToken string) error {
	subject := fmt.Sprintf("New release %s for %s", tag, repo)
	releaseURL := fmt.Sprintf("https://github.com/%s/releases/tag/%s", repo, tag)
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", n.baseURL, unsubscribeToken)

	plain := fmt.Sprintf(
		"A new release has been published for %s!\n\n"+
			"Tag: %s\n"+
			"View release: %s\n\n"+
			"To unsubscribe from these notifications:\n%s",
		repo, tag, releaseURL, unsubscribeURL,
	)
	html, err := renderTemplate("release.html", map[string]string{
		"Repo":           repo,
		"Tag":            tag,
		"ReleaseURL":     releaseURL,
		"UnsubscribeURL": unsubscribeURL,
	})
	if err != nil {
		return err
	}

	return n.send(email, subject, plain, html)
}

func (n *Notifier) send(to, subject, plain, html string) error {
	if n.host == "" {
		return fmt.Errorf("SMTP not configured: SMTP_HOST is required")
	}

	boundary := "boundary-repo-release-notifier"
	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/alternative; boundary=\"%s\"\r\n"+
			"\r\n"+
			"--%s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"Content-Transfer-Encoding: 8bit\r\n"+
			"\r\n%s\r\n"+
			"--%s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"Content-Transfer-Encoding: 8bit\r\n"+
			"\r\n%s\r\n"+
			"--%s--\r\n",
		n.username, to, subject, boundary,
		boundary, plain,
		boundary, html,
		boundary,
	)

	addr := fmt.Sprintf("%s:%s", n.host, n.port)
	auth := smtp.PlainAuth("", n.username, n.password, n.host)

	if err := smtp.SendMail(addr, auth, n.username, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}
