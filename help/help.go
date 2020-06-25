package help

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/common"
	"github.com/slack-go/slack"
)

func AskHelp(s *common.Slack, ev *slack.MessageEvent) error {

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
					Value: "The command `standup` starts a new standup in case you missed the scheduled one\nThis command needs to be pass in a direct message to Hatcher",
				},
				{
					Title: "remove",
					Value: "The command `remove` removes your user from Hatcher\nThis command needs to be pass in a direct message to Hatcher",
				},
			},
			CallbackID: fmt.Sprintf("askHelp_%s", ev.User),
		}

		params := slack.MsgOptionAttachments(attachment)

		_, _, err := s.Client.PostMessage(ev.Channel, params)
		if err != nil {
			log.WithError(err).Error("Failed to post help")
		}
		log.WithFields(log.Fields{
			"userid": ev.User,
		}).Info("Help message posted")
	}
	return nil
}
