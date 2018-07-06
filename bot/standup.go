package bot

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
)

func (s *Slack) standupYesterday(ev *slack.MessageEvent) error {

	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedStandup := map[string]bool{
		"standup": true,
	}
	if acceptedStandup[text] {
		attachment := slack.Attachment{
			Text:       "What did you do yesterday?",
			Color:      "#2896b7",
			CallbackID: fmt.Sprintf("standupYesterday_%s", ev.User),
		}

		params := slack.PostMessageParameters{
			Attachments: []slack.Attachment{
				attachment,
			},
		}
		_, timestamp, err := s.Client.PostMessage(ev.Channel, "", params)
		if err != nil {
			s.Logger.Printf("failed to post yesterday standup question: %s\n", err)
		}
		s.Logger.Printf("[INFO] Timestamp of the standupYesterday message: %s\n", timestamp)

		timer := time.NewTimer(10 * time.Minute)
		ticker := time.NewTicker(5 * time.Second)

	loop:
		for {
			select {
			case <-timer.C:
				s.standupCancelTimeout(ev.Channel)
				if err != nil {
					s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
				} else {
					s.Logger.Printf("[DEBUG] Standup Canceled.\n")
				}
				break loop
			case <-ticker.C:
				params2 := slack.HistoryParameters{
					Count:  1,
					Oldest: timestamp,
				}

				history, err := s.Client.GetIMHistory(ev.Channel, params2)
				if err != nil {
					s.Logger.Printf("[ERROR] Could not get the IM history of message with timestamp %s: %s\n", timestamp, err)
				}
				s.Logger.Printf("[INFO] Getting IM history of message with timestamp %s\n", timestamp)

				message := history.Messages

				if len(message) == 0 {

				}
				if len(message) > 0 {
					text := history.Messages[0].Msg.Text
					t := time.Now().Local().Format("2006-01-02")
					t2 := time.Now().Local().Format("15:04:05")
					date := fmt.Sprintf(t)
					time := fmt.Sprintf(t2)
					userid := history.Messages[0].Msg.User
					stamp := history.Messages[0].Msg.Timestamp
					switch text {
					case "cancel":
						s.standupCancel(ev.Channel)
						if err != nil {
							s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
						}
						s.Logger.Printf("[INFO] Standup Canceled.\n")

						break loop
					default:
						s.standupYesterdayRegister(text, stamp, date, time, userid)
						if err != nil {
							s.Logger.Printf("[ERROR] Could not start standupYesterdayRegister: %+v\n", err)
						}
						s.Logger.Printf("[INFO] Starting standupYesterdayRegister\n")

						err = s.standupToday(ev.Channel, ev.User)
						if err != nil {
							s.Logger.Printf("[ERROR] Could not start standupToday: %s\n", err)
						}
						s.Logger.Printf("[INFO] Starting standupToday.\n")

						break loop
					}
				}
			}
		}
	}
	return nil
}

func (s *Slack) standupYesterdayScheduled(userid string) error {

	_, _, channelid, _ := s.Client.OpenIMChannel(userid)

	attachment := slack.Attachment{
		Text:       "What did you do yesterday?",
		Color:      "#2896b7",
		CallbackID: fmt.Sprintf("standupYesterday_%s", userid),
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}
	_, timestamp, err := s.Client.PostMessage(channelid, "", params)
	if err != nil {
		s.Logger.Printf("failed to post yesterday standup question: %s\n", err)
	}
	s.Logger.Printf("[INFO] Timestamp of the standupYesterday message: %s\n", timestamp)

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			s.standupCancelTimeout(channelid)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
			}
			s.Logger.Printf("[INFO] Standup Canceled.\n")

			break loop
		case <-ticker.C:
			params2 := slack.HistoryParameters{
				Count:  1,
				Oldest: timestamp,
			}

			history, err := s.Client.GetIMHistory(channelid, params2)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not get the IM history of message with timestamp %s: %s\n", timestamp, err)
			}
			s.Logger.Printf("[INFO] Getting IM history of message with timestamp %s\n", timestamp)

			message := history.Messages

			if len(message) == 0 {

			}
			if len(message) > 0 {
				text := history.Messages[0].Msg.Text
				t := time.Now().Local().Format("2006-01-02")
				t2 := time.Now().Local().Format("15:04:05")
				date := fmt.Sprintf(t)
				time := fmt.Sprintf(t2)
				userid := history.Messages[0].Msg.User
				stamp := history.Messages[0].Msg.Timestamp
				switch text {
				case "cancel":
					s.standupCancel(channelid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
					}
					s.Logger.Printf("[INFO] Standup Canceled.\n")

					break loop
				default:
					s.standupYesterdayRegister(text, stamp, date, time, userid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not start standupYesterdayRegister: %+v\n", err)
					}
					s.Logger.Printf("[INFO] Starting standupYesterdayRegister\n")

					err = s.standupToday(channelid, userid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not start standupToday: %s\n", err)
					}
					s.Logger.Printf("[INFO] Starting standupToday.\n")
					break loop
				}
			}
		}
	}
	return nil
}

func (s *Slack) standupYesterdayRegister(response, timestamp, date, time, userid string) error {

	s.Logger.Printf("[INFO] Starting import in database of standupYesterday result\n")

	var id string

	sqlWrite := `
	INSERT INTO 
		hatcher.standupyesterday 
		(response, timestamp, date, time, userid)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`

	err := database.DB.QueryRow(
		sqlWrite,
		response,
		timestamp,
		date,
		time,
		userid).Scan(&id)
	if err != nil {
		s.Logger.Printf("[ERROR] Couldn't insert in the database the result of standupYesterday: %s\n", err)
	}
	s.Logger.Printf("[INFO] standupYesterday result written in database.\n")
	return nil
}

func (s *Slack) standupToday(channelid, userid string) error {

	attachment := slack.Attachment{
		Text:       "What are you doing today?",
		Color:      "#41aa3f",
		CallbackID: "standupToday",
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}
	_, timestamp, err := s.Client.PostMessage(channelid, "", params)
	if err != nil {
		s.Logger.Printf("[ERROR] Failed to post today standup question: %s", err)
	}
	s.Logger.Printf("[INFO] Posting today standup question.\n")

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			s.standupCancelTimeout(channelid)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
			}
			s.Logger.Printf("[INFO] Standup Canceled.\n")
			break loop
		case <-ticker.C:
			params2 := slack.HistoryParameters{
				Count:  1,
				Oldest: timestamp,
			}

			history, err := s.Client.GetIMHistory(channelid, params2)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not get the IM history of message with timestamp %s: %s\n", timestamp, err)
			}
			s.Logger.Printf("[INFO] Getting IM history of message with timestamp %s\n", timestamp)

			message := history.Messages

			if len(message) == 0 {

			}
			if len(message) > 0 {
				text := history.Messages[0].Msg.Text
				t := time.Now().Local().Format("2006-01-02")
				t2 := time.Now().Local().Format("15:04:05")
				date := fmt.Sprintf(t)
				time := fmt.Sprintf(t2)
				userid := history.Messages[0].Msg.User
				stamp := history.Messages[0].Msg.Timestamp
				switch text {
				case "cancel":
					s.standupCancel(channelid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
					}
					s.Logger.Printf("[INFO] Canceled standup\n")

					break loop
				default:
					err := s.standupTodayRegister(text, stamp, date, time, userid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not start standupTodayRegister: %+v\n", err)
					}
					s.Logger.Printf("[INFO] Starting standupTodayRegister\n")

					err = s.standupBlocker(channelid, userid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not start standupBlocker: %s\n.", err)
					}
					s.Logger.Printf("[INFO] Started standupBlocker.\n")

					break loop
				}
			}
		}
	}
	return nil
}

func (s *Slack) standupTodayRegister(response, timestamp, date, time, userid string) error {

	s.Logger.Printf("[INFO] Starting import in database of standupToday result\n")

	var id string

	sqlWrite := `
	INSERT INTO 
		hatcher.standuptoday 
		(response, timestamp, date, time, userid)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`

	err := database.DB.QueryRow(
		sqlWrite,
		response,
		timestamp,
		date,
		time,
		userid).Scan(&id)
	if err != nil {
		s.Logger.Printf("[ERROR] Couldn't insert in the database the result of standupToday: %s\n", err)
	}
	s.Logger.Printf("[INFO] standupToday result written in database.\n")

	return nil
}

func (s *Slack) standupBlocker(channelid, userid string) error {

	attachment := slack.Attachment{
		Text:       "Do you have any blocker?",
		Color:      "#f91b1b",
		CallbackID: "standupBlocker",
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}
	_, timestamp, err := s.Client.PostMessage(channelid, "", params)
	if err != nil {
		s.Logger.Printf("[ERROR] Failed to post blocker standup question: %s", err)
	}
	s.Logger.Printf("[INFO] Posted blocker standup question.\n")

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			s.standupCancelTimeout(channelid)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
			}
			s.Logger.Printf("[INFO] Standup Canceled.\n")

			break loop
		case <-ticker.C:
			params2 := slack.HistoryParameters{
				Count:  1,
				Oldest: timestamp,
			}

			history, err := s.Client.GetIMHistory(channelid, params2)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not get the IM history of message with timestamp %s: %s\n", timestamp, err)
			}
			s.Logger.Printf("[INFO] Getting IM history of message with timestamp %s\n", timestamp)

			message := history.Messages

			if len(message) == 0 {

			}
			if len(message) > 0 {
				text := history.Messages[0].Msg.Text
				t := time.Now().Local().Format("2006-01-02")
				t2 := time.Now().Local().Format("15:04:05")
				date := fmt.Sprintf(t)
				time := fmt.Sprintf(t2)
				userid := history.Messages[0].Msg.User
				stamp := history.Messages[0].Msg.Timestamp
				switch text {
				case "cancel":
					err := s.standupCancel(channelid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not cancel standup: %+v\n", err)
					}
					s.Logger.Printf("[INFO] Standup canceled.\n")

					break loop
				default:
					err := s.standupBlockerRegister(text, stamp, date, time, userid)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not start standupBlockerRegister: %+v\n", err)
					}
					s.Logger.Printf("[INFO] Started standupBlockerRegister.\n")

					err = s.standupDone(channelid, userid, date)
					if err != nil {
						s.Logger.Printf("[ERROR] Could not start standupDone: %+v\n", err)
					}
					s.Logger.Printf("[INFO] Started standupDone.\n")

					break loop
				}
			}
		}
	}
	return nil
}

func (s *Slack) standupBlockerRegister(response, timestamp, date, time, userid string) error {

	s.Logger.Printf("[INFO] Starting import in database of standupBlocker result\n")

	var id string

	sqlWrite := `
	INSERT INTO 
		hatcher.standupblocker
		(response, timestamp, date, time, userid)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`

	err := database.DB.QueryRow(
		sqlWrite,
		response,
		timestamp,
		date,
		time,
		userid).Scan(&id)
	if err != nil {
		s.Logger.Printf("[ERROR] Couldn't insert in the database the result of standupBlocker: %s\n", err)
	}
	s.Logger.Printf("[INFO] standupBlocker result written in database.\n")

	return nil
}

func (s *Slack) standupCancel(channelid string) error {

	attachment := slack.Attachment{
		Text:       "Standup canceled!",
		Color:      "#f91b1b",
		CallbackID: "standupCancel",
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}
	_, _, err := s.Client.PostMessage(channelid, "", params)
	if err != nil {
		s.Logger.Printf("[ERROR] failed to post standup canceled message: %s\n", err)
	}
	s.Logger.Printf("[INFO] Posted standup canceled message.\n")

	return nil
}

func (s *Slack) standupCancelTimeout(channelid string) error {

	attachment := slack.Attachment{
		Text:       "The standup was canceled for timeout.\nYou can restart your standup by sending `standup`",
		Color:      "#f91b1b",
		CallbackID: "standupCancelTimeout",
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}
	_, _, err := s.Client.PostMessage(channelid, "", params)
	if err != nil {
		s.Logger.Printf("[ERROR] failed to post standup canceled timeout message: %s\n", err)
	}
	s.Logger.Printf("[INFO] Posted standup canceled timeout message.\n")

	return nil
}

func (s *Slack) standupDone(channelid, userid, date string) error {

	attachment := slack.Attachment{
		Text:       "Standup Done! Thanks and see you tomorrow :smiley:",
		Color:      "#2896b7",
		CallbackID: "standupDone",
	}

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			attachment,
		},
	}
	_, _, err := s.Client.PostMessage(channelid, "", params)
	if err != nil {
		s.Logger.Printf("[ERROR] Failed to post standup done message: %s\n", err)
	}
	s.Logger.Printf("[INFO] Posted standup done message.\n")

	err = s.postStandupResults(userid, date)
	if err != nil {
		s.Logger.Printf("[ERROR] Could not start postStandup: %s", err)
	}
	s.Logger.Printf("[INFO] Started postStandup")

	return nil
}

func (s *Slack) postStandupResults(userid, date string) error {

	rows, err := database.DB.Query("SELECT userid, displayname, standup_channel FROM hatcher.users WHERE userid = $1;", userid)
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Printf("[ERROR] There is no results.\n")
		}
	}
	defer rows.Close()
	for rows.Next() {

		var displayname string
		var standupChannel string

		responseYesterday := s.standupResultsYesterday(userid, date, standupChannel)
		responseToday := s.standupResultsToday(userid, date, standupChannel)
		responseBlocker := s.standupResultsBlocker(userid, date, standupChannel)

		err = rows.Scan(&userid, &displayname, &standupChannel)
		if err != nil {
			s.Logger.Printf("[ERROR] During the scan.\n")
		}

		attachment := slack.Attachment{
			Pretext:    fmt.Sprintf("%s posted a daily standup note", displayname),
			Title:      "What did you do yesterday?",
			Text:       fmt.Sprintf("%s", responseYesterday),
			Color:      "#2896b7",
			CallbackID: fmt.Sprintf("resultsStandupYesterday_%s", userid),
		}

		attachment2 := slack.Attachment{
			Title:      "What are you doing today?",
			Color:      "#41aa3f",
			Text:       fmt.Sprintf("%s", responseToday),
			CallbackID: fmt.Sprintf("resultsStandupToday_%s", userid),
		}

		attachment3 := slack.Attachment{
			Title:      "Do you have any blocker?",
			Color:      "#f91b1b",
			Text:       fmt.Sprintf("%s", responseBlocker),
			CallbackID: fmt.Sprintf("resultsStandupBlocker_%s", userid),
		}

		params := slack.PostMessageParameters{
			Attachments: []slack.Attachment{
				attachment,
				attachment2,
				attachment3,
			},
		}
		_, _, err := s.Client.PostMessage(standupChannel, "", params)
		if err != nil {
			s.Logger.Printf("[ERROR] Failed to post standup results: %s\n", err)
		}
		s.Logger.Printf("[INFO] Standup posted.\n")

	}
	return nil
}

func (s *Slack) standupResultsYesterday(userid, date, standupChannel string) (responseYesterday string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standupyesterday WHERE userid = $1 and date = $2;", userid, date)
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Printf("[ERROR] There is no results.\n")
		}
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&response)
		if err != nil {
			s.Logger.Printf("[ERROR] During the scan.\n")
		}
	}
	return response
}

func (s *Slack) standupResultsToday(userid, date, standupChannel string) (responseToday string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standuptoday WHERE userid = $1 and date = $2;", userid, date)
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Printf("[ERROR] There is no results.\n")
		}
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&response)
		if err != nil {
			s.Logger.Printf("[ERROR] During the scan.\n")
		}
	}
	return response
}

func (s *Slack) standupResultsBlocker(userid, date, standupChannel string) (responseBlocker string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standupblocker WHERE userid = $1 and date = $2;", userid, date)
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Printf("[ERROR] There is no results.\n")
		}
	}
	defer rows.Close()
	for rows.Next() {

		err = rows.Scan(&response)
		if err != nil {
			s.Logger.Printf("[ERROR] During the scan.\n")
		}
	}
	return response
}
