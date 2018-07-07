package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/nlopes/slack"
)

// Listen on /slack for answer from the questions asked in bot_setup.go
// and dispatch to the good functions
func (s *Slack) slackPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	if r.URL.Path != "/slack" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("incorrect path: %s", r.URL.Path)))
		log.WithFields(log.Fields{
			"urlpath": r.URL.Path,
		}).Error("incorrect path")
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("empty body"))
		log.Error("Empty body")
		return
	}
	defer r.Body.Close()

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not parse body"))
		log.Error("Could not parse body")
		return
	}

	// slack API calls the data POST a 'payload'
	reply := r.PostFormValue("payload")
	if len(reply) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("could not find payload"))
		log.Error("Could not find payload")
		return
	}

	var payload slack.AttachmentActionCallback

	err = json.NewDecoder(strings.NewReader(reply)).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not process payload"))
		log.Error("Could not process payload")
		return
	}

	value := payload.Actions[0].Value
	name := payload.Actions[0].Name
	api := slack.New(s.Token)
	userid := fmt.Sprintf(payload.User.ID)

	user, err := api.GetUserInfo(userid)
	if err != nil {
		log.WithError(err).Error("Could not get user info")
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
				log.WithError(err).Error("Could not start resultHappinessSurvey for value happinessGood")
			}
			log.Info("Started resultHappinessSurvey for value happinessGood")

			w.Write([]byte("Awesome, have a wonderful day!"))
		}

	case "happinessNeutral":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "happinessNeutral" {
			value := fmt.Sprintf("2")

			err := s.resultHappinessSurvey(userid, value)
			if err != nil {
				log.WithError(err).Error("Could not start resultHappinessSurvey for value happinessNeutral")
			}
			log.Info("Started resultHappinessSurvey for value happinessNeutral")

			w.Write([]byte("I hope your day will get better :slightly_smiling_face:"))
		}
	case "happinessSad":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "happinessSad" {
			value := fmt.Sprintf("1")

			err := s.resultHappinessSurvey(userid, value)
			if err != nil {
				log.WithError(err).Error("Could not start resultHappinessSurvey for value happinessSad")
			}
			log.Info("Started resultHappinessSurvey for value happinessSad")

			w.Write([]byte("I am sorry to hear that. Take all the time you need to feel better."))
		}

	case "SetupYes":
		w.Write([]byte(":white_check_mark: - Starting the setup of your user."))

		err := s.initBot(userid, email, fullname, displayname)
		if err != nil {
			log.WithError(err).Error("Could not start initBot for value SetupYes")
		}
		log.Info("Started initBot for value SetupYes")

		err = s.askWhoIsManager(channelid, userid)
		if err != nil {
			log.WithError(err).Error("Could not start askWhoIsManager")
		}
		log.Info("Start askWhoIsManager")

	case "SetupNo":
		w.Write([]byte("No worries, let me know if you want to later on!"))

	case "RemoveYes":
		err = s.removeBot(userid, fullname)
		if err != nil {
			log.WithError(err).Error("Could not start removeBot")
		}
		log.Info("Started removeBot")

		err := s.GetTimeAndUsersForScheduler()
		if err != nil {
			log.WithError(err).Error("Could not start GetTimeAndUsersHappinessSurvey for value RemoveYes")
		}
		log.Info("Started GetTimeAndUsersHappinessSurvey for value RemoveYes")

		w.Write([]byte("Sorry to see you go. Your user was deleted."))

	case "RemoveNo":
		w.Write([]byte("Glad you decided to stay :smiley:"))

	case "isManagerYes":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "isManagerYes" {
			value := fmt.Sprintf("true")

			err := s.setupIsManager(userid, fullname, value)
			if err != nil {
				log.WithError(err).Error("Could not start setupIsManager for value isManagerYes")
			}
			log.Info("Start setupIsManager for value isManagerYes")

			w.Write([]byte(fmt.Sprintf(":white_check_mark: - You are setup as a manager.")))
		}

	case "isManagerNo":
		answer := fmt.Sprintf(payload.Actions[0].Value)

		if answer == "isManagerNo" {
			value := fmt.Sprintf("false")

			err := s.setupIsManager(userid, fullname, value)
			if err != nil {
				log.WithError(err).Error("Could not start setupIsManager for value isManagerNo")
			}
			log.Info("Start setupIsManager for value isManagerNo")

			w.Write([]byte(fmt.Sprintf(":white_check_mark: - You are not setup as a manager.")))
		}
	}

	switch name {
	case "ManagerChosen":
		managerid := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		manager, err := api.GetUserInfo(managerid)
		if err != nil {
			log.WithError(err).Error("Could not start GetUserInfo for name ManagerChosen")
		}
		log.Info("Getting user informations for name ManagerChosen")

		managername := fmt.Sprintf(manager.RealName)

		err = s.initManager(userid, fullname, managerid, managername)
		if err != nil {
			log.WithError(err).Error("Could not start initManager for name ManagerChosen")
		}
		log.Info("Started initManager for name ManagerChosen")

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s was setup as your manager.", managername)))

		err = s.askIfManager(channelid, userid)
		if err != nil {
			log.WithError(err).Error("Could not start askIfManager for name ManagerChosen")
		}
		log.Info("Started askIfManager for name ManagerChosen")

	case "isManagerYes", "isManagerNo":
		err := s.askTimeHappinessSurvey(channelid, userid)
		if err != nil {
			log.WithError(err).Error("Could not start askTimeHappinessSurveyfor for name isManagerYes or isManagerNo")
		}
		log.Info("Started askTimeHappinessSurveyfor for name isManagerYes or isManagerNo")
	case "HappinessTime":
		time := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		err := s.insertTimeHappinessSurvey(userid, fullname, time)
		if err != nil {
			log.WithError(err).Error("Could not start insertTimeHappinessSurvey for name HappinessTime")
		}
		log.Info("Start insertTimeHappinessSurvey for name HappinessTime")

		err = s.askTimeStandup(channelid, userid)
		if err != nil {
			log.WithError(err).Error("Could not start askTimeStandup for name HappinessTime")
		}
		log.Info("Start insertTimaskTimeStandupeHappinessSurvey for name HappinessTime")

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - Time for happiness survey selected")))

	case "StandupTime":
		time := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		err := s.insertTimeStandup(userid, fullname, time)
		if err != nil {
			log.WithError(err).Error("Could not start insertTimeStandup for name StandupTime")
		}
		log.Info("Start insertTimeStandup for name StandupTime")

		err = s.askWhichChannelStandup(channelid, userid)
		if err != nil {
			log.WithError(err).Error("Could not start askWhichChannelStandup for name StandupTime")
		}
		log.Info("Start askWhichChannelStandup for name StandupTime")

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - Time for standup selected")))

	case "ChannelStandupChosen":
		channel := fmt.Sprintf(payload.Actions[0].SelectedOptions[0].Value)

		err := s.insertChannelStandup(userid, fullname, channel)
		if err != nil {
			log.WithError(err).Error("Could not start insertChannelStandup for name ChannelStandupChosen")
		}
		log.Info("Start insertChannelStandup for name ChannelStandupChosen")

		err = s.GetTimeAndUsersForScheduler()
		if err != nil {
			log.WithError(err).Error("Could not start GetTimeAndUsersForScheduler for name ChannelStandupChosen")
		}
		log.Info("Start GetTimeAndUsersForScheduler for name ChannelStandupChosen")

		w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!", displayname)))
	}

	w.WriteHeader(http.StatusOK)
}
