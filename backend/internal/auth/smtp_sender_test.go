package auth

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSMTPMailSender_RetriesWithExponentialBackoff(t *testing.T) {
	transport := &stubTransport{errs: []error{errors.New("dial 1"), errors.New("dial 2"), nil}}
	var sleeps []time.Duration

	sender := NewSMTPMailSender(SMTPConfig{Host: "smtp.example.com", Port: "25", From: "no-reply@example.com"})
	sender.transport = transport
	sender.timeout = 10 * time.Millisecond
	sender.initialBackoff = 5 * time.Millisecond
	sender.sleep = func(d time.Duration) { sleeps = append(sleeps, d) }

	err := sender.Send("to@example.com", "subject", "body")
	if err != nil {
		t.Fatalf("expected send to succeed after retries, got: %v", err)
	}

	if transport.calls != 3 {
		t.Fatalf("expected 3 attempts (1 + 2 retries), got %d", transport.calls)
	}
	if len(sleeps) != 2 || sleeps[0] != 5*time.Millisecond || sleeps[1] != 10*time.Millisecond {
		t.Fatalf("unexpected backoff schedule: %#v", sleeps)
	}
}

func TestSMTPMailSender_TimeoutAndRetryPolicy(t *testing.T) {
	transport := &stubTransport{blockUntilCtxDone: true}
	var sleeps []time.Duration

	sender := NewSMTPMailSender(SMTPConfig{Host: "smtp.example.com", Port: "25", From: "no-reply@example.com"})
	sender.transport = transport
	sender.timeout = 5 * time.Millisecond
	sender.initialBackoff = 1 * time.Millisecond
	sender.sleep = func(d time.Duration) { sleeps = append(sleeps, d) }

	start := time.Now()
	err := sender.Send("to@example.com", "subject", "body")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if transport.calls != 3 {
		t.Fatalf("expected 3 attempts on timeout, got %d", transport.calls)
	}
	if len(sleeps) != 2 {
		t.Fatalf("expected 2 backoff sleeps, got %d", len(sleeps))
	}
	if elapsed < 10*time.Millisecond {
		t.Fatalf("expected elapsed time to include timeout attempts, got %s", elapsed)
	}
}

func TestSMTPMailSender_MockHostDoesNotBypassTransport(t *testing.T) {
	transport := &stubTransport{errs: []error{errors.New("network down"), errors.New("network down"), errors.New("network down")}}

	sender := NewSMTPMailSender(SMTPConfig{Host: "mock", Port: "25", From: "no-reply@example.com"})
	sender.transport = transport
	sender.timeout = 10 * time.Millisecond
	sender.initialBackoff = time.Millisecond
	sender.sleep = func(time.Duration) {}

	err := sender.Send("to@example.com", "subject", "body")
	if err == nil {
		t.Fatal("expected send to fail when transport fails")
	}
	if transport.calls != 3 {
		t.Fatalf("expected retry policy to execute, got %d calls", transport.calls)
	}
}

func TestSMTPTransport_RefusesUnencryptedTransport(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		_, _ = rw.WriteString("220 localhost ESMTP ready\r\n")
		_ = rw.Flush()

		for {
			line, readErr := rw.ReadString('\n')
			if readErr != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				_, _ = rw.WriteString("250-localhost\r\n250 OK\r\n")
			case strings.HasPrefix(cmd, "QUIT"):
				_, _ = rw.WriteString("221 bye\r\n")
			default:
				_, _ = rw.WriteString("250 OK\r\n")
			}
			_ = rw.Flush()
		}
	}()

	host, port, splitErr := net.SplitHostPort(ln.Addr().String())
	if splitErr != nil {
		t.Fatalf("split host port: %v", splitErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = smtpTransport{}.Send(ctx, SMTPConfig{Host: host, Port: port, From: "no-reply@example.com"}, "to@example.com", "subject", "body")
	if err == nil {
		t.Fatal("expected unencrypted SMTP transport to be rejected")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "starttls") {
		t.Fatalf("expected starttls-related error, got %v", err)
	}

	_ = ln.Close()
	<-done
}

type stubTransport struct {
	calls             int
	errs              []error
	blockUntilCtxDone bool
}

func (s *stubTransport) Send(ctx context.Context, _ SMTPConfig, _ string, _ string, _ string) error {
	s.calls++
	if s.blockUntilCtxDone {
		<-ctx.Done()
		return ctx.Err()
	}
	if len(s.errs) == 0 {
		return nil
	}
	err := s.errs[0]
	s.errs = s.errs[1:]
	return err
}
