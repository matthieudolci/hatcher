package bot

import (
	"database/sql"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
)

// Ask the first question on the user init process
// At this point the user can still cancel the stup
func (s *Slack) askSetup(ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedSetup := map[string]bool{
		"hello": true,
		"hi":    true,
		"setup": true,
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
			log.WithError(err).Error("Could not post askSetup question")
		}
		log.WithFields(log.Fields{
			"userid":  ev.User,
			"channel": ev.Channel,
		}).Info("Message for askSetup posted")
	}
	return nil
}

// initBot is the first step of using this bot.
// It will insert the user informations inside the database allowing us
// to use them
func (s *Slack) initBot(userid, email, fullname, displayname string) error {

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

// Ask if we want to remove our user from the bot
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

// removeBot remove the user from the bot/database
func (s *Slack) removeBot(userid, fullname string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Debug("User is not registered")
	case nil:
		sqlDelete := `
		DELETE FROM hatcher.users
		WHERE userid = $1;`
		_, err = database.DB.Exec(sqlDelete, userid)
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
	default:
	}
	return nil
}

// Ask who is the user manager
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

// Add the person select previously in askWhoIsManager to the user profile
func (s *Slack) initManager(userid, fullname, managerid, managername string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Debug("User is not registered")
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET managerid = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, managerid).Scan(&userid)
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
	default:
	}
	return nil
}

// Ask if the user if a manager
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

// Setup the user as a manager or not in the database
func (s *Slack) setupIsManager(userid, fullname, ismanager string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Debug("User is not registered")
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET ismanager = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, ismanager).Scan(&userid)
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
	default:
	}
	return nil
}

// Ask what time the happiness survey should be send
func (s *Slack) askTimeHappinessSurvey(channelid, userid string) error {

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "What time do you want the happiness survey to happen?",
		CallbackID: fmt.Sprintf("happTime_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			{
				Name: "HappinessTime",
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
		}).WithError(err).Error("Could not post message askTimeHappinessSurvey")
	}
	log.WithFields(log.Fields{
		"userid":  userid,
		"channel": channelid,
	}).Info("Message askTimeHappinessSurvey posted")
	return nil
}

// Insert in the database the result of askTimeHappinessSurvey
func (s *Slack) insertTimeHappinessSurvey(userid, fullname, time string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Debug("User is not registered")
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET happiness_schedule = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, time).Scan(&id)
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
	default:
	}
	return nil
}

// Ask what time the standup should happen
func (s *Slack) askTimeStandup(channelid, userid string) error {

	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "What time do you want your standup to happen?",
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

// Insert in the database the result of askTimeStandup
func (s *Slack) insertTimeStandup(userid, fullname, time string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit, exit
	case sql.ErrNoRows:
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Debug("User is not registered")
	// If the user exist we update the column ismanager
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET standup_schedule = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, time).Scan(&id)
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
	default:
	}
	return nil
}

// Ask in which channel to post the standup results
func (s *Slack) askWhichChannelStandup(channelid, userid string) error {

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

// Insert in the database the result of askWhichChannelStandup
func (s *Slack) insertChannelStandup(userid, fullname, channel string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		log.WithFields(log.Fields{
			"username": fullname,
			"userid":   userid,
		}).Debug("User is not registered")
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET standup_channel = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, channel).Scan(&id)
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
	default:
	}
	return nil
}
