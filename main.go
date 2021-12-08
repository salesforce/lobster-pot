package main

import (
	"net/http"
	"os"

	"github.com/heroku/lobster-pot/config"
	"github.com/heroku/lobster-pot/handlers"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func main() {

	var err error

	err = config.Init()
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	c, err := config.BuildAppsConfig()
	if err != nil {
		log.Fatal(err)
	}

	// setup the worker for posting to slack
	handlers.StartQueueWorker(c)

	http.Handle("/",
		handlers.AuthCheck(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) { handlers.GithubWebhookHandler(w, r, c) },
			),
		),
	)
	http.Handle("/hook",
		handlers.AuthCheck(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) { handlers.GithubWebhookHandler(w, r, c) },
			),
		),
	)

	//handler for slack callbacks
	http.Handle("/slack", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { handlers.SlackCallback(w, r, c) },
	))
	log.Debug("Starting server on port: ", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}

}
