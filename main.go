package main

import (
	"context"
	"net/http"

	_ "expvar"
	_ "net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/bot"
	"github.com/matthieudolci/hatcher/database"
)

func main() {
	database.InitDb()
	defer database.DB.Close()

	ctx := context.Background()

	s, err := bot.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}

	if err := s.GetTimeAndUsersForScheduler(); err != nil {
		log.WithError(err)
	}

	handler, err := s.APIHandler()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":9191", handler))
}
