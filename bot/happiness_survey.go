package bot

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
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
			log.WithFields(log.Fields{
				"channelid": ev.Channel,
				"userid":    ev.User,
			}).WithError(err).Error("Could not post askHappinessSurvey question")
		}
		log.WithFields(log.Fields{
			"channelid": ev.Channel,
			"userid":    ev.User,
		}).Info("askHappinessSurvey question posted")
	}
	return nil
}

// Insert into the database the result of the happiness survey
func (s *Slack) resultHappinessSurvey(userid, result string) error {

	t := time.Now().Local().Format("2006-01-02")
	t2 := time.Now().Local().Format("15:04:05")
	date := fmt.Sprintf(t)
	time := fmt.Sprintf(t2)

	sqlWrite := `
	INSERT INTO hatcher.happiness (userid, results, date, time)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	err := database.DB.QueryRow(sqlWrite, userid, result, date, time).Scan(&userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid": userid,
		}).WithError(err).Error("Couldn't insert in the database the result of the happiness survey")
	}
	log.WithFields(log.Fields{
		"userid": userid,
	}).Info("Happiness Survey Result written in database")
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
		log.WithFields(log.Fields{
			"channelid": channelid,
			"userid":    userid,
		}).WithError(err).Error("Could not post askHappinessSurveyScheduled message")
	}
	log.WithFields(log.Fields{
		"channelid": channelid,
		"userid":    userid,
	}).Info("Message for askHappinessSurveyScheduled posted")
	return nil
}
