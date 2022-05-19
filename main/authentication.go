package main

import (
	"net/http"
	"time"

	util "github.com/shiweii/utility"

	uuid "github.com/satori/go.uuid"
	"github.com/shiweii/user"
)

// createNewSecureCookie creates and return a new secure cookie
func createNewSecureCookie() *http.Cookie {
	cookie := &http.Cookie{
		Name:     util.GetEnvVar("COOKIE_NAME"),
		Expires:  time.Now().AddDate(0, 0, 1),
		Value:    uuid.NewV4().String(),
		HttpOnly: true,
		Path:     "/",
		Domain:   "localhost",
		Secure:   true,
	}
	return cookie
}

func expireCookie() *http.Cookie {
	cookie := &http.Cookie{
		Path:    "/",
		Name:    util.GetEnvVar("COOKIE_NAME"),
		Domain:  "localhost",
		MaxAge:  -1,
		Expires: time.Now().Add(-100 * time.Hour),
	}
	return cookie
}

func authenticationCheck(res http.ResponseWriter, req *http.Request, userList *user.DoublyLinkedList, checkAdmin bool) (*user.User, bool, int) {
	// Check if users is logged in
	if !alreadyLoggedIn(req, userList) {
		// Expire cookie if user's session was ended by admin
		// to prevent attacker from reusing this cookie
		cookie, err := req.Cookie(util.GetEnvVar("COOKIE_NAME"))
		if err == nil {
			cookie = expireCookie()
			http.SetCookie(res, cookie)
		}
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
	cookie, err := req.Cookie(util.GetEnvVar("COOKIE_NAME"))
	if err != nil {
		return false
	}
	username := mapSessions[cookie.Value]
	ret := (*userList).FindByUsername(username)
	return ret != nil
}

func getUser(res http.ResponseWriter, req *http.Request, userList *user.DoublyLinkedList) *user.User {
	// get current session cookie
	cookie, err := req.Cookie(util.GetEnvVar("COOKIE_NAME"))
	if err != nil {
		id := uuid.NewV4()
		cookie = &http.Cookie{
			Name:   util.GetEnvVar("COOKIE_NAME"),
			Value:  id.String(),
			Domain: "localhost",
		}
		http.SetCookie(res, cookie)
	}
	// if the user exists already, get user
	var myUser *user.User
	if username, ok := mapSessions[cookie.Value]; ok {
		userObj := (*userList).FindByUsername(username)
		myUser = userObj
	}
	return myUser
}

func terminateOtherSession(sessionID, username string) {
	// Loop session Map
	for k, v := range mapSessions {
		// If same user but different session, remove session from session map
		if k != sessionID && v == username {
			delete(mapSessions, k)
		}
	}
}

func deleteSessionByUsername(username string) {
	// Loop session Map
	for k, v := range mapSessions {
		// If map value equals username
		if v == username {
			// Remove session from session map
			delete(mapSessions, k)
		}
	}
}
