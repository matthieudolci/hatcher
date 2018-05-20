package bot

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
	dbname   = "hatcher"
)

func (s *Slack) askSetup(ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedHowAreYou := map[string]bool{
		"setup": true,
		"init":  true,
	}

	if acceptedHowAreYou[text] {
		params := slack.PostMessageParameters{}
		attachment := slack.Attachment{
			Text:       "Do you want to setup/update your user with the bot Hatcher?",
			CallbackID: fmt.Sprintf("ask_%s", ev.User),
			Color:      "#AED6F1",
			Actions: []slack.AttachmentAction{
				slack.AttachmentAction{
					Name:  "action",
					Text:  "Yes",
					Type:  "button",
					Value: "yes",
				},
				slack.AttachmentAction{
					Name:  "action",
					Text:  "No",
					Type:  "button",
					Value: "no",
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
// to use the informations
func initBot(ev *slack.MessageEvent, api *slack.Client) {
	var id string
	getuserid := ev.Msg.User
	userid := fmt.Sprintf(getuserid)
	user, err := api.GetUserInfo(userid)
	fullname := fmt.Sprintf(user.Profile.RealName)
	email := fmt.Sprintf(user.Profile.Email)
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	acceptedSetup := map[string]bool{
		"setup": true,
		"init":  true,
	}

	if acceptedSetup[text] {
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		// Check if the user already exist
		sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
		row := db.QueryRow(sqlCheckID, userid)
		switch err := row.Scan(&id); err {
		// if user doesnt exit creates it in the database
		case sql.ErrNoRows:
			sqlWrite := `
			INSERT INTO hatcher.users (user_id, email, full_name)
			VALUES ($1, $2, $3)
			RETURNING id`
			err = db.QueryRow(sqlWrite, userid, email, fullname).Scan(&userid)
			if err != nil {
				panic(err)
			}
		// If the user exist if it update it
		case nil:
			sqlUpdate := `
			UPDATE hatcher.users
			SET full_name = $2, email = $3
			WHERE user_id = $1
			RETURNING id;`
			err = db.QueryRow(sqlUpdate, userid, fullname, email).Scan(&userid)
			if err != nil {
				panic(err)
			}
		default:
			panic(err)
		}

	}
}

// // removeBot remove the user from the database
// func (s *Slack) removeBot(ev *slack.MessageEvent, userid string) {
// 	var responseRegistered string
// 	var responseRowsExist string
// 	var id string

// 	text := ev.Text
// 	text = strings.TrimSpace(text)
// 	text = strings.ToLower(text)

// 	acceptedSetup := map[string]bool{
// 		"remove": true,
// 		"delete": true,
// 	}

// 	if acceptedSetup[text] {
// 		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
// 			"password=%s dbname=%s sslmode=disable",
// 			host, port, user, password, dbname)
// 		db, err := sql.Open("postgres", psqlInfo)
// 		if err != nil {
// 			panic(err)
// 		}
// 		defer db.Close()
// 		// Check if the user already exist
// 		sqlCheckId := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
// 		row := db.QueryRow(sqlCheckId, userid)
// 		switch err := row.Scan(&id); err {
// 		case sql.ErrNoRows:
// 			responseRegistered = "Your user was not registered!"
// 			rtm.SendMessage(rtm.NewOutgoingMessage(responseRegistered, msg.Channel))
// 		// Delete the user
// 		case nil:
// 			sqlDelete := `
// 			DELETE FROM hatcher.users
// 			WHERE user_id = $1;`
// 			_, err = db.Exec(sqlDelete, userid)
// 			if err != nil {
// 				panic(err)
// 			}
// 			fmt.Println(id)
// 			responseRowsExist = "Your user was deleted!"
// 			rtm.SendMessage(rtm.NewOutgoingMessage(responseRowsExist, msg.Channel))
// 		default:
// 			panic(err)
// 		}

// 	}
// }
