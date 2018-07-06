package bot

import (
	"database/sql"
	"fmt"
	"strings"

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
			s.Logger.Printf("[ERROR] Could not post askSetup question: %s\n", err)
		}
		s.Logger.Printf("[INFO] Message for askSetup posted.\n")
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
			s.Logger.Printf("[ERROR] Could not create user %s.\n %s", fullname, err)
		}
		s.Logger.Printf("[INFO] User (%s) was created.\n", fullname)
	// If the user exist it will update it
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET full_name = $2, email = $3, displayname = $4 
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, fullname, email, displayname).Scan(&userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Couldn't update user %s with ID %s in the database: %s\n", fullname, userid, err)
		}
		s.Logger.Printf("[INFO] User (%s) was updated.\n", fullname)
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
			s.Logger.Printf("[ERROR] Could not post message for askRemove: %s\n", err)
		}
		s.Logger.Printf("[INFO] Message for askRemove posted.\n")

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
		s.Logger.Printf("[DEBUG] User %s was not registered.\n", fullname)
	case nil:
		sqlDelete := `
		DELETE FROM hatcher.users
		WHERE userid = $1;`
		_, err = database.DB.Exec(sqlDelete, userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Couldn't not check if user %s with ID %s exist in the database: %s\n", fullname, userid, err)
		}
		s.Logger.Printf("[INFO] User %s with id %s was deleted.\n", fullname, userid)
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
		s.Logger.Print("[ERROR] Could not post message askWhoIsManager: %s\n", err)
	}
	s.Logger.Printf("[INFO] Message for askWhoIsManager posted.\n")
	return nil
}

// Add the person select previously in askWhoIsManager to the user profile
func (s *Slack) initManager(userid, fullname, managerid, managername string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit, exit
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	// If the user exist we update the column managerid
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET managerid = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, managerid).Scan(&userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Couldn't update the manager %s for the user %s with ID %s: %s\n", managername, fullname, userid, err)
		}
		s.Logger.Printf("[INFO] Manager %s was added to user %s.\n", managername, fullname)
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
		s.Logger.Printf("[ERROR] Could not post message for askIfManager: %s\n", err)
	}
	s.Logger.Printf("[INFO] Message for askIfManager posted.\n")
	return nil
}

// Setup the user as a manager or not in the database
func (s *Slack) setupIsManager(userid, fullname, ismanager string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit, exit
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	// If the user exist we update the column ismanager
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET ismanager = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, ismanager).Scan(&userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Couldn't update if user %s with ID %s is a manager in the database: %s\n", fullname, userid, err)
		}
		if ismanager == "true" {
			s.Logger.Printf("[DEBUG] %s is now setup as a manager.\n", fullname)
		}
		s.Logger.Printf("[INFO] %s is not a manager.\n", fullname)
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
		s.Logger.Printf("[ERROR] Could not post message askTimeHappinessSurvey: %s\n", err)
	}
	s.Logger.Printf("[INFO] Message askTimeHappinessSurvey posted.\n")
	return nil
}

// Insert in the database the result of askTimeHappinessSurvey
func (s *Slack) insertTimeHappinessSurvey(userid, fullname, time string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit, exit
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	// If the user exist we update the column ismanager
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET happiness_schedule = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, time).Scan(&id)
		if err != nil {
			s.Logger.Printf("[ERROR] Couldn't update the time of the happiness survey for user %s with ID %s: %s\n", fullname, userid, err)
		}
		s.Logger.Printf("[INFO] Time of the happiness survey for user %s with ID %s updated.\n", fullname, userid)

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
		s.Logger.Printf("[ERROR] Could not post message askTimeStandup: %s\n", err)
	}
	s.Logger.Printf("[INFO] Message askTimeStandup posted.\n")

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
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	// If the user exist we update the column ismanager
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET standup_schedule = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, time).Scan(&id)
		if err != nil {
			s.Logger.Printf("[ERROR] Couldn't update the time for standup for user %s with ID %s: %s\n", fullname, userid, err)
		}
		s.Logger.Printf("[INFO] Time for standup for user %s with ID %s updated.\n", fullname, userid)

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
		s.Logger.Print("[ERROR] Could not post message askWhoIsManager: %s\n", err)
	}
	s.Logger.Printf("[INFO] Message for askWhoIsManager posted.\n")

	return nil
}

// Insert in the database the result of askWhichChannelStandup
func (s *Slack) insertChannelStandup(userid, fullname, channel string) error {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT userid FROM hatcher.users WHERE userid=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET standup_channel = $2
		WHERE userid = $1
		RETURNING id;`
		err = database.DB.QueryRow(sqlUpdate, userid, channel).Scan(&id)
		if err != nil {
			s.Logger.Printf("[ERROR] Couldn't update the channel for the standup results for user %s with ID %s: %s\n", fullname, userid, err)
		}
		s.Logger.Printf("[INFO] Channel of the standup results for user %s with ID %s updated.\n", fullname, userid)

	default:
	}
	return nil
}
