package main

import (
	"net/mail"
	"os"

	"github.com/resend/resend-go/v2"
)

func SendCustomEmail(subject string, recp []mail.Address, data string) error {
	apiKey := os.Getenv("RESEND_API")
	senderEmail := os.Getenv("RESEND_EMAIL")
	replyToAddr := os.Getenv("REPLY_TO_EMAIL")

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    senderEmail,
		To:      mailAddressesToStrings(recp),
		ReplyTo: replyToAddr,
		Subject: subject,
		Html:    data,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return err
	}
	return nil
}

func mailAddressesToStrings(addresses []mail.Address) []string {
	stringAddresses := make([]string, len(addresses))

	for i, addr := range addresses {
		// Use the Address.Address field to get the email address without <>
		stringAddresses[i] = addr.Address
	}

	return stringAddresses
}
