package bot

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

func (s *Slack) askHelp(ev *slack.MessageEvent) error {

	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedHelp := map[string]bool{
		"help": true,
	}
	if acceptedHelp[text] {
		attachment := slack.Attachment{
			Pretext: "Welcome to the help center",
			Color:   "#87cefa",
			Text:    "Here are the different commands you can pass to Hatcher:",
			Fields: []slack.AttachmentField{
				{
					Title: "hello",
					Value: "The command `hello` will start the setup of your user with Hatcher\nThis command needs to be passed before any other ones",
				},
				{
					Title: "standup",
					Value: "The command `standup` starts a new standup in case you missed the scheduled one",
				},
				{
					Title: "happiness setup",
					Value: "The command `happiness setup` enables the daily happiness survey",
				},
				{
					Title: "happiness remove",
					Value: "The command `happiness remove` remove your user from the daily happiness survey",
				},
			},
			CallbackID: fmt.Sprintf("askHelp_%s", ev.User),
		}

		params := slack.PostMessageParameters{
			Attachments: []slack.Attachment{
				attachment,
			},
		}
		_, _, err := s.Client.PostMessage(ev.Channel, "", params)
		if err != nil {
			log.WithError(err).Error("Failed to post help")
		}
		log.WithFields(log.Fields{
			"userid": ev.User,
		}).Info("Help message posted")
	}
	return nil
}
