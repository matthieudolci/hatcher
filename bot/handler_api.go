package bot

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// APIHandler instantiaties the web handler for listening on the API
func (s *Slack) APIHandler() (http.Handler, error) {

	router := httprouter.New()

	router.POST("/slack", s.slackPostHandler)
	router.GET("/api/happiness/userdate/:userid/:date", surveyResultsUserDayHandler)
	router.GET("/api/happiness/userallresults/:userid", surveyResultsUserAllHandler)
	router.GET("/api/happiness/userdates/:userid/:date1/:date2", surveyResultsUserBetweenDatesHandler)
	router.GET("/api/happiness/usersallresults/:date1/:date2", surveyResultsAllUserBetweenDatesHandler)
	router.GET("/api/happiness/all/results", surveyResultsAllHandler)
	router.GET("/api/slack/allusers", getAllUsersHandler)
	router.GET("/api/slack/user/:userid", getUserHandler)
	return router, nil
}
