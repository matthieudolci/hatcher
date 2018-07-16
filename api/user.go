package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/matthieudolci/hatcher/database"
)

type userSummary struct {
	UserID      string         `json:"userid"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	ManagerID   sql.NullString `json:"managerid"`
	IsManager   bool           `json:"ismanager"`
	DisplayName sql.NullString `json:"displayname"`
}

type users struct {
	Users []userSummary
}

func getAllUsersHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	u := users{}

	err := getAllUsers(&u)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(&u)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func getAllUsers(u *users) error {

	rows, err := database.DB.Query(`
		SELECT
			userid,
			full_name,
			email,
			managerid,
			ismanager,
			displayname
		FROM hatcher.users;
		`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		result := userSummary{}
		err = rows.Scan(
			&result.UserID,
			&result.Name,
			&result.Email,
			&result.ManagerID,
			&result.IsManager,
			&result.DisplayName,
		)
		if err != nil {
			return err
		}
		u.Users = append(u.Users, result)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func getUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	u := users{}

	userid := ps.ByName("userid")

	err := getUser(&u, userid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := json.Marshal(&u)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

func getUser(u *users, userid string) error {

	rows, err := database.DB.Query(`
		SELECT
			userid,
			full_name,
			email,
			managerid,
			ismanager,
			displayname
		FROM hatcher.users
		WHERE userid=$1;
		`, userid)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		result := userSummary{}
		err = rows.Scan(
			&result.UserID,
			&result.Name,
			&result.Email,
			&result.ManagerID,
			&result.IsManager,
			&result.DisplayName,
		)
		if err != nil {
			return err
		}
		u.Users = append(u.Users, result)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}
