package api

import (
	"net/http"

	"github.com/matthieudolci/hatcher/common"

	"github.com/julienschmidt/httprouter"
	"github.com/matthieudolci/hatcher/bot"
)

// Handler instantiaties the web handler for listening on the API
func Handler(s *common.Slack) (http.Handler, error) {

	router := httprouter.New()

	// main API endpoint with slack
	router.POST("/slack", bot.SlackPostHandler(s))

	return router, nil
}
