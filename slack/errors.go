package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

func renderSlackError(err error) error {
	if err == nil {
		return nil
	}

	if slackErr, ok := err.(slack.SlackErrorResponse); ok {
		return fmt.Errorf("%w %v %v", slackErr, slackErr.ResponseMetadata.Messages, slackErr.ResponseMetadata.Warnings)
	}
	return err
}
