package bot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/matthieudolci/hatcher/database"
	"github.com/nlopes/slack"
)

type resultSummary struct {
	UserID      string         `json:"userid"`
	Result      string         `json:"result"`
	Date        string         `json:"date"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	DisplayName sql.NullString `json:"displayname"`
}

type results struct {
	Results []resultSummary
}

// Ask how are the users doing
func (s *Slack) askHappinessSurvey(ev *slack.MessageEvent) error {
	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedHowAreYou := map[string]bool{
		"how's it going?":    true,
		"how are you?":       true,
		"feeling okay?":      true,
		"how are you doing?": true,
	}

	if acceptedHowAreYou[text] {
		params := slack.PostMessageParameters{}
		attachment := slack.Attachment{
			Text:       "I am good. How are you today?",
			CallbackID: fmt.Sprintf("ask_%s", ev.User),
			Color:      "#AED6F1",
			Actions: []slack.AttachmentAction{
				{
					Name:  "happinessGood",
					Text:  ":smiley:",
					Type:  "button",
					Value: "happinessGood",
				},
				{
					Name:  "happinessNeutral",
					Text:  ":neutral_face:",
					Type:  "button",
					Value: "happinessNeutral",
				},
				{
					Name:  "happinessSad",
					Text:  ":cry:",
					Type:  "button",
					Value: "happinessSad",
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
				"channelid": ev.Channel,
				"userid":    ev.User,
			}).WithError(err).Error("Could not post askHappinessSurvey question")
		}
		log.WithFields(log.Fields{
			"channelid": ev.Channel,
			"userid":    ev.User,
		}).Info("askHappinessSurvey question posted")
	}
	return nil
}

// Insert into the database the result of the happiness survey
func (s *Slack) resultHappinessSurvey(userid, result string) error {

	t := time.Now().Local().Format("2006-01-02")
	t2 := time.Now().Local().Format("15:04:05")
	date := fmt.Sprint(t)
	time := fmt.Sprint(t2)

	sqlWrite := `
	INSERT INTO hatcher.happiness (userid, results, date, time)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	err := database.DB.QueryRow(sqlWrite, userid, result, date, time).Scan(&userid)
	if err != nil {
		log.WithFields(log.Fields{
			"userid": userid,
		}).WithError(err).Error("Couldn't insert in the database the result of the happiness survey")
	}
	log.WithFields(log.Fields{
		"userid": userid,
	}).Info("Happiness Survey Result written in database")
	return nil
}

// Ask how are the users doing
func (s *Slack) askHappinessSurveyScheduled(userid string) error {

	_, _, channelid, _ := s.Client.OpenIMChannel(userid)
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "How are you today?",
		CallbackID: fmt.Sprintf("ask_%s", userid),
		Color:      "#AED6F1",
		Actions: []slack.AttachmentAction{
			{
				Name:  "happinessGood",
				Text:  ":smiley:",
				Type:  "button",
				Value: "happinessGood",
			},
			{
				Name:  "happinessNeutral",
				Text:  ":neutral_face:",
				Type:  "button",
				Value: "happinessNeutral",
			},
			{
				Name:  "happinessSad",
				Text:  ":cry:",
				Type:  "button",
				Value: "happinessSad",
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
			"channelid": channelid,
			"userid":    userid,
		}).WithError(err).Error("Could not post askHappinessSurveyScheduled message")
	}
	log.WithFields(log.Fields{
		"channelid": channelid,
		"userid":    userid,
	}).Info("Message for askHappinessSurveyScheduled posted")
	return nil
}

func surveyResultsUserDayHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	res := results{}

	userid := ps.ByName("userid")
	date := ps.ByName("date")

	err := surveyResultsUserDay(&res, userid, date)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func surveyResultsUserDay(res *results, userid, date string) error {

	rows, err := database.DB.Query(`
		SELECT
			users.userid,
			happiness.results,
			happiness.date,
			users.full_name,
			users.email,
			users.displayname
		FROM hatcher.happiness, hatcher.users
		WHERE happiness.userid = users.userid
		AND happiness.userid=$1 and happiness.date=$2;
		`, userid, date)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		result := resultSummary{}
		err = rows.Scan(
			&result.UserID,
			&result.Result,
			&result.Date,
			&result.Name,
			&result.Email,
			&result.DisplayName,
		)
		if err != nil {
			return err
		}
		res.Results = append(res.Results, result)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func surveyResultsUserAllHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	res := results{}

	userid := ps.ByName("userid")

	err := surveyResultsUserAll(&res, userid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func surveyResultsUserAll(res *results, userid string) error {

	rows, err := database.DB.Query(`
		SELECT
			users.userid,
			happiness.results,
			happiness.date,
			users.full_name,
			users.email,
			users.displayname
		FROM hatcher.happiness, hatcher.users
		WHERE happiness.userid = users.userid
		AND happiness.userid=$1;
		`, userid)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		result := resultSummary{}
		err = rows.Scan(
			&result.UserID,
			&result.Result,
			&result.Date,
			&result.Name,
			&result.Email,
			&result.DisplayName,
		)
		if err != nil {
			return err
		}
		res.Results = append(res.Results, result)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func surveyResultsUserBetweenDatesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	res := results{}

	userid := ps.ByName("userid")
	date1 := ps.ByName("date1")
	date2 := ps.ByName("date2")

	err2 := surveyResultsUserBetweenDates(&res, userid, date1, date2)
	if err2 != nil {
		http.Error(w, err2.Error(), 500)
		return
	}

	out, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func surveyResultsUserBetweenDates(res *results, userid, date1, date2 string) error {

	rows, err := database.DB.Query(`
		SELECT
			users.userid,
			happiness.results,
			happiness.date,
			users.full_name,
			users.email,
			users.displayname
		FROM hatcher.happiness, hatcher.users
		WHERE happiness.userid = users.userid
		AND happiness.userid=$1
		AND happiness.date BETWEEN $2 AND $3;
		`, userid, date1, date2)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		result := resultSummary{}
		err = rows.Scan(
			&result.UserID,
			&result.Result,
			&result.Date,
			&result.Name,
			&result.Email,
			&result.DisplayName,
		)
		if err != nil {
			return err
		}
		res.Results = append(res.Results, result)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func surveyResultsAllHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	res := results{}

	err := surveyResultsAll(&res)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func surveyResultsAll(res *results) error {

	rows, err := database.DB.Query(`
		SELECT
			users.userid,
			happiness.results,
			happiness.date,
			users.full_name,
			users.email,
			users.displayname
		FROM hatcher.happiness, hatcher.users
		WHERE happiness.userid = users.userid;
		`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		result := resultSummary{}
		err = rows.Scan(
			&result.UserID,
			&result.Result,
			&result.Date,
			&result.Name,
			&result.Email,
			&result.DisplayName,
		)
		if err != nil {
			return err
		}
		res.Results = append(res.Results, result)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func surveyResultsAllUserBetweenDatesHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	res := results{}

	date1 := ps.ByName("date1")
	date2 := ps.ByName("date2")

	err := surveyResultsAllUserBetweenDates(&res, date1, date2)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func surveyResultsAllUserBetweenDates(res *results, date1, date2 string) error {

	rows, err := database.DB.Query(`
		SELECT
			users.userid,
			happiness.results,
			happiness.date,
			users.full_name,
			users.email,
			users.displayname
		FROM hatcher.happiness, hatcher.users
		WHERE happiness.userid = users.userid
		AND happiness.date BETWEEN $1 AND $2;
		`, date1, date2)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		result := resultSummary{}
		err = rows.Scan(
			&result.UserID,
			&result.Result,
			&result.Date,
			&result.Name,
			&result.Email,
			&result.DisplayName,
		)
		if err != nil {
			return err
		}
		res.Results = append(res.Results, result)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}
