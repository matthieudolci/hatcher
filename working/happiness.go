package main

import (
	"fmt"
	"strings"

	"github.com/nlopes/slack"
)

// respondHowAreYou trigger the happiness survey
func respondHowAreYou(api *slack.Client, rtm *slack.RTM, msg *slack.MessageEvent, prefix string) error {

	text := msg.Text
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedHowAreYou := map[string]bool{
		"how's it going?":    true,
		"how are you?":       true,
		"feeling okay?":      true,
		"how are you doing?": true,
	}

	if acceptedHowAreYou[text] {
		params := slack.PostMessageParameters{}
		attachment := slack.Attachment{
			Text:       "I am good. How are you today?",
			CallbackID: fmt.Sprintf("ask_%s", msg.User),
			Color:      "#AED6F1",
			Actions: []slack.AttachmentAction{
				slack.AttachmentAction{
					Name:  "action",
					Text:  ":smiley:",
					Type:  "button",
					Value: "good",
				},
				slack.AttachmentAction{
					Name:  "action",
					Text:  ":neutral_face:",
					Type:  "button",
					Value: "neutral",
				},
				slack.AttachmentAction{
					Name:  "action",
					Text:  ":cry:",
					Type:  "button",
					Value: "sad",
				},
			},
		}

		params.Attachments = []slack.Attachment{attachment}
		params.User = msg.User
		params.AsUser = true

		_, err := api.PostEphemeral(
			msg.Channel,
			msg.User,
			slack.MsgOptionAttachments(params.Attachments...),
			slack.MsgOptionPostMessageParameters(params),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
