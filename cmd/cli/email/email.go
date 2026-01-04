package email

import (
	"fmt"
	"net/smtp"

	"github.com/spf13/viper"
)

func Send(to []string, subject string, body string) error {
	from := viper.GetString("email.sender")
	password := viper.GetString("email.password")
	host := viper.GetString("email.host")
	port := viper.GetString("email.port")

	msg := "From: " + from + "\n" +
		"To: " + to[0] + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(fmt.Sprintf("%s:%d", host, port), auth, from, to, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}
