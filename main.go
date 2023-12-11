package main

import (
	"net/http"

	"github.com/xanzy/go-gitlab"
)

const (
	// GitLab URL
	GITLAB_URL = "https://gitlab.xxx.com"

	// GitLab Access Token
	GITLAB_ACCESS_TOKEN = ""

	// GitLab Webhook Secret
	GITLAB_WEBHOOK_SECRET = ""

	// Sever Port
	PORT = "8888"
)

// Create a Webhook server to parse Gitlab events.
func main() {
	// Init logs
	InitLog()

	wh := webhook{
		Secret:         GITLAB_WEBHOOK_SECRET,
		EventsToAccept: []gitlab.EventType{gitlab.EventTypeMergeRequest},
	}

	mux := http.NewServeMux()
	mux.Handle("/", wh)
	logger.Info("HTTP server running on port: " + PORT)
	if err := http.ListenAndServe("0.0.0.0:"+PORT, mux); err != nil {
		logger.Error("HTTP server ListenAndServe: " + err.Error())

	}
}
