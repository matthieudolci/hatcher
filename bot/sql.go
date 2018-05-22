package bot

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
	dbname   = "hatcher"
)

func (s *Slack) initManager(userid, fullname, managerid string) {
	var id string

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// Check if the user already exist
	sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
	row := db.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	// if user doesnt exit, exit
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s is registered.", fullname)
	// If the user exist we update the column manager_id
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET manager_id = $2
		WHERE user_id = $1
		RETURNING id;`
		err = db.QueryRow(sqlUpdate, userid, managerid).Scan(&userid)
		if err != nil {
			panic(err)
		}
		s.Logger.Printf("[DEBUG] User (%s) was updated.\n", fullname)
	default:
		panic(err)
	}
}

// initBot is the first step of using this bot.
// It will insert the user informations inside the databse to allow us
// to use them
func (s *Slack) initBot(userid, email, fullname string) {
	var id string

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// Check if the user already exist
	sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
	row := db.QueryRow(sqlCheckID, userid)
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
		s.Logger.Printf("[DEBUG] User (%s) was created.\n", fullname)
	// If the user exist it will update it
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
		s.Logger.Printf("[DEBUG] User (%s) was updated.\n", fullname)
	default:
		panic(err)
	}
}

// removeBot remove the user from the database
func (s *Slack) removeBot(userid, fullname string) {
	var id string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// Check if the user already exist
	sqlCheckID := `SELECT user_id FROM hatcher.users WHERE user_id=$1;`
	row := db.QueryRow(sqlCheckID, userid)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		s.Logger.Printf("[DEBUG] User %s was not registered.", fullname)
	case nil:
		sqlDelete := `
		DELETE FROM hatcher.users
		WHERE user_id = $1;`
		_, err = db.Exec(sqlDelete, userid)
		if err != nil {
			panic(err)
		}
		s.Logger.Printf("[DEBUG] User %s with id %s was deleted.", fullname, userid)
	default:
		panic(err)
	}
}