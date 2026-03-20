package auth

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPSender interface {
	Send(to string, subject string, body string) error
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type SMTPMailSender struct {
	config         SMTPConfig
	transport      SMTPTransport
	timeout        time.Duration
	maxRetries     int
	initialBackoff time.Duration
	sleep          func(time.Duration)
}

type SMTPTransport interface {
	Send(ctx context.Context, config SMTPConfig, to string, subject string, body string) error
}

type smtpTransport struct{}

func NewSMTPMailSender(config SMTPConfig) *SMTPMailSender {
	if config.Port == "" {
		config.Port = "25"
	}
	if config.From == "" {
		config.From = "no-reply@xledger.local"
	}

	return &SMTPMailSender{
		config:         config,
		transport:      smtpTransport{},
		timeout:        5 * time.Second,
		maxRetries:     2,
		initialBackoff: 100 * time.Millisecond,
		sleep:          time.Sleep,
	}
}

func (s *SMTPMailSender) Send(to string, subject string, body string) error {
	if strings.EqualFold(strings.TrimSpace(s.config.Host), "mock") {
		return nil
	}

	if strings.TrimSpace(s.config.Host) == "" {
		return fmt.Errorf("smtp host is required")
	}

	attempts := s.maxRetries + 1
	backoff := s.initialBackoff
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		err := s.transport.Send(ctx, s.config, to, subject, body)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt == attempts-1 {
			break
		}
		s.sleep(backoff)
		backoff *= 2
	}

	return fmt.Errorf("send smtp mail: %w", lastErr)
}

func (smtpTransport) Send(ctx context.Context, config SMTPConfig, to string, subject string, body string) error {
	addr := net.JoinHostPort(config.Host, config.Port)
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	if config.Username != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(config.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	message := "From: " + config.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	if _, err := io.WriteString(w, message); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
