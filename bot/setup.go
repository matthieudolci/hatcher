package bot

import (
	"fmt"
	"strings"

	"github.com/nlopes/slack"
)

func (s *Slack) askSetup(ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedSetup := map[string]bool{
		"setup": true,
		"init":  true,
	}

	if acceptedSetup[text] {
		params := slack.PostMessageParameters{}
		attachment := slack.Attachment{
			Text:       "Do you want to setup/update your user with the bot Hatcher?",
			CallbackID: fmt.Sprintf("setup_%s", ev.User),
			Color:      "#AED6F1",
			Actions: []slack.AttachmentAction{
				slack.AttachmentAction{
					Name:  "SetupYes",
					Text:  "Yes",
					Type:  "button",
					Value: "SetupYes",
					Style: "primary",
				},
				slack.AttachmentAction{
					Name:  "SetupNo",
					Text:  "No",
					Type:  "button",
					Value: "SetupNo",
					Style: "danger",
				},
			},
		}
		params.Attachments = []slack.Attachment{attachment}
		params.User = ev.User
		params.AsUser = true

		_, err := s.Client.PostEphemeral(
			ev.Channel,
			ev.User,
			slack.MsgOptionAttachments(params.Attachments...),
			slack.MsgOptionPostMessageParameters(params),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Slack) askRemove(ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedRemove := map[string]bool{
		"remove": true,
		"delete": true,
	}

	if acceptedRemove[text] {
		params := slack.PostMessageParameters{}
		attachment := slack.Attachment{
			Text:       "Do you want to delete your user?",
			CallbackID: fmt.Sprintf("remove_%s", ev.User),
			Color:      "#FF0000",
			Actions: []slack.AttachmentAction{
				slack.AttachmentAction{
					Name:  "RemoveYes",
					Text:  "Yes",
					Type:  "button",
					Value: "RemoveYes",
					Style: "primary",
				},
				slack.AttachmentAction{
					Name:  "RemoveNo",
					Text:  "No",
					Type:  "button",
					Value: "RemoveNo",
					Style: "danger",
				},
			},
		}
		params.Attachments = []slack.Attachment{attachment}
		params.User = ev.User
		params.AsUser = true

		_, err := s.Client.PostEphemeral(
			ev.Channel,
			ev.User,
			slack.MsgOptionAttachments(params.Attachments...),
			slack.MsgOptionPostMessageParameters(params),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Slack) askWhoIsManager(channelid, userid string) error {

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "Who is your manager?",
		CallbackID: fmt.Sprintf("manager_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			{
				Name:       "ManagerChosen",
				Text:       "Type to filter option",
				Type:       "select",
				DataSource: "users",
			},
		},
	}
	params.Attachments = []slack.Attachment{attachment}
	params.User = userid
	params.AsUser = true

	_, err := s.Client.PostEphemeral(
		channelid,
		userid,
		slack.MsgOptionAttachments(params.Attachments...),
		slack.MsgOptionPostMessageParameters(params),
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Slack) askIfManager(channelid, userid string) error {

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "Are you a manager?",
		CallbackID: fmt.Sprintf("ismanager_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "isManagerYes",
				Text:  "Yes",
				Type:  "button",
				Value: "isManagerYes",
				Style: "primary",
			},
			slack.AttachmentAction{
				Name:  "isManagerNo",
				Text:  "No",
				Type:  "button",
				Value: "isManagerNo",
				Style: "danger",
			},
		},
	}
	params.Attachments = []slack.Attachment{attachment}
	params.User = userid
	params.AsUser = true

	_, err := s.Client.PostEphemeral(
		channelid,
		userid,
		slack.MsgOptionAttachments(params.Attachments...),
		slack.MsgOptionPostMessageParameters(params),
	)
	if err != nil {
		return err
	}
	return nil
}
