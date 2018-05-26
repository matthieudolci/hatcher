package bot

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	// Postresql Driver
	_ "github.com/lib/pq"
)

const (
	dbhost = "DBHOST"
	dbport = "DBPORT"
	dbuser = "DBUSER"
	dbpass = "DBPASS"
	dbname = "DBNAME"
)

func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, ok := os.LookupEnv(dbhost)
	if !ok {
		panic("DB_HOST environment variable required but not set")
	}
	port, ok := os.LookupEnv(dbport)
	if !ok {
		panic("DB_PORT environment variable required but not set")
	}
	user, ok := os.LookupEnv(dbuser)
	if !ok {
		panic("DB_USER environment variable required but not set")
	}
	password, ok := os.LookupEnv(dbpass)
	if !ok {
		panic("DB_PASSWORD environment variable required but not set")
	}
	name, ok := os.LookupEnv(dbname)
	if !ok {
		panic("DB_NAME environment variable required but not set")
	}
	conf[dbhost] = host
	conf[dbport] = port
	conf[dbuser] = user
	conf[dbpass] = password
	conf[dbname] = name
	return conf
}

func (s *Slack) initManager(userid, fullname, managerid, managername string) {
	config := dbConfig()
	var id string

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport], config[dbuser], config[dbpass], config[dbname])
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
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
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
		s.Logger.Printf("[DEBUG] Manager %s was added to user %s.\n", managername, fullname)
	default:
		panic(err)
	}
}

// initBot is the first step of using this bot.
// It will insert the user informations inside the databse to allow us
// to use them
func (s *Slack) initBot(userid, email, fullname string) {
	config := dbConfig()
	var id string

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport], config[dbuser], config[dbpass], config[dbname])
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
	config := dbConfig()
	var id string

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport], config[dbuser], config[dbpass], config[dbname])
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

func (s *Slack) setupIsManager(userid, fullname, ismanager string) {

	config := dbConfig()
	var id string

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport], config[dbuser], config[dbpass], config[dbname])

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
		s.Logger.Printf("[DEBUG] User %s is not registered.", fullname)
	// If the user exist we update the column manager_id
	case nil:
		sqlUpdate := `
		UPDATE hatcher.users
		SET is_manager = $2
		WHERE user_id = $1
		RETURNING id;`
		err = db.QueryRow(sqlUpdate, userid, ismanager).Scan(&userid)
		if err != nil {
			panic(err)
		}
		if ismanager == "true" {
			s.Logger.Printf("[DEBUG] %s is now setup as a manager.\n", fullname)
		} else {
			s.Logger.Printf("[DEBUG] %s is not a manager.\n", fullname)
		}

	default:
		panic(err)
	}
}

func (s *Slack) resultHappinessSurvey(userid, result string) {

	config := dbConfig()

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport], config[dbuser], config[dbpass], config[dbname])

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	sqlWrite := `
	INSERT INTO hatcher.happiness (user_id, result)
	VALUES ($1, $2)
	RETURNING id`

	err = db.QueryRow(sqlWrite, userid, result).Scan(&userid)
	if err != nil {
		panic(err)
	}

	s.Logger.Printf("[DEBUG] Happiness Survey Result written in database.\n")
}

func (s *Slack) getHappinessSurveyResults(text, userid string) string {

	config := dbConfig()
	var result string
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	times := map[string]string{
		"yesterday": fmt.Sprint(time.Now().Add(-1 * 24 * time.Hour).Truncate(24 * time.Hour).Format("2006-01-02")),
		"today":     fmt.Sprint(time.Now().Format("2006-01-02")),
	}

	var time string
	for key, value := range times {
		if strings.Contains(text, key) {
			time = value
			break
		}
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbhost], config[dbport], config[dbuser], config[dbpass], config[dbname])
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlSelect := `
	SELECT result 
	FROM hatcher.happiness 
	WHERE user_id=$1 and date=$2;`

	err = db.QueryRow(sqlSelect, userid, time).Scan(&result)
	if err != nil {
		fmt.Println(err.Error())
	}
	s.Logger.Printf("[DEBUG] Result is %s", result)
	return result
}
