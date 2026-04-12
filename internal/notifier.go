package internal

import (
	"fmt"
	"net/smtp"
)

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

	body := fmt.Sprintf(
		"You have subscribed to release notifications for %s.\n\n"+
			"Please confirm your subscription by clicking the link below:\n%s\n\n"+
			"If you did not request this, you can safely ignore this email.",
		repo, confirmURL,
	)

	return n.send(email, subject, body)
}

func (n *Notifier) SendReleaseNotification(email, repo, tag, unsubscribeToken string) error {
	subject := fmt.Sprintf("New release %s for %s", tag, repo)
	releaseURL := fmt.Sprintf("https://github.com/%s/releases/tag/%s", repo, tag)
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", n.baseURL, unsubscribeToken)

	body := fmt.Sprintf(
		"A new release has been published for %s!\n\n"+
			"Tag: %s\n"+
			"View release: %s\n\n"+
			"To unsubscribe from these notifications:\n%s",
		repo, tag, releaseURL, unsubscribeURL,
	)

	return n.send(email, subject, body)
}

func (n *Notifier) send(to, subject, body string) error {
	if n.host == "" {
		return fmt.Errorf("SMTP not configured: SMTP_HOST is required")
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		n.username, to, subject, body,
	)

	addr := fmt.Sprintf("%s:%s", n.host, n.port)
	auth := smtp.PlainAuth("", n.username, n.password, n.host)

	if err := smtp.SendMail(addr, auth, n.username, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}
