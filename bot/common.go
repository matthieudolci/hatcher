package bot

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/matthieudolci/hatcher/database"
)

var scheduler *gocron.Scheduler
var stop chan bool

// GetTimeAndUsersForScheduler gets the time selected by a user for the Happiness survey and standup
func (s *Slack) GetTimeAndUsersForScheduler() error {
	type ScheduleData struct {
		TimesHappiness string
		TimesStandup   string
		UserID         string
	}

	rows, err := database.DB.Query("SELECT to_char(happiness_schedule, 'HH24:MI'), to_char(standup_schedule, 'HH24:MI'), userid FROM hatcher.users;")
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Printf("[ERROR] There is no result time or userid.\n")
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
			s.Logger.Printf("[ERROR] During the scan.\n")
		}
		fmt.Println(scheduledata)
		s.runScheduler(scheduledata.TimesStandup, scheduledata.TimesHappiness, scheduledata.UserID)
	}
	stop = scheduler.Start()
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		s.Logger.Printf("[ERROR] During iteration.\n")
	}
	return nil
}

// Runs the jobs standupYesterdayScheduled and askHappinessSurveyScheduled at a times defined by the user
func (s *Slack) runScheduler(timeStandup, timeHappiness, userid string) error {

	location, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		log.Println("Unfortunately can't load a location")
		log.Println(err)
	} else {
		gocron.ChangeLoc(location)
	}

	scheduler.Every(1).Monday().At(timeStandup).Do(s.standupYesterdayScheduled, userid)
	scheduler.Every(1).Tuesday().At(timeStandup).Do(s.standupYesterdayScheduled, userid)
	scheduler.Every(1).Wednesday().At(timeStandup).Do(s.standupYesterdayScheduled, userid)
	scheduler.Every(1).Thursday().At(timeStandup).Do(s.standupYesterdayScheduled, userid)
	scheduler.Every(1).Friday().At(timeStandup).Do(s.standupYesterdayScheduled, userid)

	scheduler.Every(1).Monday().At(timeHappiness).Do(s.askHappinessSurveyScheduled, userid)
	scheduler.Every(1).Tuesday().At(timeHappiness).Do(s.askHappinessSurveyScheduled, userid)
	scheduler.Every(1).Wednesday().At(timeHappiness).Do(s.askHappinessSurveyScheduled, userid)
	scheduler.Every(1).Thursday().At(timeHappiness).Do(s.askHappinessSurveyScheduled, userid)
	scheduler.Every(1).Friday().At(timeHappiness).Do(s.askHappinessSurveyScheduled, userid)

	return nil
}
