package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/matthieudolci/hatcher/database"
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
