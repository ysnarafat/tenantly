package services

import (
	"crypto/tls"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type InvitationEmailSender interface {
	SendInvitation(email, organizationName, token string) error
}

type SMTPInvitationEmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	appURL   string
}

func NewSMTPInvitationEmailSender(host string, port int, username, password, from, appURL string) *SMTPInvitationEmailSender {
	return &SMTPInvitationEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		appURL:   strings.TrimRight(appURL, "/"),
	}
}

func (s *SMTPInvitationEmailSender) SendInvitation(email, organizationName, token string) error {
	if s.host == "" || s.port < 1 || s.username == "" || s.password == "" || s.from == "" || s.appURL == "" {
		return fmt.Errorf("invitation email configuration is incomplete")
	}
	if token == "" {
		return fmt.Errorf("invitation token is required")
	}
	from, err := mail.ParseAddress(s.from)
	if err != nil {
		return fmt.Errorf("invalid invitation sender address")
	}
	to, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid invitation recipient address")
	}
	link, err := s.invitationURL(token)
	if err != nil {
		return err
	}

	body := strings.Join([]string{
		fmt.Sprintf("You have been invited to administer %s on Tenantly.", organizationName),
		"",
		"Create your account and choose a password using this one-time link:",
		link,
		"",
		"This invitation expires in 7 days. If you did not expect this message, you can ignore it.",
		"",
	}, "\r\n")
	var encodedBody strings.Builder
	encoder := quotedprintable.NewWriter(&encodedBody)
	if _, err := io.WriteString(encoder, body); err != nil {
		return fmt.Errorf("encode invitation email: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("finish encoding invitation email: %w", err)
	}

	message := strings.Join([]string{
		"From: " + from.String(),
		"To: " + to.String(),
		"Subject: Tenantly organization invitation",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: quoted-printable",
		"",
	}, "\r\n") + encodedBody.String()

	address := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	connection, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer func() { _ = connection.Close() }()
	_ = connection.SetDeadline(time.Now().Add(20 * time.Second))

	client, err := smtp.NewClient(connection, s.host)
	if err != nil {
		return fmt.Errorf("initialize SMTP client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("SMTP server does not support STARTTLS")
	}
	if err := client.StartTLS(&tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}); err != nil {
		return fmt.Errorf("start SMTP TLS: %w", err)
	}
	if err := client.Auth(smtp.PlainAuth("", s.username, s.password, s.host)); err != nil {
		return fmt.Errorf("authenticate with SMTP server: %w", err)
	}
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	if _, err := io.WriteString(writer, message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("send SMTP message: %w", err)
	}
	return nil
}

func (s *SMTPInvitationEmailSender) invitationURL(token string) (string, error) {
	base, err := url.Parse(s.appURL)
	if err != nil || (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" {
		return "", fmt.Errorf("invalid invitation application URL")
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/accept-invitation"
	query := base.Query()
	query.Set("token", token)
	base.RawQuery = query.Encode()
	return base.String(), nil
}
