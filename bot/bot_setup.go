package bot

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

func (s *Slack) askSetup(ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedSetup := map[string]bool{
		"setup": true,
		"init":  true,
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
			return err
		}
	}
	return nil
}

// initBot is the first step of using this bot.
// It will insert the user informations inside the databse to allow us
// to use them
func (s *Slack) initBot(userid, email, fullname, displayname string) {

	var id string

	sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit creates it in the database
	case sql.ErrNoRows:
		sqlWrite := `
		INSERT INTO hatcher.users (user_id, email, full_name, displayname)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
		if err = database.DB.QueryRow(sqlWrite, userid, email, fullname, displayname).Scan(&userid); err != nil {
			err = errors.Wrapf(err,
				"Couldn't create user %s with ID %s in the database.\n", fullname, userid)
			return
		}
		s.Logger.Printf("[DEBUG] User (%s) was created.\n", fullname)
	// If the user exist it will update it
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET full_name = $2, email = $3, displayname = $4 
		WHERE user_id = $1
		RETURNING id;`
		if err = database.DB.QueryRow(sqlUpdate, userid, fullname, email, displayname).Scan(&userid); err != nil {
			err = errors.Wrapf(err,
				"Couldn't update user %s with ID %s in the database.\n", fullname, userid)
			return
		}
		s.Logger.Printf("[DEBUG] User (%s) was updated.\n", fullname)
	default:
		panic(err)
	}
}

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
			return err
		}
	}
	return nil
}

// removeBot remove the user from the database
func (s *Slack) removeBot(userid, fullname string) {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s was not registered.\n", fullname)
	case nil:
		sqlDelete := `
		DELETE FROM hatcher.users
		WHERE user_id = $1;`
		if _, err = database.DB.Exec(sqlDelete, userid); err != nil {
			err = errors.Wrapf(err,
				"Couldn't not check if user %s with ID %s exist in the database.\n", fullname, userid)
			return
		}
		s.Logger.Printf("[DEBUG] User %s with id %s was deleted.\n", fullname, userid)
	default:
		panic(err)
	}
}

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
		return err
	}
	return nil
}

func (s *Slack) initManager(userid, fullname, managerid, managername string) {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit, exit
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	// If the user exist we update the column manager_id
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET manager_id = $2
		WHERE user_id = $1
		RETURNING id;`
		if err = database.DB.QueryRow(sqlUpdate, userid, managerid).Scan(&userid); err != nil {
			err = errors.Wrapf(err,
				"Couldn't update the manager %s for the user %s with ID %s.\n", managername, fullname, userid)
			return
		}
		s.Logger.Printf("[DEBUG] Manager %s was added to user %s.\n", managername, fullname)
	default:
		panic(err)
	}
}

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
		return err
	}
	return nil
}

func (s *Slack) setupIsManager(userid, fullname, ismanager string) {

	var id string

	// Check if the user already exist
	sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
	row := database.DB.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit, exit
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	// If the user exist we update the column is_manager
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET is_manager = $2
		WHERE user_id = $1
		RETURNING id;`
		if err = database.DB.QueryRow(sqlUpdate, userid, ismanager).Scan(&userid); err != nil {
			err = errors.Wrapf(err,
				"Couldn't update if user %s with ID %s is a manager in the database.\n", fullname, userid)
			return
		}
		if ismanager == "true" {
			s.Logger.Printf("[DEBUG] %s is now setup as a manager.\n", fullname)
		} else {
			s.Logger.Printf("[DEBUG] %s is not a manager.\n", fullname)
		}

	default:
		panic(err)
	}
}
