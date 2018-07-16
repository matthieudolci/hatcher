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

	// API endpoints to get data from the happiness survey
	router.GET("/api/happiness/userdate/:userid/:date", surveyResultsUserDayHandler)
	router.GET("/api/happiness/userallresults/:userid", surveyResultsUserAllHandler)
	router.GET("/api/happiness/userdates/:userid/:date1/:date2", surveyResultsUserBetweenDatesHandler)
	router.GET("/api/happiness/usersallresults/:date1/:date2", surveyResultsAllUserBetweenDatesHandler)
	router.GET("/api/happiness/all/results", surveyResultsAllHandler)

	//API endpoints to get informations about the slack users
	router.GET("/api/slack/allusers", getAllUsersHandler)
	router.GET("/api/slack/user/:userid", getUserHandler)

	return router, nil
}
