package bot

import (
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
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

// IndexHandler returns data to the default port
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
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

func (s *Slack) postHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
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
	case "good":
		w.Write([]byte("good!"))
	case "neutral":
		w.Write([]byte("neutral!"))
	case "sad":
		w.Write([]byte("sad!"))
	case "SetupYes":
		w.Write([]byte(":white_check_mark: - Starting the setup of your user."))
		s.initBot(userid, email, fullname)
		s.askWhoIsManager(channelid, userid)
	case "SetupNo":
		w.Write([]byte("No worries, let me know if you want to later on!"))
	case "RemoveYes":
		s.removeBot(userid, fullname)
		w.Write([]byte("Sorry to see you go. Your user was deleted."))
	case "RemoveNo":
		w.Write([]byte("Glad you decided to stay :smiley:"))
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
	case "isManagerYes":
		answer := fmt.Sprintf(payload.Actions[0].Value)
		if answer == "isManagerYes" {
			value := fmt.Sprintf("true")
			s.setupIsManager(userid, fullname, value)
			w.Write([]byte(fmt.Sprintf(":white_check_mark: - %s your user is now setup!", displayname)))
		}
	}
	w.WriteHeader(http.StatusOK)
}

// NewHandler instantiaties the web handler for listening on the API
func (s *Slack) NewHandler() (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(middleware.NoCache)
	r.Use(middleware.Heartbeat("/ping"))

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	r.Get("/", indexHandler)
	r.Post("/", s.postHandler)

	r.Get("/debug/pprof/*", pprof.Index)
	r.Get("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		first := true
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "{\n")
		expvar.Do(func(kv expvar.KeyValue) {
			if !first {
				fmt.Fprintf(w, ",\n")
			}
			first = false
			fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
		})
		fmt.Fprintf(w, "\n}\n")
	})

	return r, nil
}
