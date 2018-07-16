package main

import (
	"context"
	"net/http"

	"github.com/matthieudolci/hatcher/common"

	_ "expvar"
	_ "net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/matthieudolci/hatcher/api"
	"github.com/matthieudolci/hatcher/bot"
	"github.com/matthieudolci/hatcher/database"
	"github.com/matthieudolci/hatcher/scheduler"
)

func main() {

	var s *common.Slack

	database.InitDb()
	defer database.DB.Close()

	ctx := context.Background()

	s, err := bot.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := bot.Run(ctx, s); err != nil {
		log.Fatal(err)
	}

	if err := scheduler.GetTimeAndUsersForScheduler(s); err != nil {
		log.WithError(err)
	}

	handler, err := api.Handler(s)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":9191", handler))
}
