package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/nlopes/slack"
)

// Listen on /slack for answer from the questions asked in bot_setup.go
// and dispatch to the good functions
func (s *Slack) slackPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

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
			value := fmt.Sprintf("3")
			s.resultHappinessSurvey(userid, value)
			w.Write([]byte("Awesome, have a wonderful day!"))
		}
	case "happinessNeutral":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "happinessNeutral" {
			value := fmt.Sprintf("2")
			s.resultHappinessSurvey(userid, value)
			w.Write([]byte("I hope your day will get better :slightly_smiling_face:"))
		}
	case "happinessSad":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "happinessSad" {
			value := fmt.Sprintf("1")
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
		s.GetTimeAndUsersHappinessSurvey()
		w.Write([]byte("Sorry to see you go. Your user was deleted."))
	case "RemoveNo":
		w.Write([]byte("Glad you decided to stay :smiley:"))
	case "isManagerYes":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "isManagerYes" {
			value := fmt.Sprintf("true")
			s.setupIsManager(userid, fullname, value)
			w.Write([]byte(fmt.Sprintf(":white_check_mark: - You are setup as a manager.")))
		}
	case "isManagerNo":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "isManagerNo" {
			value := fmt.Sprintf("false")
			s.setupIsManager(userid, fullname, value)
			w.Write([]byte(fmt.Sprintf(":white_check_mark: - You are not setup as a manager.")))
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
		managername := fmt.Sprintf(manager.RealName)
		s.initManager(userid, fullname, managerid, managername)
		w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s was setup as your manager.", managername)))
		s.askIfManager(channelid, userid)
	case "isManagerYes", "isManagerNo":
		err := s.askTimeHappinessSurvey(channelid, userid)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	case "HappinessTime":
		time := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)
		s.insertTimeHappinessSurvey(userid, fullname, time)
		w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!", displayname)))
		s.GetTimeAndUsersHappinessSurvey()
	}

	w.WriteHeader(http.StatusOK)
}