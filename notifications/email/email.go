package email

import (
	"fmt"
	"os"
	"strings"

	"github.com/CPU-commits/USACH.dev-Server/settings"
	"gopkg.in/gomail.v2"
)

// Settings
var settingsData = settings.GetSettings()

// Templates
const (
	TEMPLATE_VALIDATE_USER = "validate_user"
)

// MailSender
type MailConfig struct {
	From           string
	To             string
	Subject        string
	TextHtml       string
	Template       string
	TemplateParams map[string]string
}

func getTemplateText(fileName string, params map[string]string) (string, error) {
	file, err := os.ReadFile(fmt.Sprintf("notifications/email/templates/%v.html", fileName))
	if err != nil {
		return "", nil
	}
	fileStr := string(file)
	for key := range params {
		fileStr = strings.ReplaceAll(fileStr, key, params[key])
	}
	return fileStr, nil
}

func SendEmail(mailConfig *MailConfig) error {
	// Init new message
	msg := gomail.NewMessage()

	msg.SetHeader("From", mailConfig.From)
	msg.SetHeader("To", mailConfig.To)
	msg.SetHeader("Subject", mailConfig.Subject)
	// Get text message
	var message string
	if mailConfig.Template != "" {
		textFile, err := getTemplateText(
			mailConfig.Template,
			mailConfig.TemplateParams,
		)
		if err != nil {
			return err
		}
		message = textFile
	} else {
		message = mailConfig.TextHtml
	}
	msg.SetBody("text/html", message)
	// Set dialer
	n := gomail.NewDialer(
		settingsData.SMTP_HOST,
		settingsData.SMTP_PORT,
		settingsData.SMTP_USER,
		settingsData.SMTP_PASSWORD,
	)

	// Send
	if err := n.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}
