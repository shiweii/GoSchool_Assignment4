package main

import (
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shiweii/user"
)

func createNewCookie(id string) *http.Cookie {
	myCookie := &http.Cookie{
		Name:     "myCookie",
		Expires:  time.Now().AddDate(0, 0, 1),
		Value:    id,
		HttpOnly: true,
		Path:     "/",
		Domain:   "localhost",
		Secure:   true,
	}
	return myCookie
}

func authenticationCheck(res http.ResponseWriter, req *http.Request, userList *user.DoublyLinkedList, checkAdmin bool) (*user.User, bool, int) {
	// Check if users is logged in
	if !alreadyLoggedIn(req, userList) {
		// Expire cookie to prevent attacker from reusing this cookie
		myCookie, _ := req.Cookie("myCookie")
		myCookie = &http.Cookie{
			Path:    "/",
			Name:    "myCookie",
			MaxAge:  -1,
			Expires: time.Now().Add(-100 * time.Hour),
		}
		http.SetCookie(res, myCookie)
		return nil, true, http.StatusSeeOther
	}
	// Get info of logged-in user
	myUser := getUser(res, req, userList)
	if myUser == nil {
		return nil, true, http.StatusSeeOther
	}
	// Allow access for admin only
	if checkAdmin {
		if myUser.Role != enumAdmin {
			return nil, true, http.StatusUnauthorized
		}
	}
	return myUser, false, 0
}

func alreadyLoggedIn(req *http.Request, userList *user.DoublyLinkedList) bool {
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		return false
	}
	username := mapSessions[myCookie.Value]
	ret := (*userList).FindByUsername(username)
	return ret != nil
}

func getUser(res http.ResponseWriter, req *http.Request, userList *user.DoublyLinkedList) *user.User {
	// get current session cookie
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		id := uuid.NewV4()
		myCookie = &http.Cookie{
			Name:  "myCookie",
			Value: id.String(),
		}
		http.SetCookie(res, myCookie)
	}
	// if the user exists already, get user
	var myUser *user.User
	if username, ok := mapSessions[myCookie.Value]; ok {
		userObj := (*userList).FindByUsername(username)
		myUser = userObj
	}
	return myUser
}

func killOtherSession(newCookie *http.Cookie) {
	for k, v := range mapSessions {
		sessionID := newCookie.Value
		username := mapSessions[newCookie.Value]
		if k != sessionID && v == username {
			delete(mapSessions, k)
		}
	}
}

func deleteSessionByUsername(username string) {
	for k, v := range mapSessions {
		if v == username {
			delete(mapSessions, k)
		}
	}
}
