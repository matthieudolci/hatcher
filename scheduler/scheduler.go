package scheduler

import (
	"database/sql"
	"time"

	"github.com/matthieudolci/hatcher/common"
	"github.com/matthieudolci/hatcher/standup"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/gocron"
	"github.com/matthieudolci/hatcher/database"
	"github.com/matthieudolci/hatcher/happiness"
)

var scheduler *gocron.Scheduler
var stop chan bool

// GetTimeAndUsersForScheduler gets the time selected by a user for the Happiness survey and standup
func GetTimeAndUsersForScheduler(s *common.Slack) error {
	type ScheduleData struct {
		TimesHappiness sql.NullString
		TimesStandup   sql.NullString
		UserID         string
	}

	rows, err := database.DB.Query("SELECT to_char(happiness_schedule, 'HH24:MI'), to_char(standup_schedule, 'HH24:MI'), userid FROM hatcher.users;")
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
		err = rows.Scan(&scheduledata.TimesHappiness, &scheduledata.TimesStandup, &scheduledata.UserID)
		if err != nil {
			log.WithError(err).Error("During the scan")
		}

		if scheduledata.TimesHappiness.Valid {
			err := runSchedulerHappiness(s, scheduledata.TimesHappiness.String, scheduledata.UserID)
			if err != nil {
				log.WithError(err).Error("Running runSchedulerHappiness failed")
			}
		} else {
			log.Info("Nothing to schedule happiness")
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

// Runs the jobs askHappinessSurveyScheduled at a times defined by the user
func runSchedulerHappiness(s *common.Slack, timeHappiness, userid string) error {

	location, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		log.Println("Unfortunately can't load a location")
		log.Println(err)
	} else {
		gocron.ChangeLoc(location)
	}

	scheduler.Every(1).Monday().At(timeHappiness).Do(happiness.AskHappinessSurveyScheduled, s, userid)
	scheduler.Every(1).Tuesday().At(timeHappiness).Do(happiness.AskHappinessSurveyScheduled, s, userid)
	scheduler.Every(1).Wednesday().At(timeHappiness).Do(happiness.AskHappinessSurveyScheduled, s, userid)
	scheduler.Every(1).Thursday().At(timeHappiness).Do(happiness.AskHappinessSurveyScheduled, s, userid)
	scheduler.Every(1).Friday().At(timeHappiness).Do(happiness.AskHappinessSurveyScheduled, s, userid)
	log.WithFields(log.Fields{
		"userid": userid,
		"time":   timeHappiness,
	}).Info("Happiness Survey schedule tasks posted")

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
