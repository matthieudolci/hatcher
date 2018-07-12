package bot

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
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
			log.WithError(err).Error("Failed to post yesterday standup question")
		}
		log.WithFields(log.Fields{
			"userid":    ev.User,
			"timestamp": timestamp,
		}).Info("Timestamp of the standupYesterday message")

		timer := time.NewTimer(10 * time.Minute)
		ticker := time.NewTicker(5 * time.Second)

	loop:
		for {
			select {
			case <-timer.C:
				s.standupCancelTimeout(ev.Channel)
				if err != nil {
					log.WithError(err).Error("Could not cancel standup")
				}
				log.Info("Standup Canceled")
				break loop
			case <-ticker.C:
				params2 := slack.HistoryParameters{
					Count:  1,
					Oldest: timestamp,
				}

				history, err := s.Client.GetIMHistory(ev.Channel, params2)
				if err != nil {
					log.WithFields(log.Fields{
						"timestamp": timestamp,
					}).WithError(err).Error("Could not get the IM history of the message")
				}
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).Debug("Getting IM history of the message")

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
							log.WithError(err).Error("Could not cancel standup")
						}
						log.Info("Standup Canceled")

						break loop
					default:
						s.standupYesterdayRegister(text, stamp, date, time, userid)
						if err != nil {
							log.WithError(err).Error("Could not start standupYesterdayRegister")
						}
						log.Info("Starting standupYesterdayRegister")

						err = s.standupToday(ev.Channel, ev.User)
						if err != nil {
							log.WithError(err).Error("Could not start standupToday")
						}
						log.Info("Starting standupToday")

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
		log.WithError(err).Error("Failed to post yesterday standup question")
	}
	log.WithFields(log.Fields{
		"timestamp": timestamp,
	}).Info("Timestamp of the standupYesterday message")

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			s.standupCancelTimeout(channelid)
			if err != nil {
				log.WithError(err).Error("Could not cancel standup")
			}
			log.Info("Standup Canceled")

			break loop
		case <-ticker.C:
			params2 := slack.HistoryParameters{
				Count:  1,
				Oldest: timestamp,
			}

			history, err := s.Client.GetIMHistory(channelid, params2)
			if err != nil {
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).WithError(err).Error("Could not get the IM history of the message")
			}
			log.WithFields(log.Fields{
				"timestamp": timestamp,
			}).Debug("Getting IM history of the message with timestamp")

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
						log.WithError(err).Error("Could not cancel standup")
					}
					log.Info("Standup Canceled")

					break loop
				default:
					s.standupYesterdayRegister(text, stamp, date, time, userid)
					if err != nil {
						log.WithError(err).Error("Could not start standupYesterdayRegister")
					}
					log.Info("Starting standupYesterdayRegister")

					err = s.standupToday(channelid, userid)
					if err != nil {
						log.WithError(err).Error("Could not start standupToday")
					}
					log.Info("Starting standupToday")
					break loop
				}
			}
		}
	}
	return nil
}

func (s *Slack) standupYesterdayRegister(response, timestamp, date, time, userid string) error {

	log.Info("Starting import in database of standupYesterday result")

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
		log.WithError(err).Error("Couldn't insert in the database the result of standupYesterday")
	}
	log.Info("standupYesterday result written in database")
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
		log.WithError(err).Error("Failed to post today standup question")
	}
	log.Info("Posting today standup question")

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			s.standupCancelTimeout(channelid)
			if err != nil {
				log.WithError(err).Error("Could not cancel standup")
			}
			log.Info("Standup Canceled")
			break loop
		case <-ticker.C:
			params2 := slack.HistoryParameters{
				Count:  1,
				Oldest: timestamp,
			}

			history, err := s.Client.GetIMHistory(channelid, params2)
			if err != nil {
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).WithError(err).Error("Could not get the IM history of the message")
			}
			log.WithFields(log.Fields{
				"timestamp": timestamp,
			}).Debug("Getting IM history of the message")

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
						log.WithError(err).Error("Could not cancel standup")
					}
					log.Info("Canceled standup")

					break loop
				default:
					err := s.standupTodayRegister(text, stamp, date, time, userid)
					if err != nil {
						log.WithError(err).Error("Could not start standupTodayRegister")
					}
					log.Info("Starting standupTodayRegister")

					err = s.standupBlocker(channelid, userid)
					if err != nil {
						log.WithError(err).Error("Could not start standupBlocker")
					}
					log.Info("Started standupBlocker")

					break loop
				}
			}
		}
	}
	return nil
}

func (s *Slack) standupTodayRegister(response, timestamp, date, time, userid string) error {

	log.Info("Starting import in database of standupToday result")

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
		log.WithError(err).Error("Couldn't insert in the database the result of standupToday")
	}
	log.Info("standupToday result written in database")

	return nil
}

func (s *Slack) standupBlocker(channelid, userid string) error {

	attachment := slack.Attachment{
		Text:       "Do you have any blockers?",
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
		log.WithError(err).Error("Failed to post blocker standup question")
	}
	log.Info("Posted blocker standup question")

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			s.standupCancelTimeout(channelid)
			if err != nil {
				log.WithError(err).Error("Could not cancel standup")
			}
			log.Info("Standup Canceled")

			break loop
		case <-ticker.C:
			params2 := slack.HistoryParameters{
				Count:  1,
				Oldest: timestamp,
			}

			history, err := s.Client.GetIMHistory(channelid, params2)
			if err != nil {
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).WithError(err).Error("Could not get the IM history of the message")
			}
			log.WithFields(log.Fields{
				"timestamp": timestamp,
			}).Debug("Getting IM history of the message")
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
						log.WithError(err).Error("Could not cancel standup")
					}
					log.Info("Standup canceled")

					break loop
				default:
					err := s.standupBlockerRegister(text, stamp, date, time, userid)
					if err != nil {
						log.WithError(err).Error("Could not start standupBlockerRegister")
					}
					log.Info("Started standupBlockerRegister")

					err = s.standupDone(channelid, userid, date)
					if err != nil {
						log.WithError(err).Error("Could not start standupDone")
					}
					log.Info("Started standupDone")

					break loop
				}
			}
		}
	}
	return nil
}

func (s *Slack) standupBlockerRegister(response, timestamp, date, time, userid string) error {

	log.Info("Starting import in database of standupBlocker result")

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
		log.WithError(err).Error("Couldn't insert in the database the result of standupBlocker")
	}
	log.Info("standupBlocker result written in database")

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
		log.WithError(err).Error("Failed to post standup canceled message")
	}
	log.Info("Posted standup canceled message")

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
		log.WithError(err).Error("Failed to post standup canceled timeout message")
	}
	log.Info("Posted standup canceled timeout message")

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
		log.WithError(err).Error("Failed to post standup done message")
	}
	log.Info("Posted standup done message")

	err = s.postStandupResults(userid, date)
	if err != nil {
		log.WithError(err).Error("Could not start postStandup")
	}
	log.Info("Started postStandup")

	return nil
}

func (s *Slack) postStandupResults(userid, date string) error {

	rows, err := database.DB.Query("SELECT userid, displayname, standup_channel FROM hatcher.users WHERE userid = $1;", userid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithError(err).Error("There is no results")
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
			log.WithError(err).Error("During the scan")
		}

		attachment := slack.Attachment{
			Pretext:    fmt.Sprintf("*%s* posted a daily standup note", displayname),
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
			Title:      "Do you have any blockers?",
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
			log.WithError(err).Error("Failed to post standup results")
		}
		log.Info("Standup posted")

	}
	return nil
}

func (s *Slack) standupResultsYesterday(userid, date, standupChannel string) (responseYesterday string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standupyesterday WHERE userid = $1 and date = $2;", userid, date)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithError(err).Error("There is no results")
		}
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&response)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}
	}
	return response
}

func (s *Slack) standupResultsToday(userid, date, standupChannel string) (responseToday string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standuptoday WHERE userid = $1 and date = $2;", userid, date)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithError(err).Error("There is no results")
		}
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&response)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}
	}
	return response
}

func (s *Slack) standupResultsBlocker(userid, date, standupChannel string) (responseBlocker string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standupblocker WHERE userid = $1 and date = $2;", userid, date)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithError(err).Error("There is no results")
		}
	}
	defer rows.Close()
	for rows.Next() {

		err = rows.Scan(&response)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}
	}
	return response
}
