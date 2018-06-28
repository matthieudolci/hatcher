package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"

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
			return err
		}
	}
	return nil
}

// Insert into the database the result of the happiness survey
func (s *Slack) resultHappinessSurvey(userid, result string) {

	sqlWrite := `
	INSERT INTO hatcher.happiness (user_id, results)
	VALUES ($1, $2)
	RETURNING id`

	err := database.DB.QueryRow(sqlWrite, userid, result).Scan(&userid)
	if err != nil {
		s.Logger.Printf("[ERROR] Couldn't insert in the database the result of the happiness survey for user ID %s.\n %s", userid, err)
	}
	s.Logger.Printf("[DEBUG] Happiness Survey Result written in database.\n")
}

// GetTimeAndUsersHappinessSurvey gets the time selected by a user for the Happiness survey
func (s *Slack) GetTimeAndUsersHappinessSurvey() error {
	type ScheduleData struct {
		Times  string
		UserID string
	}

	rows, err := database.DB.Query("SELECT to_char(happiness_schedule, 'HH24:MI'), user_id FROM hatcher.users;")
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Printf("[ERROR] There is no result time or user_id.\n")
		} else {
			panic(err)
		}
	}
	defer rows.Close()
	gocron.Clear()
	for rows.Next() {
		scheduledata := ScheduleData{}
		err = rows.Scan(&scheduledata.Times, &scheduledata.UserID)
		if err != nil {
			s.Logger.Printf("[ERROR] During the scan.\n")
		}
		fmt.Println(scheduledata)
		s.runHappinessSurveySchedule(scheduledata.Times, scheduledata.UserID)
	}
	channel := make(chan int)
	go startCron(channel)
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		s.Logger.Printf("[ERROR] During iteration.\n")
	}
	return nil
}

// Runs the job askHappinessSurveyScheduled at a time defined by the user
func (s *Slack) runHappinessSurveySchedule(times, userid string) {
	location, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		log.Println("Unfortunately can't load a location")
		log.Println(err)
	} else {
		gocron.ChangeLoc(location)
	}
	gocron.Every(1).Day().At(times).Do(s.askHappinessSurveyScheduled, userid)
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
		return err
	}
	fmt.Printf("Scheduled happyness survey for user %s posted", userid)
	return nil
}

// Starts gocron
func startCron(channel chan int) {
	<-gocron.Start()
}
