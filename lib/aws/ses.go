// Package to help with making SES AWS calls
package aws

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"

	c "rob/lib/common/constants"

	log "github.com/sirupsen/logrus"
)

var (
	sesRegion      = "us-west-2"
	client         *ses.SES
	defaultCharset = "UTF-8"
	// This should be enabled in main.go for emails to actually be sent
	// By default it's disabled, so that we do not trigger emails in tests
	DisableModule = true
)

func SendEmail(from string, to []string, subject, bodyText, bodyHtml string) error {
	funcName := "aws/ses.go: SendEmail"
	log.WithFields(log.Fields{
		"from":      from,
		"to_length": len(to),
		"subject":   subject,
		// Intentionally not logging all to addresses, bodyText, etc
		// Those additional fields will never help with debug but only
		// create extra noise in logs
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	if DisableModule {
		log.Debug("Email module disabled.")
		// Return no error as this is intended
		return nil
	}

	if len(to) > 50 {
		return errors.New("Cannot send more than 50 emails at once.")
	}

	toAddresses := make([]*string, len(to))
	for i, a := range to {
		toAddresses[i] = s(a)
	}

	input := &ses.SendEmailInput{
		Source: s(from),
		Destination: &ses.Destination{
			ToAddresses: toAddresses,
		},
		Message: &ses.Message{
			Subject: getContent(subject),
			Body: &ses.Body{
				Text: getContent(bodyText),
				Html: getContent(bodyHtml),
			},
		},
	}

	_, err := client.SendEmail(input)
	return err
}

func s(p string) *string {
	return aws.String(p)
}

func getContent(p string) *ses.Content {
	return &ses.Content{
		Charset: s(defaultCharset),
		Data:    s(p),
	}
}

// Sends out verification email
func VerificationEmail(email, firstName, code string) error {
	sub := "Twiq - Verify your account"
	htmlBody := fmt.Sprintf("Hello %s,<br><br>Welcome to Twiq.<br>You can verify your Twiq account by <b><a href='https://twiq.in/api/vr?token=%s' target='_blank'>clicking here</a></b>.<br><br>If the above link did not work, you can manually enter the below code in Twiq App.<br>Code: <b>%s</b><br>", firstName, code, code)

	plainBody := fmt.Sprintf("Hello %s,\n\nWelcome to Twiq.\nYou can verify your Twiq account by entering the code below in Twiq App.\nCode: %s\n", firstName, code)
	return SendEmail(c.EmailInfo, []string{email}, sub, plainBody, htmlBody)
}

func ForgotPasswordEmail(email, firstName, code string) error {
	sub := "Twiq - Resetting the account password"
	body := fmt.Sprintf("Hello %s , \n Your reset token is %s.\n Please enter the above code in the reset screen", firstName, code)
	return SendEmail(c.EmailInfo, []string{email}, sub, body, body)
}

// Can be used to check if the mail server is working fine
// This will send an email from c.EmailInfo to c.EmailInfo
func SendTestEmail() {
	log.Debug("Entered SendTestEmail")
	err := SendEmail(c.EmailInfo, []string{c.EmailInfo}, "testing",
		"This is testing body", "This is <b>testing html</b> body")
	if err != nil {
		log.Errorf("Failed to send test email: %s", err.Error())
		return
	}
	log.Debug("Successfully sent email")
}

func init() {
	creds := credentials.NewSharedCredentials(c.AWSCredsFile, c.AWSSESProfile)
	config := aws.NewConfig().WithCredentials(creds).WithRegion(sesRegion)
	sess := session.Must(session.NewSession(config))
	client = ses.New(sess)
	snsClient = sns.New(sess)
}
