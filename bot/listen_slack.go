package bot

import (
	"context"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

// Slack is the primary struct for our slackbot
type Slack struct {
	Name  string
	Token string

	User   string
	UserID string

	Client       *slack.Client
	MessageEvent *slack.MessageEvent
}

// New returns a new instance of the Slack struct, primary for our slackbot
func New() (*Slack, error) {
	token := os.Getenv("SLACK_TOKEN")
	if len(token) == 0 {
		return nil, fmt.Errorf("Could not discover API token")
	}

	return &Slack{Client: slack.New(token), Token: token, Name: "hatcher"}, nil
}

// Run is the primary service to generate and kick off the slackbot listener
// This portion receives all incoming Real Time Messages notices from the workspace
// as registered by the API token
func (s *Slack) Run(ctx context.Context) error {
	authTest, err := s.Client.AuthTest()
	if err != nil {
		return fmt.Errorf("Did not authenticate: %+v", err)
	}

	s.User = authTest.User
	s.UserID = authTest.UserID

	log.WithFields(log.Fields{
		"username": s.User,
		"userid":   s.UserID,
	}).Info("Bot is now registered")

	go s.run(ctx)
	return nil

}

func (s *Slack) run(ctx context.Context) {
	// slack.SetLogger()
	// s.Client.SetDebug(true)

	rtm := s.Client.NewRTM()
	go rtm.ManageConnection()

	log.Info("Now listening for incoming messages...")

	for msg := range rtm.IncomingEvents {
		switch x := msg.Data.(type) {
		case *slack.MessageEvent:
			if x.SubType == "message_changed" {
				err := s.checkIfYesterdayMessageEdited(x)
				if err != nil {
					log.WithFields(log.Fields{
						"userid": x.User,
					}).WithError(err).Error("Checking if yesterday standup was edited")
				}
				err = s.checkIfTodayMessageEdited(x)
				if err != nil {
					log.WithFields(log.Fields{
						"userid": x.User,
					}).WithError(err).Error("Checking if yesterday standup was edited")
				}
				err = s.checkIfBlockerMessageEdited(x)
				if err != nil {
					log.WithFields(log.Fields{
						"userid": x.User,
					}).WithError(err).Error("Checking if yesterday standup was edited")
				}
			}
		}
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if len(ev.User) == 0 {
				continue
			}

			// check if we have a DM, or standard channel post
			direct := strings.HasPrefix(ev.Msg.Channel, "D")
			inchannel := strings.Contains(ev.Msg.Text, "@"+s.UserID)

			if !direct && !inchannel {
				// msg not for us!
				continue
			}

			user, err := s.Client.GetUserInfo(ev.User)
			if err != nil {
				log.WithFields(log.Fields{
					"userid": ev.User,
				}).Printf("Could not grab user information.")
				continue
			}

			log.WithFields(log.Fields{
				"username": user.Profile.RealName,
				"userid":   ev.User,
			}).Info("Received message")

			err = s.askSetup(ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting setup reply to user")
			}

			err = s.askRemove(ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting remove reply to user")
			}

			err = s.askHappinessSurvey(ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting happiness survey reply to user")
			}

			err = s.standupYesterday(ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting yesterday standup note question to user")
			}

			err = s.askSetupTimeHappinessSurvey(ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting happiness survey question to user")
			}

			err = s.askRemoveHappiness(ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting remove from happiness survey question to user")
			}

			err = s.askHelp(ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting help message")
			}

		case *slack.RTMError:
			log.Errorf("%s", ev.Error())
		}
	}
}
