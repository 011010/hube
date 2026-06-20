package external

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPClient struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPClient(host, port, user, pass, from string) *SMTPClient {
	if port == "" {
		port = "587"
	}
	return &SMTPClient{host: host, port: port, user: user, pass: pass, from: from}
}

func (c *SMTPClient) Send(to []string, subject, body string) error {
	addr := c.host + ":" + c.port

	msg := "From: " + c.from + "\r\n" +
		"To: " + strings.Join(to, ", ") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body

	auth := smtp.PlainAuth("", c.user, c.pass, c.host)

	// Try STARTTLS first (port 587), fall back to TLS (port 465)
	if c.port == "465" {
		tlsCfg := &tls.Config{ServerName: c.host}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, c.host)
		if err != nil {
			return err
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return err
		}
		if err := client.Mail(c.from); err != nil {
			return err
		}
		for _, r := range to {
			if err := client.Rcpt(r); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, c.from, to, []byte(msg))
}
