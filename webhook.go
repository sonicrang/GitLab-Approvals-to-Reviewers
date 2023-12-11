package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/xanzy/go-gitlab"
)

// webhook is a HTTP Handler for Gitlab Webhook events.
type webhook struct {
	Secret         string
	EventsToAccept []gitlab.EventType
}

// ServeHTTP tries to parse Gitlab events sent and calls handle function
// with the successfully parsed events.
func (hook webhook) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	event, err := hook.parse(request)
	if err != nil {
		writer.WriteHeader(500)
		logger.Error("could parse the webhook event: " + err.Error())
		return
	}

	// Handle the event before we return.
	if err := hook.addApprovalToReviwer(event); err != nil {
		writer.WriteHeader(500)
		logger.Error("error handling the event: " + err.Error())
		return
	}

	// Write a response when were done.
	writer.WriteHeader(204)
}

func (hook webhook) addApprovalToReviwer(event interface{}) error {
	str, err := json.Marshal(event)
	if err != nil {
		return errors.New("could not marshal json event for logging: " + err.Error())
	}

	// Just write the event for this example.
	logger.Info("handle event:\n" + string(str))

	mr := &gitlab.MergeEvent{}
	if err := json.Unmarshal(str, mr); err != nil {
		return errors.New("could not convert event to merge_request struct: " + err.Error())
	}

	// get approval in mr via gitlab api
	git := InitAPI()
	ApprovalstoReviewer(git, mr.Project.ID, mr.ObjectAttributes.IID)
	return nil
}

// parse verifies and parses the events specified in the request and
// returns the parsed event or an error.
func (hook webhook) parse(r *http.Request) (interface{}, error) {
	defer func() {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			logger.Error("could discard request body: " + err.Error())
		}
		if err := r.Body.Close(); err != nil {
			logger.Error("could not close request body: " + err.Error())
		}
	}()

	if r.Method != http.MethodPost {
		return nil, errors.New("invalid HTTP Method")
	}

	// If we have a secret set, we should check if the request matches it.
	if len(hook.Secret) > 0 {
		signature := r.Header.Get("X-Gitlab-Token")
		if signature != hook.Secret {
			return nil, errors.New("token validation failed")
		}
	}

	event := r.Header.Get("X-Gitlab-Event")
	if strings.TrimSpace(event) == "" {
		return nil, errors.New("missing X-Gitlab-Event Header")
	}

	eventType := gitlab.EventType(event)
	if !isEventSubscribed(eventType, hook.EventsToAccept) {
		return nil, errors.New("event not defined to be parsed")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, errors.New("error reading request body")
	}

	return gitlab.ParseWebhook(eventType, payload)
}

func isEventSubscribed(event gitlab.EventType, events []gitlab.EventType) bool {
	for _, e := range events {
		if event == e {
			return true
		}
	}
	return false
}
