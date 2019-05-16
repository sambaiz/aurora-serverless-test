package secret

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func GetSecretString(id string) (string, error) {
	sess := session.New()
	svc := secretsmanager.New(sess, aws.NewConfig())
	output, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(id),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		return "", err
	}
	if output.SecretString == nil {
		return "", fmt.Errorf("SecretString is nil. id: %s", id)
	}
	return *output.SecretString, nil
}
