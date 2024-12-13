package main

import (
	"net"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type Mailer interface {
	SendEmail(*email.Email) error
}

type SMTPMailer struct {
	serverHost, serverPort, username, password, fromEmail string
}

func NewSMTPMailer(serverAddr, username, password, fromEmail string) (*SMTPMailer, error) {
	serverHost, serverPort, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, err
	}
	return &SMTPMailer{serverHost, serverPort, username, password, fromEmail}, nil
}

func (mailer *SMTPMailer) SendEmail(email *email.Email) error {
	serverAddr := net.JoinHostPort(mailer.serverHost, mailer.serverPort)
	email.From = mailer.fromEmail
	return email.Send(serverAddr, smtp.PlainAuth("", mailer.username, mailer.password, mailer.serverHost))
}

type TestingMailer struct {
	email *email.Email
}

func (mailer *TestingMailer) SendEmail(email *email.Email) error {
	mailer.email = email
	return nil
}

func (mailer *TestingMailer) CheckEmail() *email.Email {
	ret := mailer.email
	mailer.email = nil
	return ret
}
