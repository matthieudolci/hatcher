package bot

import (
	"database/sql"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/database"
	uuid "github.com/satori/go.uuid"
)

func (s *Slack) rowExists(query string, args ...interface{}) bool {

	var exists bool

	err := database.DB.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.WithFields(log.Fields{
			"arg": args,
		}).WithError(err).Error("error checking if row exists")
	}
	return exists
}

func (s *Slack) queryRow(query string, args ...interface{}) error {

	var id int

	err := database.DB.QueryRow(query, args...).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		log.WithFields(log.Fields{
			"arg": args,
		}).WithError(err).Error("error querying the row")
	}
	return nil
}

func (s *Slack) queryUUID(query string, args ...interface{}) (string, error) {

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

func (s *Slack) createsUUID() string {
	u := uuid.NewV4()
	uuid := fmt.Sprint(u)
	return uuid
}
