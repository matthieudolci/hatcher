package bot

import (
	"fmt"
	"strings"

	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
)

// Ask how are the users doing
func (s *Slack) askHappinessSurvey(ev *slack.MessageEvent) error {
	text := ev.Text
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
			CallbackID: fmt.Sprintf("ask_%s", ev.User),
			Color:      "#AED6F1",
			Actions: []slack.AttachmentAction{
				slack.AttachmentAction{
					Name:  "happinessGood",
					Text:  ":smiley:",
					Type:  "button",
					Value: "happinessGood",
				},
				slack.AttachmentAction{
					Name:  "happinessNeutral",
					Text:  ":neutral_face:",
					Type:  "button",
					Value: "happinessNeutral",
				},
				slack.AttachmentAction{
					Name:  "happinessSad",
					Text:  ":cry:",
					Type:  "button",
					Value: "happinessSad",
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
			s.Logger.Printf("[ERROR] Could not post askHappinessSurvey question: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] askHappinessSurvey question posted.\n")
		}
	}
	return nil
}

// Insert into the database the result of the happiness survey
func (s *Slack) resultHappinessSurvey(userid, result string) error {

	sqlWrite := `
	INSERT INTO hatcher.happiness (userid, results)
	VALUES ($1, $2)
	RETURNING id`

	err := database.DB.QueryRow(sqlWrite, userid, result).Scan(&userid)
	if err != nil {
		s.Logger.Printf("[ERROR] Couldn't insert in the database the result of the happiness survey for user ID %s.\n %s", userid, err)
	} else {
		s.Logger.Printf("[DEBUG] Happiness Survey Result written in database.\n")
	}
	return nil
}

// Ask how are the users doing
func (s *Slack) askHappinessSurveyScheduled(userid string) error {

	_, _, channelid, _ := s.Client.OpenIMChannel(userid)
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "How are you today?",
		CallbackID: fmt.Sprintf("ask_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "happinessGood",
				Text:  ":smiley:",
				Type:  "button",
				Value: "happinessGood",
			},
			slack.AttachmentAction{
				Name:  "happinessNeutral",
				Text:  ":neutral_face:",
				Type:  "button",
				Value: "happinessNeutral",
			},
			slack.AttachmentAction{
				Name:  "happinessSad",
				Text:  ":cry:",
				Type:  "button",
				Value: "happinessSad",
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
		s.Logger.Printf("[ERROR] Could not post askHappinessSurveyScheduled message: %s\n", err)
	} else {
		s.Logger.Printf("[DEBUG] Message for askHappinessSurveyScheduled posted.\n")
	}
	return nil
}
