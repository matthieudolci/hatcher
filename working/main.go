package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:

		case *slack.ConnectedEvent:
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)
			info := rtm.GetInfo()
			prefix := fmt.Sprintf("<@%s> ", info.User.ID)
			direct := strings.HasPrefix(ev.Msg.Channel, "D")
			getuserid := ev.Msg.User
			userid := fmt.Sprintf(getuserid)
			user, err := api.GetUserInfo(userid)
			fullName := fmt.Sprintf(user.Profile.RealName)
			email := fmt.Sprintf(user.Profile.Email)

			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}

			if direct || ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
				respondHowAreYou(api, rtm, ev, prefix)
			}

			if direct || ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {

				initBot(rtm, ev, prefix, userid, fullName, email)
			}

			if direct || ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
				removeBot(rtm, ev, prefix, userid)
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")

		}
	}
}
