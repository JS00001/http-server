package lib

import (
	"http-server/config"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

var AwsSession *session.Session
var SesSession *ses.SES

func SendMail(to, subject, body string) error {
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(config.GetConfig("AWS_SES_SENDER")),
	}

	_, err := SesSession.SendEmail(input)

	// print error
	if err != nil {
		log.Println(err)
	}

	return err
}
