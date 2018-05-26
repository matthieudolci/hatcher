package bot

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/nlopes/slack"
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

func (s *Slack) askForHappinessResult(ev *slack.MessageEvent, rtm *slack.RTM) error {

	config := dbConfig()
	var response string
	userid := ev.User
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedHappinessResults := map[string]bool{
		"results today":     true,
		"results yesterday": true,
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport], config[dbuser], config[dbpass], config[dbname])

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var r string
	// check if the user is a manager. If a manager can get surhey result. If
	err = db.QueryRow("SELECT is_manager FROM hatcher.users WHERE user_id = $1", userid).Scan(&r)
	if err != nil {
		fmt.Println(err.Error())
	}
	if r == "true" {
		if acceptedHappinessResults[text] {
			params := slack.PostMessageParameters{}
			attachment := slack.Attachment{
				Text:       "Select a user:",
				CallbackID: fmt.Sprintf("happiness_%s-%s", ev.User, text),
				Color:      "#AED6F1",
				Actions: []slack.AttachmentAction{
					{
						Name:       "ResultPosted",
						Text:       "Type to filter option",
						Type:       "select",
						DataSource: "users",
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
	}
	if r == "false" {
		response = "Sorry you are not a manager and can't get results"
		rtm.SendMessage(rtm.NewOutgoingMessage(response, ev.Channel))
	}

	return nil
}
