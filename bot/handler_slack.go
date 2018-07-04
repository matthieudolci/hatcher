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
		s.Logger.Printf("[ERROR] incorrect path: %s\n", r.URL.Path)
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("empty body"))
		s.Logger.Printf("[ERROR] Empty body.\n")
		return
	}
	defer r.Body.Close()

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not parse body"))
		s.Logger.Printf("[ERROR] Could not parse body.\n")
		return
	}

	// slack API calls the data POST a 'payload'
	reply := r.PostFormValue("payload")
	if len(reply) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("could not find payload"))
		s.Logger.Printf("[ERROR] Could not find payload.\n")
		return
	}

	var payload slack.AttachmentActionCallback

	err = json.NewDecoder(strings.NewReader(reply)).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not process payload"))
		s.Logger.Printf("[ERROR] Could not process payload.\n")
		return
	}

	value := payload.Actions[0].Value
	name := payload.Actions[0].Name
	api := slack.New(s.Token)
	userid := fmt.Sprintf(payload.User.ID)

	user, err := api.GetUserInfo(userid)
	if err != nil {
		s.Logger.Printf("[ERROR] Could not get user info: %s.\n", err)
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

			err := s.resultHappinessSurvey(userid, value)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not start resultHappinessSurvey for value happinessGood: %s.\n", err)
			} else {
				s.Logger.Printf("[DEBUG] Started resultHappinessSurvey for value happinessGood\n")
			}

			w.Write([]byte("Awesome, have a wonderful day!"))
		}

	case "happinessNeutral":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "happinessNeutral" {
			value := fmt.Sprintf("2")

			err := s.resultHappinessSurvey(userid, value)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not start resultHappinessSurvey for value happinessNeutral: %s.\n", err)
			} else {
				s.Logger.Printf("[DEBUG] Started resultHappinessSurvey for value happinessNeutral.\n")
			}

			w.Write([]byte("I hope your day will get better :slightly_smiling_face:"))
		}
	case "happinessSad":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "happinessSad" {
			value := fmt.Sprintf("1")

			err := s.resultHappinessSurvey(userid, value)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not start resultHappinessSurvey for value happinessSad: %s.\n", err)
			} else {
				s.Logger.Printf("[DEBUG] Started resultHappinessSurvey for value happinessSad.\n")
			}

			w.Write([]byte("I am sorry to hear that. Take all the time you need to feel better."))
		}

	case "SetupYes":
		w.Write([]byte(":white_check_mark: - Starting the setup of your user."))

		err := s.initBot(userid, email, fullname, displayname)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start initBot for value SetupYes: %s.\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Started initBot for value SetupYes.\n")
		}

		err = s.askWhoIsManager(channelid, userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start askWhoIsManager: %s.\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Start askWhoIsManager.\n")
		}

	case "SetupNo":
		w.Write([]byte("No worries, let me know if you want to later on!"))

	case "RemoveYes":
		err = s.removeBot(userid, fullname)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start removeBot: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Started removeBot.\n")
		}

		err := s.GetTimeAndUsersHappinessSurvey()
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start GetTimeAndUsersHappinessSurvey for value RemoveYes: %s.\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Started GetTimeAndUsersHappinessSurvey for value RemoveYes.\n")
		}
		w.Write([]byte("Sorry to see you go. Your user was deleted."))

	case "RemoveNo":
		w.Write([]byte("Glad you decided to stay :smiley:"))

	case "isManagerYes":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "isManagerYes" {
			value := fmt.Sprintf("true")

			err := s.setupIsManager(userid, fullname, value)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not start setupIsManager for value isManagerYes: %s.\n", err)
			} else {
				s.Logger.Printf("[DEBUG] Start setupIsManager for value isManagerYes.\n")
			}

			w.Write([]byte(fmt.Sprintf(":white_check_mark: - You are setup as a manager.")))
		}

	case "isManagerNo":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "isManagerNo" {
			value := fmt.Sprintf("false")

			err := s.setupIsManager(userid, fullname, value)
			if err != nil {
				s.Logger.Printf("[ERROR] Could not start setupIsManager for value isManagerNo: %s.\n", err)
			} else {
				s.Logger.Printf("[DEBUG] Start setupIsManager for value isManagerNo.\n")
			}

			w.Write([]byte(fmt.Sprintf(":white_check_mark: - You are not setup as a manager.")))
		}
	}

	switch name {
	case "ManagerChosen":
		managerid := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		manager, err := api.GetUserInfo(managerid)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start GetUserInfo for name ManagerChosen: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Getting user informations for name ManagerChosen\n")
		}

		managername := fmt.Sprintf(manager.RealName)

		err = s.initManager(userid, fullname, managerid, managername)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start initManager for name ManagerChosen: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Started initManager for name ManagerChosen.\n")
		}

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s was setup as your manager.", managername)))

		err = s.askIfManager(channelid, userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start askIfManager for name ManagerChosen: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Started askIfManager for name ManagerChosen.\n")
		}
	case "isManagerYes", "isManagerNo":
		err := s.askTimeHappinessSurvey(channelid, userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start askTimeHappinessSurveyfor for name isManagerYes or isManagerNo: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Started askTimeHappinessSurveyfor for name isManagerYes or isManagerNo.\n")
		}
	case "HappinessTime":
		time := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		err := s.insertTimeHappinessSurvey(userid, fullname, time)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start insertTimeHappinessSurvey for name HappinessTime: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Start insertTimeHappinessSurvey for name HappinessTime.\n")
		}

		err = s.GetTimeAndUsersHappinessSurvey()
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start GetTimeAndUsersHappinessSurvey for name HappinessTime: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Started GetTimeAndUsersHappinessSurvey for name HappinessTime.\n")
		}
		err = s.askTimeStandup(channelid, userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start askTimeStandup for name HappinessTime: %s\n", err)
		}

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - Time for happiness survey selected")))

	case "StandupTime":
		time := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		err := s.insertTimeStandup(userid, fullname, time)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start insertTimeStandup for name StandupTime: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Start insertTimeStandup for name StandupTime.\n")
		}

		err = s.askWhichChannelStandup(channelid, userid)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start askWhichChannelStandup for name StandupTime: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Start askWhichChannelStandup for name StandupTime.\n")
		}

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - Time for standup selected")))

	case "ChannelStandupChosen":
		channel := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		err := s.insertChannelStandup(userid, fullname, channel)
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start insertChannelStandup for name ChannelStandupChosen: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Start insertChannelStandup for name ChannelStandupChosen.\n")
		}

		err = s.GetTimeAndUsersStandup()
		if err != nil {
			s.Logger.Printf("[ERROR] Could not start GetTimeAndUsersStandup for name ChannelStandupChosen: %s\n", err)
		} else {
			s.Logger.Printf("[DEBUG] Start GetTimeAndUsersStandup for name ChannelStandupChosen.\n")
		}

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!", displayname)))
	}

	w.WriteHeader(http.StatusOK)
}
