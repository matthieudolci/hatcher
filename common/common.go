package common

import (
	"database/sql"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
	uuid "github.com/satori/go.uuid"
)

// Slack is the primary struct for our slackbot
type Slack struct {
	Name  string
	Token string

	User   string
	UserID string

	Client       *slack.Client
	MessageEvent *slack.MessageEvent
}

func RowExists(query string, args ...interface{}) bool {

	var exists bool

	err := database.DB.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.WithFields(log.Fields{
			"arg": args,
		}).WithError(err).Error("error checking if row exists")
	}
	return exists
}

func QueryRow(query string, args ...interface{}) error {

	var id int

	err := database.DB.QueryRow(query, args...).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		log.WithFields(log.Fields{
			"arg": args,
		}).WithError(err).Error("error querying the row")
	}
	return nil
}

func QueryUUID(query string, args ...interface{}) (string, error) {

	var uuid string

	err := database.DB.QueryRow(query, args...).Scan(&uuid)
	if err != nil && err != sql.ErrNoRows {
		log.WithFields(log.Fields{
			"arg": args,
		}).WithError(err).Error("error querying the row")
	}

	u := fmt.Sprint(uuid)

	return u, err
}

func CreatesUUID() string {
	u := uuid.NewV4()
	uuid := fmt.Sprint(u)
	return uuid
}
