package main

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)


func registerHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)
	if loggedInUser != nil {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}

	var registerError string
	isPostRequest := r.Method == http.MethodPost

	defer func() {
		if isPostRequest && registerError == "" { return }
		data := LoginPageData{
			Error: registerError,
			Flashes: session.Flashes(),
			Endpoint: serverEndpoint,
		}
		
		session.Save(r, w)
		registerTmpl.Execute(w, data)
	}()
	
	if isPostRequest {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		if password != password2 {
			registerError = "The two passwords do not match"
			return
		} 

		body := []byte(fmt.Sprintf(`{
			"username": "%s",
			"email": "%s",
			"pwdHash": "%s"
		}`, username, email, GeneratePasswordHash(password)))

		requestURL := fmt.Sprintf("%s/register", serverEndpoint)
		res, err := http.Post(requestURL, "application/json" ,bytes.NewBuffer(body))
		if err != nil {
			fmt.Println(err)
			totalErrors.Inc()
			registerError = "Server error"
			return
		}

		var resErr ServerError
		json.NewDecoder(res.Body).Decode(&resErr)

		if resErr.ErrorMsg != "" {
			registerError = resErr.ErrorMsg
			return
		}

		session.AddFlash("You were successfully registered and can login now")
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}


func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)
	if loggedInUser != nil {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}
	
	var loginError string
	isPostRequest := r.Method == http.MethodPost

	defer func() {
		if isPostRequest && loginError == "" { return }
		data := LoginPageData{
			Error: loginError,
			Flashes: session.Flashes(),
			Endpoint: serverEndpoint,
		}
		
		session.Save(r, w)
		loginTmpl.Execute(w, data)
	}()
	
	if isPostRequest {
		username := r.FormValue("username")
		password := r.FormValue("password")

		body := []byte(fmt.Sprintf(`{
			"username": "%s"
		}`, username))

		requestURL := fmt.Sprintf("%s/login", serverEndpoint)
		res, err := http.Post(requestURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println(err)
			loginError = "Server error"
			return
		}

		type Salt struct {
			ErrorMsg string
			Salt string
		}
		var saltResponse Salt
		json.NewDecoder(res.Body).Decode(&saltResponse)
		if saltResponse.ErrorMsg != "" {
			loginError = saltResponse.ErrorMsg
			return
		}

		body = []byte(fmt.Sprintf(`{
			"username": "%s",
			"pwdHash": "%s"
		}`, username, hashString(saltResponse.Salt, password)))

		res, err = http.Post(requestURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			totalErrors.Inc()
			fmt.Println(err)
			loginError = "Server error"
			return
		}

		type LoginResponse struct {
			ErrorMsg string
			UserID int
		}
		var loginResponse LoginResponse
		json.NewDecoder(res.Body).Decode(&loginResponse)
		if loginResponse.ErrorMsg != "" {
			loginError = loginResponse.ErrorMsg
			return
		}

		if loginResponse.UserID == 0 {
			fmt.Printf("%v", loginResponse)
			loginError = "Unexpected error"
			return
		}

		session.Values["user_id"] = loginResponse.UserID
		session.AddFlash("You were logged in")
		saveError := session.Save(r, w)
		if saveError != nil {
			totalErrors.Inc()
			fmt.Println(saveError.Error())
			loginError = "Unexpected error"
			return
		}

		loginError = ""
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
	}
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.AddFlash("You were logged out")
	session.Save(r,w)
	http.Redirect(w,r, "/public_timeline", http.StatusSeeOther)
}