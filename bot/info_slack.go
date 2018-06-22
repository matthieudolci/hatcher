package bot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

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

func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {

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
			user_id,
			full_name,
			email,
			manager_id,
			is_manager,
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

func getUserHandler(w http.ResponseWriter, r *http.Request) {

	u := users{}

	params, err := parse1Params(r, "/api/slack/users/", 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = getUser(&u, params[0])
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
			user_id,
			full_name,
			email,
			manager_id,
			is_manager,
			displayname
		FROM hatcher.users
		WHERE user_id=$1;
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
