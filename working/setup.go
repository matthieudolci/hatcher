package main

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

// initBot is the first step of using this bot.
// It will insert the user informations inside the databse to allow us
// to use the informations
func initBot(rtm *slack.RTM, msg *slack.MessageEvent, prefix, userid, fullname, email string) {
	var responseRegistered string
	var responseRowsExist string
	var id string

	text := msg.Text
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

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
		sqlCheckId := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
		row := db.QueryRow(sqlCheckId, userid)
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
			responseRegistered = "Your user was registered!"
			rtm.SendMessage(rtm.NewOutgoingMessage(responseRegistered, msg.Channel))
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
			fmt.Println(id, email)
			responseRowsExist = "Your user was updated!"
			rtm.SendMessage(rtm.NewOutgoingMessage(responseRowsExist, msg.Channel))
		default:
			panic(err)
		}

	}
}

// removeBot remove the user from the database
func removeBot(rtm *slack.RTM, msg *slack.MessageEvent, prefix, userid string) {
	var responseRegistered string
	var responseRowsExist string
	var id string

	text := msg.Text
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedSetup := map[string]bool{
		"remove": true,
		"delete": true,
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
		sqlCheckId := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
		row := db.QueryRow(sqlCheckId, userid)
		switch err := row.Scan(&id); err {
		case sql.ErrNoRows:
			responseRegistered = "Your user was not registered!"
			rtm.SendMessage(rtm.NewOutgoingMessage(responseRegistered, msg.Channel))
		// Delete the user
		case nil:
			sqlDelete := `
			DELETE FROM hatcher.users
			WHERE user_id = $1;`
			_, err = db.Exec(sqlDelete, userid)
			if err != nil {
				panic(err)
			}
			fmt.Println(id)
			responseRowsExist = "Your user was deleted!"
			rtm.SendMessage(rtm.NewOutgoingMessage(responseRowsExist, msg.Channel))
		default:
			panic(err)
		}

	}
}
