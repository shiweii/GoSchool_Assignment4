package main

import (
	dll "GoSchool_Assignment4/doublylinkedlist"
	"log"
	"net/http"
	"strconv"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func authenticationCheck(res http.ResponseWriter, req *http.Request, userList **dll.DoublyLinkedlist, checkAdmin bool) (*User, bool, int) {
	// Check if users is logged in
	if !alreadyLoggedIn(req, userList) {
		return nil, true, http.StatusSeeOther
	}
	// Get info of logged in user
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

func sessionListHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		_, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		type SessionStrcut struct {
			SessionID string
			Username  string
			Role      string
		}

		sessions := []SessionStrcut{}

		for k, v := range mapSessions {
			user := (**userList).FindByUsername(v).(*User)
			sessions = append(sessions, SessionStrcut{SessionID: k, Username: v, Role: user.Role})
		}

		// Process form submission
		if req.Method == http.MethodPost {
			req.ParseForm()
			for key, values := range req.Form {
				for _, value := range values {
					if key == "sessionsDel" {
						delete(mapSessions, value)
					}
				}
			}
			http.Redirect(res, req, "/sessions", http.StatusSeeOther)
		}

		err := tpl.ExecuteTemplate(res, "sessions.gohtml", sessions)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func signupHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}
		var myUser User
		// process form submission
		if req.Method == http.MethodPost {
			// get form values
			username := req.FormValue("username")
			password := req.FormValue("password")
			firstname := req.FormValue("firstname")
			lastname := req.FormValue("lastname")
			mobileNumber := req.FormValue("mobileNum")
			mobileNum, _ := strconv.Atoi(mobileNumber)

			if username != "" {
				// check if username exist/ taken
				userItf := (**userList).FindByUsername(username)
				if userItf != nil {
					http.Error(res, "Username already taken", http.StatusForbidden)
					return
				}
				// create session
				id := uuid.NewV4()
				myCookie := &http.Cookie{
					Name:  "myCookie",
					Value: id.String(),
				}
				http.SetCookie(res, myCookie)
				mapSessions[myCookie.Value] = username

				bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
				if err != nil {
					http.Error(res, "Internal server error", http.StatusInternalServerError)
					return
				}

				myUser.Username = username
				myUser.Password = string(bPassword)
				myUser.FirstName = firstname
				myUser.LastName = lastname
				myUser.Role = enumPatient
				myUser.MobileNumber = mobileNum

				// Add into linklist and JSON
				(**userList).Add(&myUser)
				(**userList).InsertionSort()
				addUserDate(&myUser)
			}
			// redirect to patient landing page
			http.Redirect(res, req, "/patient", http.StatusSeeOther)
			return

		}
		tpl.ExecuteTemplate(res, "signup.gohtml", myUser)
	}
}

func loginHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}

		// process form submission
		if req.Method == http.MethodPost {
			username := req.FormValue("username")
			password := req.FormValue("password")

			// check if user exist with username
			userItf := (**userList).FindByUsername(username)
			if userItf == nil {
				http.Error(res, "Username and/or password do not match", http.StatusUnauthorized)
				return
			}
			user := userItf.(*User)
			// Check if user is deleted
			if user.IsDeleted {
				http.Error(res, "Username and/or password do not match", http.StatusUnauthorized)
				return
			}
			// Matching of password entered
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				http.Error(res, "Username and/or password do not match", http.StatusForbidden)
				return
			}
			id := uuid.NewV4()
			myCookie := &http.Cookie{
				Name:  "myCookie",
				Value: id.String(),
			}
			http.SetCookie(res, myCookie)
			mapSessions[myCookie.Value] = user.Username
			http.Redirect(res, req, "/landing", http.StatusSeeOther)
			return
		}
		tpl.ExecuteTemplate(res, "login.gohtml", nil)
	}
}

func logoutHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}
		myCookie, _ := req.Cookie("myCookie")
		// delete the session
		delete(mapSessions, myCookie.Value)
		// remove the cookie
		myCookie = &http.Cookie{
			Name:   "myCookie",
			Value:  "",
			MaxAge: -1,
		}
		http.SetCookie(res, myCookie)

		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func alreadyLoggedIn(req *http.Request, userList **dll.DoublyLinkedlist) bool {
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		return false
	}
	username := mapSessions[myCookie.Value]

	ret := (**userList).FindByUsername(username)
	return ret != nil
}

func getUser(res http.ResponseWriter, req *http.Request, userList **dll.DoublyLinkedlist) *User {
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
	var myUser *User
	if username, ok := mapSessions[myCookie.Value]; ok {
		ret := (**userList).FindByUsername(username)
		myUser = ret.(*User)
	}

	return myUser
}

func updateCookie(res http.ResponseWriter, req *http.Request, username string) {
	// get current session cookie
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		id := uuid.NewV4()
		myCookie = &http.Cookie{
			Name:  "myCookie",
			Value: id.String(),
		}

	}
	http.SetCookie(res, myCookie)

	// if the user exists already, update cookie username
	mapSessions[myCookie.Value] = username
}
