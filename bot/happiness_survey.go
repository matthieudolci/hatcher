package bot

import (
	"fmt"
	"strings"

	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

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
			return err
		}
	}
	return nil
}

// Insert into the database the result of the happiness survey
func (s *Slack) resultHappinessSurvey(userid, result string) {

	sqlWrite := `
	INSERT INTO hatcher.happiness (user_id, result)
	VALUES ($1, $2)
	RETURNING id`

	if err := database.DB.QueryRow(sqlWrite, userid, result).Scan(&userid); err != nil {
		err = errors.Wrapf(err,
			"Couldn't insert in the database the result of the happiness survey for user ID %s.\n", userid)
		return
	}

	s.Logger.Printf("[DEBUG] Happiness Survey Result written in database.\n")
}
