package bot

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

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

	r.Get("/slack", slackHandler)
	r.Post("/slack", s.slackPostHandler)
	r.Get("/api/happiness/results/date/*/*", surveyResultsUserDayHandler)
	r.Get("/api/happiness/results/all/user/*", surveyResultsUserAllHandler)
	r.Get("/api/happiness/results/dates/*/*/*", surveyResultsUserBetweenDatesHandler)
	r.Get("/api/happiness/results/dates/all/*/*", surveyResultsAllUserBetweenDatesHandler)
	r.Get("/api/happiness/results/all", surveyResultsAllHandler)
	r.Get("/api/slack/users/all", getAllUsersHandler)
	r.Get("/api/slack/users/*", getUserHandler)

	return r, nil
}
