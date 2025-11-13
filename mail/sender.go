package mail

import (
	"github.com/rs/zerolog/log"
	"github.com/wneessen/go-mail"
)

const (
	smtpHost = "smtp.gmail.com"
	smtpPort = 587
)

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func (sender *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {

	message := mail.NewMsg()

	if err := message.From(sender.fromEmailAddress); err != nil {
		log.Error().
			Err(err).
			Msg("failed to set From email address")

		return err
	}

	if err := message.To(to...); err != nil {
		log.Error().
			Err(err).
			Msg("failed to set To email address(es)")

		return err
	}

	if err := message.Cc(cc...); err != nil {
		log.Error().
			Err(err).
			Msg("failed to set Cc email address(es)")

		return err
	}

	if err := message.Bcc(bcc...); err != nil {
		log.Error().
			Err(err).
			Msg("failed to set Bcc email address(es)")

		return err
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, content)

	// file attachments
	// file opened on Send() call so no error returned here
	for _, f := range attachFiles {
		message.AttachFile(f)
	}

	client, err := mail.NewClient(smtpHost,
		mail.WithPort(smtpPort),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(sender.fromEmailAddress),
		mail.WithPassword(sender.fromEmailPassword),
	)

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create mail client")

		return err
	}

	if err := client.DialAndSend(message); err != nil {
		log.Error().
			Err(err).
			Msg("failed to send message")

		return err

	}

	return nil
}

func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}
