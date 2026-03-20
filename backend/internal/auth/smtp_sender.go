package auth

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
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
	config SMTPConfig
}

func NewSMTPMailSender(config SMTPConfig) *SMTPMailSender {
	if config.Port == "" {
		config.Port = "25"
	}
	if config.From == "" {
		config.From = "no-reply@xledger.local"
	}

	return &SMTPMailSender{config: config}
}

func (s *SMTPMailSender) Send(to string, subject string, body string) error {
	if strings.EqualFold(strings.TrimSpace(s.config.Host), "mock") {
		return nil
	}

	if strings.TrimSpace(s.config.Host) == "" {
		return fmt.Errorf("smtp host is required")
	}

	addr := net.JoinHostPort(s.config.Host, s.config.Port)
	msg := "From: " + s.config.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	if err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("send smtp mail: %w", err)
	}

	return nil
}
