package scheduler

import (
	"database/sql"
	"time"

	"github.com/matthieudolci/hatcher/common"
	"github.com/matthieudolci/hatcher/standup"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/gocron"
	"github.com/matthieudolci/hatcher/database"
)

var scheduler *gocron.Scheduler
var stop chan bool

// GetTimeAndUsersForScheduler gets the time selected by a user for standup
func GetTimeAndUsersForScheduler(s *common.Slack) error {
	type ScheduleData struct {
		TimesStandup sql.NullString
		UserID       string
	}

	rows, err := database.DB.Query("SELECT to_char(standup_schedule, 'HH24:MI'), userid FROM hatcher.users;")
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithError(err).Error("There is no result time or userid")
		}
	}
	defer rows.Close()
	if scheduler != nil {
		stop <- true
		scheduler.Clear()
	}
	scheduler = gocron.NewScheduler()
	for rows.Next() {
		scheduledata := ScheduleData{}
		err = rows.Scan(&scheduledata.TimesStandup, &scheduledata.UserID)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}

		if scheduledata.TimesStandup.Valid {
			err := runSchedulerStandup(s, scheduledata.TimesStandup.String, scheduledata.UserID)
			if err != nil {
				log.WithError(err).Error("Running runSchedulerStandup failed")
			}
		} else {
			log.Info("Nothing to schedule standup")

		}
	}
	stop = scheduler.Start()
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		log.WithError(err).Error("During iteration")
	}
	return nil
}

// Runs the jobs standupYesterdayScheduled at a times defined by the user
func runSchedulerStandup(s *common.Slack, timeStandup, userid string) error {

	location, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		log.Println("Unfortunately can't load a location")
		log.Println(err)
	} else {
		gocron.ChangeLoc(location)
	}

	scheduler.Every(1).Monday().At(timeStandup).Do(standup.AskStandupYesterdayScheduled, s, userid)
	scheduler.Every(1).Tuesday().At(timeStandup).Do(standup.AskStandupYesterdayScheduled, s, userid)
	scheduler.Every(1).Wednesday().At(timeStandup).Do(standup.AskStandupYesterdayScheduled, s, userid)
	scheduler.Every(1).Thursday().At(timeStandup).Do(standup.AskStandupYesterdayScheduled, s, userid)
	scheduler.Every(1).Friday().At(timeStandup).Do(standup.AskStandupYesterdayScheduled, s, userid)
	log.WithFields(log.Fields{
		"userid": userid,
		"time":   timeStandup,
	}).Info("Standup schedule tasks posted")

	return nil
}
