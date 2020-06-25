package bot

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/matthieudolci/hatcher/common"
	"github.com/matthieudolci/hatcher/help"
	"github.com/matthieudolci/hatcher/setup"
	"github.com/matthieudolci/hatcher/standup"

	log "github.com/Sirupsen/logrus"
	"github.com/slack-go/slack"
)

// New returns a new instance of the Slack struct, primary for our slackbot
func New() (*common.Slack, error) {
	token := os.Getenv("SLACK_TOKEN")
	if len(token) == 0 {
		return nil, fmt.Errorf("Could not discover API token")
	}

	return &common.Slack{Client: slack.New(token), Token: token, Name: "hatcher"}, nil
}

// Run is the primary service to generate and kick off the slackbot listener
// This portion receives all incoming Real Time Messages notices from the workspace
// as registered by the API token
func Run(ctx context.Context, s *common.Slack) error {
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

	go run(ctx, s)
	return nil

}

func run(ctx context.Context, s *common.Slack) {
	// slack.SetLogger()
	// s.Client.SetDebug(true)

	rtm := s.Client.NewRTM()
	go rtm.ManageConnection()

	log.Info("Now listening for incoming messages...")

	for msg := range rtm.IncomingEvents {
		switch x := msg.Data.(type) {
		case *slack.MessageEvent:
			if x.SubType == "message_changed" {
				err := standup.CheckIfYesterdayMessageEdited(s, x)
				if err != nil {
					log.WithFields(log.Fields{
						"userid": x.User,
					}).WithError(err).Error("Checking if yesterday standup was edited")
				}
				err = standup.CheckIfTodayMessageEdited(s, x)
				if err != nil {
					log.WithFields(log.Fields{
						"userid": x.User,
					}).WithError(err).Error("Checking if yesterday standup was edited")
				}
				err = standup.CheckIfBlockerMessageEdited(s, x)
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

			err = setup.AskSetup(s, ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting setup reply to user")
			}

			err = setup.AskRemove(s, ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting remove reply to user")
			}

			err = standup.AskStandupYesterday(s, ev)
			if err != nil {
				log.WithFields(log.Fields{
					"username": user.Profile.RealName,
					"userid":   ev.User,
				}).WithError(err).Error("Posting yesterday standup note question to user")
			}

			err = help.AskHelp(s, ev)
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
