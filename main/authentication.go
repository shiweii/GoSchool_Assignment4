package main

import (
	"net/http"
	"time"

	dll "github.com/shiweii/doublylinkedlist"
	"github.com/shiweii/user"

	uuid "github.com/satori/go.uuid"
)

func authenticationCheck(res http.ResponseWriter, req *http.Request, userList **dll.DoublyLinkedList, checkAdmin bool) (*user.User, bool, int) {
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

func alreadyLoggedIn(req *http.Request, userList **dll.DoublyLinkedList) bool {
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		return false
	}
	username := mapSessions[myCookie.Value]

	ret := (**userList).FindByUsername(username)
	return ret != nil
}

func getUser(res http.ResponseWriter, req *http.Request, userList **dll.DoublyLinkedList) *user.User {
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
		ret := (**userList).FindByUsername(username)
		myUser = ret.(*user.User)
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
