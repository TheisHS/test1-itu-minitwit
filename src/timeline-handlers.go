package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	promtailClient := getPromtailClient("timelineHandler")
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)
	if loggedInUser == nil {
		http.Redirect(w, r, "/public_timeline", http.StatusFound)
		return
	}

	requestURL := fmt.Sprintf("%s/msgsMy/%s?no=%d", serverEndpoint, loggedInUser.Username, perPage)
	res, err := http.Get(requestURL)
	if err != nil {
		message := fmt.Sprintf("error making http request to %s: %s\n", requestURL, err)
		promtailClient.Errorf(message)
		fmt.Print(message)
		totalErrors.Inc()
		http.Error(w, message, http.StatusNotFound)
	}
	posts := messageToPost(res)
	promtailClient.Infof("Successfully retrieved %d posts for user %s", len(posts), loggedInUser.Username)

	data := TimelinePageData{
		User: loggedInUser,
		Posts: posts,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}


func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	promtailClient := getPromtailClient("publicTimelineHandler")
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)

	requestURL := fmt.Sprintf("%s/msgs?no=%d", serverEndpoint, perPage)
	res, err := http.Get(requestURL)
	if err != nil {
		message := fmt.Sprintf("error making http request to %s: %s\n", requestURL, err)
		promtailClient.Errorf(message)
		fmt.Print(message)
		totalErrors.Inc()
		http.Error(w, message, http.StatusNotFound)
	}
	posts := messageToPost(res)
	promtailClient.Infof("Successfully retrieved %d posts for user %s", len(posts), loggedInUser.Username)


	data := TimelinePageData{
		User: loggedInUser,
		Posts: posts,
		IsPublic: true,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}


func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
	promtailClient := getPromtailClient("userTimelineHandler")
	session, _ := store.Get(r, "session")
  vars := mux.Vars(r)
	profileUsername := vars["username"]
	profileUser, _ := getUserFromUsername(profileUsername)
	if profileUser.Username != profileUsername {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	followed := false

	loggedInUser := getLoggedInUser(r)
	if loggedInUser != nil {
		requestURLa := fmt.Sprintf("%s/fllws/%s/%s", serverEndpoint, loggedInUser.Username, profileUsername)
		fllwsRes, err := http.Get(requestURLa)
		if err != nil {
			message := fmt.Sprintf("error making http request to %s: %s\n", requestURLa, err)
			promtailClient.Errorf(message)
			fmt.Print(message)
			totalErrors.Inc()
			http.Error(w, message, http.StatusNotFound)
		}
		fllwsBody, _ := io.ReadAll(fllwsRes.Body)
		followed, err = strconv.ParseBool(string(fllwsBody))
		if err != nil {
			fmt.Printf("could not convert %s to boolean, returning false instead: %s.\n", string(fllwsBody), err)
			followed = false
		}
	}

	requestURLb := fmt.Sprintf("%s/msgs/%s?no=%d", serverEndpoint, profileUsername, perPage)
	res, err := http.Get(requestURLb)
	if err != nil {
		message := fmt.Sprintf("error making http request to %s: %s\n", requestURLb, err)
		promtailClient.Errorf(message)
		fmt.Print(message)
		totalErrors.Inc()
		http.Error(w, message, http.StatusNotFound)
	}
	posts := messageToPost(res)
	promtailClient.Infof("Successfully retrieved %d posts for user %s", len(posts), loggedInUser.Username)

	// for rendering the HTML template
	data := TimelinePageData{
		User: loggedInUser,
		ProfileUser: profileUsername,
		Followed: followed,
		Posts: posts,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}

