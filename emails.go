package main

import (
	"net/mail"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

func SendCustomEmail(subject string, recp []mail.Address, data string, app *pocketbase.PocketBase) error {
	message := &mailer.Message{
		From: mail.Address{
			Address: app.Settings().Meta.SenderAddress,
			Name:    app.Settings().Meta.SenderName,
		},
		To:      recp,
		Subject: subject,
		HTML:    data,
		// bcc, cc, attachments and custom headers are also supported...
	}

	return app.NewMailClient().Send(message)
}

/**
func mailAddressesToStrings(addresses []mail.Address) []string {
	stringAddresses := make([]string, len(addresses))

	for i, addr := range addresses {
		// Use the Address.Address field to get the email address without <>
		stringAddresses[i] = addr.Address
	}

	return stringAddresses
}
*/
