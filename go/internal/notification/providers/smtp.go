package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/koblas/besops/internal/notification"
)

type SMTPNotifier struct{}

func (n *SMTPNotifier) Name() string { return "smtp" }

func (n *SMTPNotifier) Send(_ context.Context, config map[string]any, msg string, monitor *notification.MonitorInfo, _ *notification.HeartbeatInfo) error {
	host, _ := config["smtpHost"].(string)
	if host == "" {
		return fmt.Errorf("smtpHost is required")
	}

	portFloat, _ := config["smtpPort"].(float64)
	port := int(portFloat)
	if port == 0 {
		port = 587
	}

	from, _ := config["smtpFrom"].(string)
	if from == "" {
		return fmt.Errorf("smtpFrom is required")
	}

	to, _ := config["smtpTo"].(string)
	if to == "" {
		return fmt.Errorf("smtpTo is required")
	}

	username, _ := config["smtpUsername"].(string)
	password, _ := config["smtpPassword"].(string)
	secure, _ := config["smtpSecurity"].(string)
	ignoreTLS, _ := config["smtpIgnoreTLSError"].(bool)

	subject := "Bes Ops Alert"
	if monitor != nil {
		subject = fmt.Sprintf("[Bes Ops] %s", monitor.Name)
	}

	recipients := strings.Split(to, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		from, to, subject)
	body := headers + msg

	addr := fmt.Sprintf("%s:%d", host, port)

	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	if secure == "TLS" || port == 465 {
		return sendViaTLS(addr, host, auth, from, recipients, body, ignoreTLS)
	}

	return sendViaStartTLS(addr, host, auth, from, recipients, body, secure == "STARTTLS", ignoreTLS)
}

func sendViaTLS(addr, host string, auth smtp.Auth, from string, to []string, body string, ignoreTLS bool) error {
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: ignoreTLS, //nolint:gosec // user-configurable setting
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client: %w", err)
	}
	defer client.Close()

	return sendMail(client, auth, from, to, body)
}

func sendViaStartTLS(addr, host string, auth smtp.Auth, from string, to []string, body string, requireTLS, ignoreTLS bool) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("TCP dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok || requireTLS {
		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: ignoreTLS, //nolint:gosec // user-configurable setting
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS: %w", err)
		}
	}

	return sendMail(client, auth, from, to, body)
}

func sendMail(client *smtp.Client, auth smtp.Auth, from string, to []string, body string) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL: %w", err)
	}

	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("SMTP RCPT %s: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}

	if _, err := w.Write([]byte(body)); err != nil {
		return fmt.Errorf("writing email body: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("closing email body: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("SMTP QUIT: %w", err)
	}
	return nil
}
