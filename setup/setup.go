package setup

import (
	"database/sql"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/common"
	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
)

// Ask the first question on the user init process
// At this point the user can still cancel the stup
func AskSetup(s *common.Slack, ev *slack.MessageEvent) error {

	m := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
	if len(m) == 0 || m[0] != "hello" {
		n := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[:1]
		if len(n) == 0 || n[0] != "hello" {
			log.Debug("The message doesn't contain hello")
		}
	}

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "Do you want to setup/update your user with the bot Hatcher?",
		CallbackID: fmt.Sprintf("setup_%s", ev.User),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			{
				Name:  "SetupYes",
				Text:  "Yes",
				Type:  "button",
				Value: "SetupYes",
				Style: "primary",
			},
			{
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
		log.WithError(err).Error("Could not post askSetup question")
	}
	log.WithFields(log.Fields{
		"userid":  ev.User,
		"channel": ev.Channel,
	}).Info("Message for askSetup posted")

	return nil
}

// InitBot is the first step of using this bot.
// It will insert the user informations inside the database allowing us
// to use them
func InitBot(userid, email, fullname, displayname string) error {

	var id string

	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit creates it in the database
	case sql.ErrNoRows:
		sqlWrite := `
		INSERT INTO hatcher.users (userid, email, full_name, displayname)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
		err = database.DB.QueryRow(sqlWrite, userid, email, fullname, displayname).Scan(&userid)
		if err != nil {
			log.WithFields(log.Fields{
				"username": fullname,
				"userid":   userid,
			}).WithError(err).Error("Could not create user")
		}
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Info("User was created")

	// If the user exist it will update it
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET full_name = $2, email = $3, displayname = $4 
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, fullname, email, displayname).Scan(&userid)
		if err != nil {
			log.WithFields(log.Fields{
				"username": fullname,
				"userid":   userid,
			}).WithError(err).Error("Couldn't update user in the database")
		}
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Info("User was creaupdatedted")
	default:
	}
	return nil
}

// AskRemove Asks if we want to remove our user from the bot
// This will delete the user from the datatbase
func AskRemove(s *common.Slack, ev *slack.MessageEvent) error {
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
				{
					Name:  "RemoveYes",
					Text:  "Yes",
					Type:  "button",
					Value: "RemoveYes",
					Style: "primary",
				},
				{
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
			log.WithFields(log.Fields{
				"channel": ev.Channel,
				"userid":  ev.User,
			}).WithError(err).Error("Could not post message for askRemove")
		}
		log.WithFields(log.Fields{
			"channel": ev.Channel,
			"userid":  ev.User,
		}).Info("Message for askRemove posted")
	}
	return nil
}

// RemoveBot remove the user from the bot/database
func RemoveBot(userid, fullname string) error {

	sqlDelete := `
		DELETE FROM hatcher.users
		WHERE userid = $1;`
	_, err := database.DB.Exec(sqlDelete, userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid":   userid,
			"username": fullname,
		}).WithError(err).Error("Couldn't not check if user exist in the database")
	}
	log.WithFields(log.Fields{
		"username": fullname,
		"userid":   userid,
	}).Info("User was deleted")

	return nil
}

//AskRemoveHappiness Asks if we want to remove our user from the happiness survey
func AskRemoveHappiness(s *common.Slack, ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedRemoveHappiness := map[string]bool{
		"happiness remove": true,
	}

	if acceptedRemoveHappiness[text] {
		params := slack.PostMessageParameters{}
		attachment := slack.Attachment{
			Text:       "Do you want to remove your user form the happiness survey?",
			CallbackID: fmt.Sprintf("remove_%s", ev.User),
			Color:      "#FF0000",
			Actions: []slack.AttachmentAction{
				{
					Name:  "RemoveHapinnessYes",
					Text:  "Yes",
					Type:  "button",
					Value: "RemoveHapinnessYes",
					Style: "primary",
				},
				{
					Name:  "RemoveHappinessNo",
					Text:  "No",
					Type:  "button",
					Value: "RemoveHappinessNo",
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
			log.WithFields(log.Fields{
				"channel": ev.Channel,
				"userid":  ev.User,
			}).WithError(err).Error("Could not post message for askRemoveHappiness")
		}
		log.WithFields(log.Fields{
			"channel": ev.Channel,
			"userid":  ev.User,
		}).Info("Message for askRemoveHappiness posted")
	}
	return nil
}

// RemoveHappiness remove the user from the bot/database
func RemoveHappiness(userid, fullname string) error {

	sqlDelete := `
		UPDATE hatcher.users
		SET happiness_schedule = NULL
		WHERE userid = $1;`
	_, err := database.DB.Exec(sqlDelete, userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid":   userid,
			"username": fullname,
		}).WithError(err).Error("Couldn't not check if user exist in the database")
	}
	log.WithFields(log.Fields{
		"username": fullname,
		"userid":   userid,
	}).Info("Happiness survey time was removed")

	return nil
}

// AskWhoIsManager Ask who is the user manager
func AskWhoIsManager(s *common.Slack, channelid, userid string) error {

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
			{
				Name:  "NoManagerChosen",
				Text:  "No Manager",
				Type:  "button",
				Value: "NoManagerChosen",
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
		log.WithFields(log.Fields{
			"userid":  userid,
			"channel": channelid,
		}).WithError(err).Error("Could not post message askWhoIsManager")
	}
	log.WithFields(log.Fields{
		"userid":  userid,
		"channel": channelid,
	}).Info("Message for askWhoIsManager posted")
	return nil
}

// InitManager Add the person select previously in askWhoIsManager to the user profile
func InitManager(userid, fullname, managerid, managername string) error {

	sqlUpdate := `
		UPDATE hatcher.users
		SET managerid = $2
		WHERE userid = $1
		RETURNING id;`
	err := database.DB.QueryRow(sqlUpdate, userid, managerid).Scan(&userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid":    userid,
			"username":  fullname,
			"managerid": managerid,
		}).WithError(err).Error("Couldn't update the manager")
	}
	log.WithFields(log.Fields{
		"username":  fullname,
		"userid":    userid,
		"managerid": managerid,
	}).Info("Manager was added to user")
	return nil
}

// AskIfManager Asks if the user if a manager
func AskIfManager(s *common.Slack, channelid, userid string) error {

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "Do you manage a team with a daily standup?",
		CallbackID: fmt.Sprintf("ismanager_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			{
				Name:  "isManagerYes",
				Text:  "Yes",
				Type:  "button",
				Value: "isManagerYes",
				Style: "primary",
			},
			{
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
		log.WithFields(log.Fields{
			"userid":  userid,
			"channel": channelid,
		}).WithError(err).Error("Could not post message for askIfManager")
	}
	log.WithFields(log.Fields{
		"userid":  userid,
		"channel": channelid,
	}).Info("Message for askIfManager posted")
	return nil
}

// IsManager Setups the user as a manager or not in the database
func IsManager(userid, fullname, ismanager string) error {

	sqlUpdate := `
		UPDATE hatcher.users
		SET ismanager = $2
		WHERE userid = $1
		RETURNING id;`
	err := database.DB.QueryRow(sqlUpdate, userid, ismanager).Scan(&userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid":   userid,
			"username": fullname,
		}).WithError(err).Error("Couldn't update if user is a manager")
	}
	if ismanager == "true" {
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Info("Now setup as a manager")
	}
	log.WithFields(log.Fields{
		"username": fullname,
		"userid":   userid,
	}).Info("Is not setup as a manager")

	return nil
}

// AskSetupTimeHappinessSurvey Ask what time the happiness survey should be send
func AskSetupTimeHappinessSurvey(s *common.Slack, ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedHappinnessSurvey := map[string]bool{
		"happiness setup": true,
	}

	if acceptedHappinnessSurvey[text] {
		params := slack.PostMessageParameters{}
		attachment := slack.Attachment{
			Text:       "What time do you want the happiness survey to happen?",
			CallbackID: fmt.Sprintf("happTime_%s", ev.User),
			Color:      "#AED6F1",
			Actions: []slack.AttachmentAction{
				{
					Name: "SetupHappinessTime",
					Type: "select",
					Options: []slack.AttachmentActionOption{
						{
							Text:  "09:00",
							Value: "09:00",
						},
						{
							Text:  "09:15",
							Value: "09:15",
						},
						{
							Text:  "09:30",
							Value: "09:30",
						},
						{
							Text:  "09:45",
							Value: "09:45",
						},
						{
							Text:  "10:00",
							Value: "10:00",
						},
						{
							Text:  "10:15",
							Value: "10:15",
						},
						{
							Text:  "10:30",
							Value: "10:30",
						},
						{
							Text:  "10:45",
							Value: "10:45",
						},
						{
							Text:  "11:00",
							Value: "11:00",
						},
						{
							Text:  "11:15",
							Value: "11:15",
						},
						{
							Text:  "11:30",
							Value: "11:30",
						},
						{
							Text:  "11:45",
							Value: "11:45",
						},
					},
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
				"userid":  ev.User,
				"channel": ev.Channel,
			}).WithError(err).Error("Could not post message askTimeHappinessSurvey")
		}
		log.WithFields(log.Fields{
			"userid":  ev.User,
			"channel": ev.Channel,
		}).Info("Message askTimeHappinessSurvey posted")
	}
	return nil
}

// InsertTimeHappinessSurvey Inserts in the database the result of askTimeHappinessSurvey
func InsertTimeHappinessSurvey(userid, fullname, time string) error {

	sqlUpdate := `
		UPDATE hatcher.users
		SET happiness_schedule = $2
		WHERE userid = $1
		RETURNING id;`
	err := database.DB.QueryRow(sqlUpdate, userid, time).Scan(&userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid":   userid,
			"username": fullname,
		}).WithError(err).Error("Couldn't update the time of the happiness survey")
	}
	log.WithFields(log.Fields{
		"username": fullname,
		"userid":   userid,
	}).Info("Time of the happiness survey for user was updated")
	return nil
}

// AskTimeStandup Asks what time the standup should happen
func AskTimeStandup(s *common.Slack, channelid, userid string) error {

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "What time do you want your team standup to happen?",
		CallbackID: fmt.Sprintf("standupTime_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			{
				Name: "StandupTime",
				Type: "select",
				Options: []slack.AttachmentActionOption{
					{
						Text:  "09:00",
						Value: "09:00",
					},
					{
						Text:  "09:15",
						Value: "09:15",
					},
					{
						Text:  "09:30",
						Value: "09:30",
					},
					{
						Text:  "09:45",
						Value: "09:45",
					},
					{
						Text:  "10:00",
						Value: "10:00",
					},
					{
						Text:  "10:15",
						Value: "10:15",
					},
					{
						Text:  "10:30",
						Value: "10:30",
					},
					{
						Text:  "10:45",
						Value: "10:45",
					},
				},
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
			"userid":  userid,
			"channel": channelid,
		}).WithError(err).Error("Could not post message askTimeStandup")
	}
	log.WithFields(log.Fields{
		"userid":  userid,
		"channel": channelid,
	}).Info("Message askTimeStandup posted")
	return nil
}

// InsertTimeStandup Insert in the database the result of askTimeStandup
func InsertTimeStandup(userid, fullname, time string) error {

	sqlUpdate := `
		UPDATE hatcher.users
		SET standup_schedule = $2
		WHERE managerid = $1 OR userid = $1
		RETURNING id;`
	err := database.DB.QueryRow(sqlUpdate, userid, time).Scan(&userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid":   userid,
			"username": fullname,
		}).WithError(err).Error("Couldn't update the standup time")
	}
	log.WithFields(log.Fields{
		"username": fullname,
		"userid":   userid,
	}).Info("Standup time updated")

	return nil
}

// GetStandupTimeFromManager Gets a standup time based on the manager standup time selected
func GetStandupTimeFromManager(managerid string) (timeStandup string) {

	var time string

	rows, err := database.DB.Query("SELECT to_char(standup_schedule, 'HH24:MI') FROM hatcher.users WHERE userid = $1", managerid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithError(err).Error("There is no results")
		}
	}
	defer rows.Close()
	for rows.Next() {

		err = rows.Scan(&time)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}
	}
	return time
}

// GetStandupChannelFromManager Gets a standup channel id based on the manager standup channel id selected
func GetStandupChannelFromManager(managerid string) (channelStandup string) {

	var channel string

	rows, err := database.DB.Query("SELECT standup_channel FROM hatcher.users WHERE userid = $1", managerid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithError(err).Error("There is no results")
		}
	}
	defer rows.Close()
	for rows.Next() {

		err = rows.Scan(&channel)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}
	}
	return channel
}

// UpdateTimeStandup Updates the standup time of a new user based on the time of the manager
func UpdateTimeStandup(managerid, userid, time string) error {

	var id string
	if time == "NULL" {
		sqlUpdate := `
		UPDATE hatcher.users
		SET standup_schedule = NULL
		WHERE managerid = $1
		RETURNING id;`
		err := database.DB.QueryRow(sqlUpdate, managerid).Scan(&id)
		if err != nil {
			log.WithFields(log.Fields{
				"managerid": managerid,
				"userid":    userid,
			}).WithError(err).Error("Couldn't update the standup time")
		}
		log.WithFields(log.Fields{
			"userid":    userid,
			"managerid": managerid,
		}).Info("Standup time updated")
	} else {
		sqlUpdate := `
		UPDATE hatcher.users
		SET standup_schedule = $2
		WHERE managerid = $1
		RETURNING id;`
		err := database.DB.QueryRow(sqlUpdate, managerid, time).Scan(&id)
		if err != nil {
			log.WithFields(log.Fields{
				"managerid": managerid,
				"userid":    userid,
			}).WithError(err).Error("Couldn't update the standup time")
		}
		log.WithFields(log.Fields{
			"userid":    userid,
			"managerid": managerid,
		}).Info("Standup time updated")
	}

	return nil
}

// UpdateChannelStandup Updates the standup channel of a new user based on the channel of the manager
func UpdateChannelStandup(managerid, userid, channel string) error {

	sqlUpdate := `
		UPDATE hatcher.users
		SET standup_channel = $2
		WHERE managerid = $1
		RETURNING id;`
	err := database.DB.QueryRow(sqlUpdate, managerid, channel).Scan(&managerid)
	if err != nil {
		log.WithFields(log.Fields{
			"managerid": managerid,
			"userid":    userid,
		}).WithError(err).Error("Couldn't update the standup time")
	}
	log.WithFields(log.Fields{
		"userid":    userid,
		"managerid": managerid,
	}).Info("Standup time updated")
	return nil
}

// AskWhichChannelStandup Asks in which channel to post the standup results
func AskWhichChannelStandup(s *common.Slack, channelid, userid string) error {

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "In which channel do you want to post your standup results?",
		CallbackID: fmt.Sprintf("manager_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			{
				Name:       "ChannelStandupChosen",
				Text:       "Type to filter option",
				Type:       "select",
				DataSource: "channels",
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
			"userid":  userid,
			"channel": channelid,
		}).WithError(err).Error("Could not post message askWhichChannelStandup")
	}
	log.WithFields(log.Fields{
		"userid":  userid,
		"channel": channelid,
	}).Info("Message for askWhichChannelStandup posted")
	return nil
}

// InsertChannelStandup Inserts in the database the result of askWhichChannelStandup
func InsertChannelStandup(userid, fullname, channel string) error {

	sqlUpdate := `
		UPDATE hatcher.users
		SET standup_channel = $2
		WHERE userid = $1 OR managerid = $1
		RETURNING id;`
	err := database.DB.QueryRow(sqlUpdate, userid, channel).Scan(&userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid":   userid,
			"username": fullname,
		}).WithError(err).Error("Couldn't update the channel for the standup results")
	}
	log.WithFields(log.Fields{
		"username": fullname,
		"userid":   userid,
	}).Info("Channel for standup results updated")
	return nil
}
