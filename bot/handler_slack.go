package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/nlopes/slack"
)

// Slack is the primary struct for our slackbot
type Slack struct {
	Name  string
	Token string

	User   string
	UserID string

	Logger *log.Logger

	Client       *slack.Client
	MessageEvent *slack.MessageEvent
}

// PostMap is a global map to handle callbacks depending on the provided user
// This mapping stores off the userID to reply to
var PostMap map[string]string

// PostLock is the complement for the global PostMap to ensure concurrent
// access doesn't race
var PostLock sync.RWMutex

func slackHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/slack" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("incorrect path: %s", r.URL.Path)))
		return
	}

	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%v", `¯\_(ツ)_/¯ GET`)))
		return
	case "POST":
		w.WriteHeader(http.StatusMovedPermanently)
		w.Write([]byte("cannot post to this endpoint"))
		return
	default:
	}
}

func (s *Slack) slackPostHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/slack" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("incorrect path: %s", r.URL.Path)))
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("empty body"))
		return
	}
	defer r.Body.Close()

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not parse body"))
		return
	}

	// slack API calls the data POST a 'payload'
	reply := r.PostFormValue("payload")
	if len(reply) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("could not find payload"))
		return
	}

	var payload slack.AttachmentActionCallback
	err = json.NewDecoder(strings.NewReader(reply)).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not process payload"))
		return
	}

	value := payload.Actions[0].Value
	name := payload.Actions[0].Name
	api := slack.New(s.Token)
	userid := fmt.Sprintf(payload.User.ID)
	user, err := api.GetUserInfo(userid)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fullname := fmt.Sprintf(user.Profile.RealName)
	displayname := fmt.Sprintf(user.Profile.DisplayName)
	email := fmt.Sprintf(user.Profile.Email)
	channelid := fmt.Sprintf(payload.Channel.ID)

	switch value {
	case "happinessGood":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "happinessGood" {
			value := fmt.Sprintf("good")
			s.resultHappinessSurvey(userid, value)
			w.Write([]byte("Awesome, have a wonderful day!"))
		}
	case "happinessNeutral":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "happinessNeutral" {
			value := fmt.Sprintf("neutral")
			s.resultHappinessSurvey(userid, value)
			w.Write([]byte("I hope your day will get better :slightly_smiling_face:"))
		}
	case "happinessSad":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "happinessSad" {
			value := fmt.Sprintf("sad")
			s.resultHappinessSurvey(userid, value)
			w.Write([]byte("I am sorry to hear that. Take all the time you need to feel better."))
		}
	case "SetupYes":
		w.Write([]byte(":white_check_mark: - Starting the setup of your user."))
		s.initBot(userid, email, fullname, displayname)
		s.askWhoIsManager(channelid, userid)
	case "SetupNo":
		w.Write([]byte("No worries, let me know if you want to later on!"))
	case "RemoveYes":
		s.removeBot(userid, fullname)
		w.Write([]byte("Sorry to see you go. Your user was deleted."))
	case "RemoveNo":
		w.Write([]byte("Glad you decided to stay :smiley:"))
	case "isManagerYes":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "isManagerYes" {
			value := fmt.Sprintf("true")
			s.setupIsManager(userid, fullname, value)
			w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!", displayname)))
		}
	case "isManagerNo":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "isManagerNo" {
			value := fmt.Sprintf("false")
			s.setupIsManager(userid, fullname, value)
			w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!", displayname)))
		}
	}

	switch name {
	case "ManagerChosen":
		managerid := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)
		manager, err := api.GetUserInfo(managerid)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		managername := fmt.Sprintf(manager.Profile.DisplayName)
		s.initManager(userid, fullname, managerid, managername)
		w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s was setup as your manager.", managername)))
		s.askIfManager(channelid, userid)
	}

	w.WriteHeader(http.StatusOK)
}
