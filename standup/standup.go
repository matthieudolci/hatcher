package standup

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/common"
	"github.com/matthieudolci/hatcher/database"
	"github.com/slack-go/slack"
)

func AskStandupYesterday(s *common.Slack, ev *slack.MessageEvent) error {

	t := time.Now().Local().Format("2006-01-02")
	t2 := time.Now().Local().Format("15:04:05")
	date := fmt.Sprint(t)
	times := fmt.Sprint(t2)

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)

	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedStandup := map[string]bool{
		"standup": true,
	}

	if acceptedStandup[text] {

		uuid := common.CreatesUUID()
		log.WithFields(log.Fields{
			"uuid": uuid,
		}).Info("Standup uuid generated")

		attachment := slack.Attachment{
			Text:       "What did you do yesterday?",
			Color:      "#2896b7",
			CallbackID: fmt.Sprintf("Yesterday_%s", ev.User),
		}

		params := slack.MsgOptionAttachments(attachment)

		_, timestamp, err := s.Client.PostMessage(ev.Channel, params)
		if err != nil {
			log.WithError(err).Error("Failed to post yesterday standup question")
		}
		log.WithFields(log.Fields{
			"userid":    ev.User,
			"timestamp": timestamp,
		}).Info("Timestamp of the Standup Yesterday message")

		go postStandupResults(s, ev.User, date, times, uuid)

	loop:
		for {
			select {
			case <-timer.C:
				err = postStandupCancelTimeout(s, ev.Channel)
				if err != nil {
					log.WithError(err).Error("Could not cancel standup")
				}
				log.Info("Standup Canceled")
				break loop
			case <-ticker.C:
				params2 := &slack.GetConversationHistoryParameters{
					Limit:     1,
					Oldest:    timestamp,
					ChannelID: ev.Channel,
				}

				history, err := s.Client.GetConversationHistory(params2)
				if err != nil {
					log.WithFields(log.Fields{
						"timestamp": timestamp,
					}).WithError(err).Error("Could not get the IM history of the message")
					continue
				}
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).Debug("Getting IM history of the message")

				message := history.Messages

				if len(message) == 0 {

				}
				if len(message) > 0 {
					text := history.Messages[0].Msg.Text
					userid := history.Messages[0].Msg.User
					stamp := history.Messages[0].Msg.Timestamp
					switch text {
					case "cancel":
						err = postStandupCancel(s, ev.Channel)
						if err != nil {
							log.WithError(err).Error("Could not cancel standup")
						}
						log.Info("Standup Canceled")

						break loop
					default:
						err = yesterdayRegister(text, stamp, date, times, userid, uuid)
						if err != nil {
							log.WithError(err).Error("Could not start yesterdayRegister")
						}
						log.Info("Starting yesterdayRegister")

						err = askStandupToday(s, ev.Channel, ev.User, date, times, uuid)
						if err != nil {
							log.WithError(err).Error("Could not start askStandupToday")
						}
						log.Info("Starting askStandupToday")

						break loop
					}
				}
			}
		}
	}
	return nil
}

func AskStandupYesterdayScheduled(s *common.Slack, userid string) error {

	t := time.Now().Local().Format("2006-01-02")
	t2 := time.Now().Local().Format("15:04:05")
	date := fmt.Sprint(t)
	times := fmt.Sprint(t2)

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)

	uuid := common.CreatesUUID()
	log.WithFields(log.Fields{
		"uuid": uuid,
	}).Info("Standup uuid generated")

	params := &slack.OpenConversationParameters{
		Users: []string{userid},
	}

	channelid, _, _, err := s.Client.OpenConversation(params)
	if err != nil {
		log.WithError(err).Error("OpenIMChannel could not get channel id")
	}

	attachment := slack.Attachment{
		Text:       "What did you do yesterday?",
		Color:      "#2896b7",
		CallbackID: fmt.Sprintf("standupYesterday_%s", userid),
	}

	params2 := slack.MsgOptionAttachments(attachment)
	_, timestamp, err := s.Client.PostMessage(channelid.ID, params2)
	if err != nil {
		log.WithError(err).Error("Failed to post yesterday standup question")
	}
	log.WithFields(log.Fields{
		"timestamp": timestamp,
	}).Info("Timestamp of the standupYesterday message")

	go postStandupResults(s, userid, date, times, uuid)

loop:
	for {
		select {
		case <-timer.C:
			err = postStandupCancelTimeout(s, channelid.ID)
			if err != nil {
				log.WithError(err).Error("Could not cancel standup")
			}
			log.Info("Standup Canceled")

			break loop
		case <-ticker.C:
			params2 := &slack.GetConversationHistoryParameters{
				Limit:     1,
				Oldest:    timestamp,
				ChannelID: channelid.ID,
			}

			history, err := s.Client.GetConversationHistory(params2)
			if err != nil {
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).WithError(err).Error("Could not get the IM history of the message")
				continue
			}
			log.WithFields(log.Fields{
				"timestamp": timestamp,
			}).Debug("Getting IM history of the message with timestamp")

			message := history.Messages

			if len(message) == 0 {

			}
			if len(message) > 0 {
				text := history.Messages[0].Msg.Text
				userid := history.Messages[0].Msg.User
				stamp := history.Messages[0].Msg.Timestamp
				switch text {
				case "cancel":
					err = postStandupCancel(s, channelid.ID)
					if err != nil {
						log.WithError(err).Error("Could not cancel standup")
					}
					log.Info("Standup Canceled")

					break loop
				default:
					err = yesterdayRegister(text, stamp, date, times, userid, uuid)
					if err != nil {
						log.WithError(err).Error("Could not start yesterdayRegister")
					}
					log.Info("Starting yesterdayRegister")

					err = askStandupToday(s, channelid.ID, userid, date, times, uuid)
					if err != nil {
						log.WithError(err).Error("Could not start askStandupToday")
					}
					log.Info("Starting askStandupToday")
					break loop
				}
			}
		}
	}
	return nil
}

func yesterdayRegister(response, timestamp, date, time, userid, uuid string) error {

	log.Info("Starting import in database of standupYesterday result")

	var id string

	sqlWrite := `
	INSERT INTO 
		hatcher.standupyesterday 
		(response, timestamp, date, time, userid, uuid)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	err := database.DB.QueryRow(
		sqlWrite,
		response,
		timestamp,
		date,
		time,
		userid,
		uuid).Scan(&id)
	if err != nil {
		log.WithError(err).Error("Couldn't insert in the database the result of AskStandupYesterday")
	}
	log.Info("AskStandupYesterday result written in database")
	return nil
}

func askStandupToday(s *common.Slack, channelid, userid, date, times, uuid string) error {

	attachment := slack.Attachment{
		Text:       "What are you doing today?",
		Color:      "#41aa3f",
		CallbackID: "standupToday",
	}

	params := slack.MsgOptionAttachments(attachment)
	_, timestamp, err := s.Client.PostMessage(channelid, params)
	if err != nil {
		log.WithError(err).Error("Failed to post today standup question")
	}
	log.Info("Posting today standup question")

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			err = postStandupCancelTimeout(s, channelid)
			if err != nil {
				log.WithError(err).Error("Could not cancel standup")
			}
			log.Info("Standup Canceled")
			break loop
		case <-ticker.C:
			params2 := &slack.GetConversationHistoryParameters{
				Limit:     1,
				Oldest:    timestamp,
				ChannelID: channelid,
			}

			history, err := s.Client.GetConversationHistory(params2)
			if err != nil {
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).WithError(err).Error("Could not get the IM history of the message")
				continue
			}
			log.WithFields(log.Fields{
				"timestamp": timestamp,
			}).Debug("Getting IM history of the message")

			message := history.Messages

			if len(message) == 0 {

			}
			if len(message) > 0 {
				text := history.Messages[0].Msg.Text
				userid := history.Messages[0].Msg.User
				stamp := history.Messages[0].Msg.Timestamp
				switch text {
				case "cancel":
					err = postStandupCancel(s, channelid)
					if err != nil {
						log.WithError(err).Error("Could not cancel standup")
					}
					log.Info("Canceled standup")

					break loop
				default:
					err := TodayRegister(text, stamp, date, times, userid, uuid)
					if err != nil {
						log.WithError(err).Error("Could not start TodayRegister")
					}
					log.Info("Starting TodayRegister")

					err = askStandupBlocker(s, channelid, userid, date, times, uuid)
					if err != nil {
						log.WithError(err).Error("Could not start askStandupBlocker")
					}
					log.Info("Started askStandupBlocker")

					break loop
				}
			}
		}
	}
	return nil
}

func TodayRegister(response, timestamp, date, time, userid, uuid string) error {

	log.Info("Starting import in database of AskStandupToday result")

	var id string

	sqlWrite := `
	INSERT INTO 
		hatcher.standuptoday 
		(response, timestamp, date, time, userid, uuid)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	err := database.DB.QueryRow(
		sqlWrite,
		response,
		timestamp,
		date,
		time,
		userid,
		uuid).Scan(&id)
	if err != nil {
		log.WithError(err).Error("Couldn't insert in the database the result of AskStandupToday")
	}
	log.Info("AskStandupToday result written in database")

	return nil
}

func askStandupBlocker(s *common.Slack, channelid, userid, date, times, uuid string) error {

	attachment := slack.Attachment{
		Text:       "Do you have any blockers?",
		Color:      "#f91b1b",
		CallbackID: "standupBlocker",
	}

	params := slack.MsgOptionAttachments(attachment)

	_, timestamp, err := s.Client.PostMessage(channelid, params)
	if err != nil {
		log.WithError(err).Error("Failed to post blocker standup question")
	}
	log.Info("Posted blocker standup question")

	timer := time.NewTimer(10 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)

loop:
	for {
		select {
		case <-timer.C:
			err = postStandupCancelTimeout(s, channelid)
			if err != nil {
				log.WithError(err).Error("Could not cancel standup")
			}
			log.Info("Standup Canceled")

			break loop
		case <-ticker.C:
			params2 := &slack.GetConversationHistoryParameters{
				Limit:     1,
				Oldest:    timestamp,
				ChannelID: channelid,
			}

			history, err := s.Client.GetConversationHistory(params2)
			if err != nil {
				log.WithFields(log.Fields{
					"timestamp": timestamp,
				}).WithError(err).Error("Could not get the IM history of the message")
				continue
			}
			log.WithFields(log.Fields{
				"timestamp": timestamp,
			}).Debug("Getting IM history of the message")

			message := history.Messages

			if len(message) == 0 {

			}
			if len(message) > 0 {
				text := history.Messages[0].Msg.Text
				userid := history.Messages[0].Msg.User
				stamp := history.Messages[0].Msg.Timestamp
				switch text {
				case "cancel":
					err := postStandupCancel(s, channelid)
					if err != nil {
						log.WithError(err).Error("Could not cancel standup")
					}
					log.Info("Standup canceled")

					break loop
				default:
					err := BlockerRegister(text, stamp, date, times, userid, uuid)
					if err != nil {
						log.WithError(err).Error("Could not start BlockerRegister")
					}
					log.Info("Started BlockerRegister")

					err = postStandupDone(s, channelid, userid, date, times, uuid)
					if err != nil {
						log.WithError(err).Error("Could not start postStandupDone")
					}
					log.Info("Started postStandupDone")

					break loop
				}
			}
		}
	}
	return nil
}

func BlockerRegister(response, timestamp, date, time, userid, uuid string) error {

	log.Info("Starting import in database of AskStandupBlocker result")

	var id string

	sqlWrite := `
	INSERT INTO 
		hatcher.standupblocker
		(response, timestamp, date, time, userid, uuid)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	err := database.DB.QueryRow(
		sqlWrite,
		response,
		timestamp,
		date,
		time,
		userid,
		uuid).Scan(&id)
	if err != nil {
		log.WithError(err).Error("Couldn't insert in the database the result of AskStandupBlocker")
	}
	log.Info("AskStandupBlocker result written in database")

	return nil
}

func postStandupCancel(s *common.Slack, channelid string) error {

	attachment := slack.Attachment{
		Text:       "Standup canceled!",
		Color:      "#f91b1b",
		CallbackID: "standupCancel",
	}

	params := slack.MsgOptionAttachments(attachment)

	_, _, err := s.Client.PostMessage(channelid, params)
	if err != nil {
		log.WithError(err).Error("Failed to post standup canceled message")
	}
	log.Info("Posted standup canceled message")

	return nil
}

func postStandupCancelTimeout(s *common.Slack, channelid string) error {

	attachment := slack.Attachment{
		Text:       "The standup was canceled for timeout.\nYou can start a new standup by sending `standup`",
		Color:      "#f91b1b",
		CallbackID: "standupCancelTimeout",
	}

	params := slack.MsgOptionAttachments(attachment)

	_, _, err := s.Client.PostMessage(channelid, params)
	if err != nil {
		log.WithError(err).Error("Failed to post standup canceled timeout message")
	}
	log.Info("Posted standup canceled timeout message")

	return nil
}

func postStandupDone(s *common.Slack, channelid, userid, date, time, uuid string) error {

	attachment := slack.Attachment{
		Text:       "Standup Done! Your results will be posted in 10 mins\nEditing your answers will edit the results in the standup channel.\nThanks and see you tomorrow :smiley:",
		Color:      "#2896b7",
		CallbackID: "standupDone",
	}

	params := slack.MsgOptionAttachments(attachment)

	_, _, err := s.Client.PostMessage(channelid, params)
	if err != nil {
		log.WithError(err).Error("Failed to post standup done message")
	}
	log.Info("Posted standup done message")

	return nil
}

func getResultsYesterday(userid, date, standupChannel, uuid string) (responseYesterday string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standupyesterday WHERE userid = $1 and date = $2 and uuid = $3;", userid, date, uuid)
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

func getResultsToday(userid, date, standupChannel, uuid string) (responseToday string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standuptoday WHERE userid = $1 and date = $2 and uuid = $3;", userid, date, uuid)
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

func getResultsBlocker(userid, date, standupChannel, uuid string) (responseBlocker string) {

	var response string

	rows, err := database.DB.Query("SELECT response FROM hatcher.standupblocker WHERE userid = $1 and date = $2 and uuid = $3;", userid, date, uuid)
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

func CheckIfYesterdayMessageEdited(s *common.Slack, ev *slack.MessageEvent) error {

	timestamp := ev.SubMessage.Timestamp
	text := ev.SubMessage.Text
	userid := ev.SubMessage.User

	if common.RowExists("SELECT exists (SELECT timestamp FROM hatcher.standupyesterday WHERE timestamp=$1)", timestamp) {
		err := common.QueryRow("UPDATE hatcher.standupyesterday SET response=$2 WHERE timestamp=$1", timestamp, text)
		if err != nil {
			log.WithError(err).Error("Could not edit the yesterday standup row")
		}
		log.WithFields(log.Fields{
			"timestamp": timestamp,
		}).Info("The yesterday standup row was edited")

		var standupUUID string

		standupUUID, err = common.QueryUUID("SELECT uuid FROM hatcher.standupyesterday WHERE timestamp=$1", timestamp)
		if err != nil {
			log.WithError(err).Error("Could get the standup uuid")
		}
		log.WithFields(log.Fields{
			"uuid":      standupUUID,
			"timestamp": timestamp,
		}).Info("The standup uuid was retrieve")

		err = updateStandupResults(s, userid, standupUUID)
		if err != nil {
			log.WithError(err).Error("Could not start the standup update for yesterday notes")
		}

	} else {
		log.Println("The row doesn't exist")
	}
	return nil
}

func CheckIfTodayMessageEdited(s *common.Slack, ev *slack.MessageEvent) error {

	timestamp := ev.SubMessage.Timestamp
	text := ev.SubMessage.Text
	userid := ev.SubMessage.User

	if common.RowExists("SELECT exists (SELECT timestamp FROM hatcher.standuptoday WHERE timestamp=$1)", timestamp) {
		err := common.QueryRow("UPDATE hatcher.standuptoday SET response=$2 WHERE timestamp=$1", timestamp, text)
		if err != nil {
			log.WithError(err).Error("Could not edit the today standup row")
		}
		log.WithFields(log.Fields{
			"timestamp": timestamp,
		}).Info("The today standup row was edited")

		var standupUUID string

		standupUUID, err = common.QueryUUID("SELECT uuid FROM hatcher.standuptoday WHERE timestamp=$1", timestamp)
		if err != nil {
			log.WithError(err).Error("Could get the standup uuid")
		}
		log.WithFields(log.Fields{
			"uuid":      standupUUID,
			"timestamp": timestamp,
		}).Info("The standup uuid was retrieve")

		err = updateStandupResults(s, userid, standupUUID)
		if err != nil {
			log.WithError(err).Error("Could not start the standup update for Today notes")
		}

	} else {
		log.Println("The row doesn't exist")
	}
	return nil
}

func CheckIfBlockerMessageEdited(s *common.Slack, ev *slack.MessageEvent) error {

	timestamp := ev.SubMessage.Timestamp
	text := ev.SubMessage.Text
	userid := ev.SubMessage.User

	if common.RowExists("SELECT exists (SELECT timestamp FROM hatcher.standupblocker WHERE timestamp=$1)", timestamp) {
		err := common.QueryRow("UPDATE hatcher.standupblocker SET response=$2 WHERE timestamp=$1", timestamp, text)
		if err != nil {
			log.WithError(err).Error("Could not edit the blocker standup row")
		}
		log.WithFields(log.Fields{
			"timestamp": timestamp,
		}).Info("The blocker standup row was edited")

		var standupUUID string

		standupUUID, err = common.QueryUUID("SELECT uuid FROM hatcher.standupblocker WHERE timestamp=$1", timestamp)
		if err != nil {
			log.WithError(err).Error("Could get the standup uuid")
		}
		log.WithFields(log.Fields{
			"uuid":      standupUUID,
			"timestamp": timestamp,
		}).Info("The standup uuid was retrieve")

		err = updateStandupResults(s, userid, standupUUID)
		if err != nil {
			log.WithError(err).Error("Could not start the standup update for blocker")
		}

	} else {
		log.Println("The row doesn't exist")
	}
	return nil
}

func postStandupResults(s *common.Slack, userid, date, times, uuid string) {

	timer := time.NewTimer(10 * time.Minute)

	<-timer.C

	if common.RowExists("SELECT exists (SELECT * FROM hatcher.standupblocker WHERE uuid=$1)", uuid) {
		rows, err := database.DB.Query("SELECT displayname, standup_channel FROM hatcher.users WHERE userid = $1;", userid)
		if err != nil {
			if err == sql.ErrNoRows {
				log.WithError(err).Error("There is no results")
			}
		}
		defer rows.Close()
		for rows.Next() {

			var displayname string
			var standupChannel string

			responseYesterday := getResultsYesterday(userid, date, standupChannel, uuid)
			responseToday := getResultsToday(userid, date, standupChannel, uuid)
			responseBlocker := getResultsBlocker(userid, date, standupChannel, uuid)

			err = rows.Scan(&displayname, &standupChannel)
			if err != nil {
				log.WithError(err).Error("During the scan")
			}

			attachment := slack.Attachment{
				Title:      "What did you do yesterday?",
				Text:       responseYesterday,
				Color:      "#2896b7",
				CallbackID: fmt.Sprintf("resultsStandupYesterday_%s", userid),
			}

			attachment2 := slack.Attachment{
				Title:      "What are you doing today?",
				Color:      "#41aa3f",
				Text:       responseToday,
				CallbackID: fmt.Sprintf("resultsStandupToday_%s", userid),
			}

			attachment3 := slack.Attachment{
				Title:      "Do you have any blockers?",
				Color:      "#f91b1b",
				Text:       responseBlocker,
				CallbackID: fmt.Sprintf("resultsStandupBlocker_%s", userid),
			}

			params := slack.MsgOptionAttachments(attachment, attachment2, attachment3)

			text := fmt.Sprintf("%s posted a daily standup note", displayname)
			_, respTimestamp, err := s.Client.PostMessage(
				standupChannel,
				slack.MsgOptionText(text, false),
				params)
			if err != nil {
				log.WithError(err).Error("Failed to post standup results")
			}
			log.WithFields(log.Fields{
				"timestamp": respTimestamp,
			}).Info("Standup results posted")

			err = common.QueryRow("INSERT INTO hatcher.standupresults (timestamp, date, time, uuid) VALUES ($1, $2, $3, $4)", respTimestamp, date, times, uuid)
			if err != nil {
				log.WithError(err).Error("Could not edit the standup result timestamp row")
			}
			log.WithFields(log.Fields{
				"timestamp": respTimestamp,
			}).Info("The standup result timestamp row was edited")
		}
	}
}

func updateStandupResults(s *common.Slack, userid, uuid string) error {

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
		var timestamp string
		var date string

		err := database.DB.QueryRow("SELECT timestamp, date FROM hatcher.standupresults WHERE uuid=$1", uuid).Scan(&timestamp, &date)
		if err != nil && err != sql.ErrNoRows {
			log.WithError(err).Error("Impossible to get the standup result uuid")
		}

		responseYesterday := getResultsYesterday(userid, date, standupChannel, uuid)
		responseToday := getResultsToday(userid, date, standupChannel, uuid)
		responseBlocker := getResultsBlocker(userid, date, standupChannel, uuid)

		err = rows.Scan(&userid, &displayname, &standupChannel)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}

		attachment := slack.Attachment{
			Title:      "What did you do yesterday?",
			Color:      "#2896b7",
			Text:       responseYesterday,
			CallbackID: fmt.Sprintf("resultsStandupYesterday_%s", userid),
		}

		attachment2 := slack.Attachment{
			Title:      "What are you doing today?",
			Color:      "#41aa3f",
			Text:       responseToday,
			CallbackID: fmt.Sprintf("resultsStandupToday_%s", userid),
		}

		attachment3 := slack.Attachment{
			Title:      "Do you have any blockers?",
			Color:      "#f91b1b",
			Text:       responseBlocker,
			CallbackID: fmt.Sprintf("resultsStandupBlocker_%s", userid),
		}

		text := fmt.Sprintf("%s posted a daily standup note", displayname)
		_, _, _, err = s.Client.SendMessage(
			standupChannel,
			slack.MsgOptionText(text, false),
			slack.MsgOptionUpdate(timestamp),
			slack.MsgOptionAttachments(
				attachment,
				attachment2,
				attachment3,
			),
		)
		if err != nil {
			log.WithError(err).Error("Failed to post updated standup results")
		}
		log.Info("Updated standup results posted")
	}
	return nil
}
