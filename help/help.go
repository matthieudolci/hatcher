package help

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/common"
	"github.com/nlopes/slack"
)

func AskHelp(s *common.Slack, ev *slack.MessageEvent) error {

	m := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	if len(m) == 0 || m[0] != "help" {
		n := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[:1]
		if len(n) == 0 || n[0] != "help" {
			return fmt.Errorf("The message doesn't contain help")
		}
	}

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
				Title: "happiness setup",
				Value: "The command `happiness setup` enables the daily happiness survey\nThis command needs to be pass in a direct message to Hatcher",
			},
			{
				Title: "happiness remove",
				Value: "The command `happiness remove` remove your user from the daily happiness survey\nThis command needs to be pass in a direct message to Hatcher",
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

	return nil
}
