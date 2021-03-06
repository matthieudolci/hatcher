package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/matthieudolci/hatcher/common"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/matthieudolci/hatcher/scheduler"
	"github.com/matthieudolci/hatcher/setup"
	"github.com/slack-go/slack"
)

const isManagerYes = "isManagerYes"

// SlackPostHandler listen on /slack for answer from the questions asked in setup.go
// and dispatch to the good functions
func SlackPostHandler(s *common.Slack) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if r.URL.Path != "/slack" {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte(fmt.Sprintf("incorrect path: %s", r.URL.Path)))
			if err != nil {
				log.WithError(err).Error("Could not post the message incorrect path")
			}
			log.WithFields(log.Fields{
				"urlpath": r.URL.Path,
			}).Error("incorrect path")
			return
		}

		if r.Body == nil {
			w.WriteHeader(http.StatusNotAcceptable)
			_, err := w.Write([]byte("empty body"))
			if err != nil {
				log.WithError(err).Error("empty body")
			}
			return
		}
		defer r.Body.Close()

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusGone)
			_, err = w.Write([]byte("could not parse body"))
			if err != nil {
				log.WithError(err).Error("could not parse body")
			}
			return
		}

		// slack API calls the data POST a 'payload'
		reply := r.PostFormValue("payload")
		if len(reply) == 0 {
			w.WriteHeader(http.StatusNoContent)
			_, err = w.Write([]byte("could not find payload"))
			if err != nil {
				log.WithError(err).Error("Could not find payload")
			}
			return
		}

		var payload slack.InteractionCallback

		err = json.NewDecoder(strings.NewReader(reply)).Decode(&payload)
		if err != nil {
			w.WriteHeader(http.StatusGone)
			_, err = w.Write([]byte("could not process payload"))
			if err != nil {
				log.WithError(err).Error("Could not process payload")
			}
			return
		}

		value := payload.ActionCallback.AttachmentActions[0].Value
		name := payload.ActionCallback.AttachmentActions[0].Name
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
		case "SetupYes":
			_, err := w.Write([]byte(":white_check_mark: - Starting the setup of your user."))
			if err != nil {
				log.WithError(err).Error("Could not post the message Starting the setup of your user")
			}

			err = setup.InitBot(userid, email, fullname, displayname)
			if err != nil {
				log.WithError(err).Error("Could not start initBot for value SetupYes")
			}
			log.Info("Started initBot for value SetupYes")

			err = setup.AskWhoIsManager(s, channelid, userid)
			if err != nil {
				log.WithError(err).Error("Could not start askWhoIsManager")
			}
			log.Info("Start askWhoIsManager")

		case "SetupNo":
			_, err := w.Write([]byte("No worries, let me know if you want to later on!"))
			if err != nil {
				log.WithError(err).Error("Could not post the message No worries, let me know if you want to later on")
			}

		case "RemoveYes":
			err = setup.RemoveBot(userid, fullname)
			if err != nil {
				log.WithError(err).Error("Could not start removeBot")
			}
			log.Info("Started removeBot")

			err := scheduler.GetTimeAndUsersForScheduler(s)
			if err != nil {
				log.WithError(err).Error("Could not start GetTimeAndUsersForScheduler for value RemoveYes")
			}
			log.Info("Started GetTimeAndUsersForScheduler for value RemoveYes")

			_, err = w.Write([]byte("Sorry to see you go. Your user was deleted."))
			if err != nil {
				log.WithError(err).Error("Could not post the message Sorry to see you go")
			}

		case "RemoveNo":
			_, err := w.Write([]byte("Glad you decided to stay :smiley:"))
			if err != nil {
				log.WithError(err).Error("Could not post the message Glad you decided to stay")
			}

		case isManagerYes:
			answer := fmt.Sprintf(payload.ActionCallback.AttachmentActions[0].Value)

			if answer == isManagerYes {
				value := fmt.Sprintf("true")

				err := setup.IsManager(userid, fullname, value)
				if err != nil {
					log.WithError(err).Error("Could not start IsManager for value isManagerYes")
				}
				log.Info("Start IsManager for value isManagerYes")

				_, err = w.Write([]byte(fmt.Sprintf(":white_check_mark: - You are setup as a manager.")))
				if err != nil {
					log.WithError(err).Error("Could not post You are setup as a manager message")
				}
			}

		case "isManagerNo":
			answer := fmt.Sprintf(payload.ActionCallback.AttachmentActions[0].Value)

			if answer == "isManagerNo" {
				value := fmt.Sprintf("false")

				err := setup.IsManager(userid, fullname, value)
				if err != nil {
					log.WithError(err).Error("Could not start IsManager for value isManagerNo")
				}
				log.Info("Start IsManager for value isManagerNo")

				_, err = w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!\n", displayname)))
				if err != nil {
					log.WithError(err).Error("Could not post the final message of the setup for none manager user")
				}
			}
		}

		switch name {
		case "ManagerChosen":
			managerid := fmt.Sprintf(payload.ActionCallback.AttachmentActions[0].SelectedOptions[0].Value)

			manager, err := api.GetUserInfo(managerid)
			if err != nil {
				log.WithError(err).Error("Could not start GetUserInfo for name ManagerChosen")
			}
			log.Info("Getting user informations for name ManagerChosen")

			managername := fmt.Sprintf(manager.RealName)

			err = setup.InitManager(userid, fullname, managerid, managername)
			if err != nil {
				log.WithError(err).Error("Could not start initManager for name ManagerChosen")
			}
			log.Info("Started initManager for name ManagerChosen")

			_, err = w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s was setup as your manager.", managername)))
			if err != nil {
				log.WithError(err).Error("Could not post the you are setup as a manager")
			}

			timestandup := setup.GetStandupTimeFromManager(managerid)
			channel := setup.GetStandupChannelFromManager(managerid)

			if len(timestandup) > 0 {
				err = setup.UpdateTimeStandup(managerid, userid, timestandup)
				if err != nil {
					log.WithError(err).Error("Could not start updateTimeStandup for name ManagerChosen")
				}
				log.Info("Started updateTimeStandup for name ManagerChosen")
			} else {
				err = setup.UpdateTimeStandup(managerid, userid, "NULL")
				if err != nil {
					log.WithError(err).Error("Could not start updateTimeStandup for name ManagerChosen")
				}
				log.Info("Started updateTimeStandup for name ManagerChosen")
			}

			err = setup.UpdateChannelStandup(managerid, userid, channel)
			if err != nil {
				log.WithError(err).Error("Could not start updateTimeStandup for name ManagerChosen")
			}
			log.Info("Started updateTimeStandup for name ManagerChosen")

			err = scheduler.GetTimeAndUsersForScheduler(s)
			if err != nil {
				log.WithError(err).Error("Could not start GetTimeAndUsersForScheduler for name ChannelStandupChosen")
			}
			log.Info("Start GetTimeAndUsersForScheduler for name ChannelStandupChosen")

			err = setup.AskIfManager(s, channelid, userid)
			if err != nil {
				log.WithError(err).Error("Could not start askIfManager for name ManagerChosen")
			}
			log.Info("Started askIfManager for name ManagerChosen")

		case "NoManagerChosen":
			_, err = w.Write([]byte(fmt.Sprintf(":white_check_mark: - No manager selected")))
			if err != nil {
				log.WithError(err).Error("Could not post the you are setup as a manager")
			}

			err = setup.AskIfManager(s, channelid, userid)
			if err != nil {
				log.WithError(err).Error("Could not start askIfManager for name NoManagerChosen")
			}
			log.Info("Started askIfManager for name NoManagerChosen")

		case isManagerYes:
			err = setup.AskTimeStandup(s, channelid, userid)
			if err != nil {
				log.WithError(err).Error("Could not start askTimeStandup")
			}
			log.Info("Start askTimeStandup")

		case "StandupTime":
			time := fmt.Sprintf(payload.ActionCallback.AttachmentActions[0].SelectedOptions[0].Value)

			err := setup.InsertTimeStandup(userid, fullname, time)
			if err != nil {
				log.WithError(err).Error("Could not start insertTimeStandup for name StandupTime")
			}
			log.Info("Start insertTimeStandup for name StandupTime")

			err = setup.AskWhichChannelStandup(s, channelid, userid)
			if err != nil {
				log.WithError(err).Error("Could not start askWhichChannelStandup for name StandupTime")
			}
			log.Info("Start askWhichChannelStandup for name StandupTime")

			_, err = w.Write([]byte(fmt.Sprintf(":white_check_mark: - Team standup setup")))
			if err != nil {
				log.WithError(err).Error("Could not post the message Team standup setup")
			}

		case "ChannelStandupChosen":
			channel := fmt.Sprintf(payload.ActionCallback.AttachmentActions[0].SelectedOptions[0].Value)

			err := setup.InsertChannelStandup(userid, fullname, channel)
			if err != nil {
				log.WithError(err).Error("Could not start insertChannelStandup for name ChannelStandupChosen")
			}
			log.Info("Start insertChannelStandup for name ChannelStandupChosen")

			err = scheduler.GetTimeAndUsersForScheduler(s)
			if err != nil {
				log.WithError(err).Error("Could not start GetTimeAndUsersForScheduler for name ChannelStandupChosen")
			}
			log.Info("Start GetTimeAndUsersForScheduler for name ChannelStandupChosen")

			_, err = w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!", displayname)))
			if err != nil {
				log.WithError(err).Error("Could not post the message Your user is now setup")
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
