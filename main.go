package main

import (
	"context"
	"log"
	"net/http"
	"os"

	_ "expvar"
	_ "net/http/pprof"

	"github.com/matthieudolci/hatcher/bot"
	"github.com/matthieudolci/hatcher/database"
)

func main() {
	database.InitDb()
	defer database.DB.Close()

	lg := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	ctx := context.Background()

	s, err := bot.New()
	if err != nil {
		lg.Fatal(err)
	}
	s.Logger = lg

	if err := s.Run(ctx); err != nil {
		s.Logger.Fatal(err)
	}

	handler, err := s.ApiHandler()
	if err != nil {
		s.Logger.Fatal(err)
	}

	lg.Fatal(http.ListenAndServe(":9191", handler))
}
