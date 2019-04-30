package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"
	log "github.com/sirupsen/logrus"
)

var (
	DisableSmsModule = true
	snsClient        *sns.SNS
)

func SendSms(num, body string) error {
	funcName := "aws/sns.go:SendSms"
	log.WithFields(log.Fields{
		"num":  num,
		"body": body,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	input := &sns.PublishInput{
		Message:     s(body),
		PhoneNumber: s("+91" + num),
	}
	log.Debugf("sns input: %+v", input)

	if DisableSmsModule {
		log.Debugf("SNS module disabled. Exiting")
		return nil
	}

	output, err := snsClient.Publish(input)

	log.Debugf("sns output: %+v %v", output, err)
	return err
}

func SendOtp(num, code string) error {
	body := fmt.Sprintf("Use %s to verify your Twiq account. #twiq", code)
	return SendSms(num, body)
}

func SendResetOtp(num, code string) error {
	body := fmt.Sprintf("Use %s to reset your Twiq account password", code)
	return SendSms(num, body)
}
