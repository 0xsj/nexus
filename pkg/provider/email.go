package provider

import (
	"github.com/0xsj/nexus-go/pkg/email"
	"github.com/0xsj/nexus-go/pkg/email/smtp"
)

// ProvideEmailSender creates an SMTP email sender.
func ProvideEmailSender(config email.Config) email.Sender {
	return smtp.New(config)
}
