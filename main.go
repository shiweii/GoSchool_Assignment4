package main

import (
	bst "GoSchool_Assignment4/binarysearchtree"
	dll "GoSchool_Assignment4/doublylinkedlist"
	"GoSchool_Assignment4/logger"
	util "GoSchool_Assignment4/utility"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	tpl         *template.Template
	mapSessions = map[string]string{}
	fm          = template.FuncMap{
		"addOne":           util.AddOne,
		"getDay":           util.GetDay,
		"formatDate":       util.FormatDate,
		"firstCharToUpper": util.FirstCharToUpper,
	}
)

func init() {
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/*"))

	s3Bucket := util.GetEnvVar("S3_BUCKET")
	secretKey := util.GetEnvVar("SECRET_KEY")

	fmt.Println(s3Bucket, secretKey)
}

func main() {

	logger.Info.Println("Server Start...")

	//go logger.Checksum()
	// encrypt decrypt file

	// Initialize new doubly linkedlist and binary search tree
	var (
		appointmentTree        = bst.New()
		userList               = dll.New()
		appointmentSessionList = dll.New()
	)

	// Initialize Sample Data
	appointmentSessionList.Add(AppointmentSession{1, "09:00", "10:00", true})
	appointmentSessionList.Add(AppointmentSession{2, "10:00", "11:00", true})
	appointmentSessionList.Add(AppointmentSession{3, "11:00", "12:00", true})
	appointmentSessionList.Add(AppointmentSession{4, "13:00", "14:00", true})
	appointmentSessionList.Add(AppointmentSession{5, "14:00", "15:00", true})
	appointmentSessionList.Add(AppointmentSession{6, "15:00", "16:00", true})
	appointmentSessionList.Add(AppointmentSession{7, "16:00", "17:00", true})

	// Loading Data from JSON
	users := getUserData()
	for _, user := range users {
		userList.Add(user)
	}
	userList.InsertionSort()

	appointments := getAppointmentData()
	for _, v := range appointments {
		patient := userList.FindByUsername(v.Patient)
		dentist := userList.FindByUsername(v.Dentist)
		appointmentTree.Add(v.ID, v.Date, v.Session, dentist, patient)
	}

	router := mux.NewRouter()

	// Handler functions
	router.HandleFunc("/", indexHandler(&userList))
	router.HandleFunc("/signup", signupHandler(&userList))
	router.HandleFunc("/login", loginHandler(&userList))
	router.HandleFunc("/logout", logoutHandler(&userList))
	router.HandleFunc("/landing", landingHandler(&userList))
	router.Handle("/favicon.ico", http.NotFoundHandler())

	// Appointment
	router.HandleFunc("/appointments", appointmentListHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc("/appointments/search", appointmentSearchHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc("/appointment/create", appointmentCreateHandler(&userList))
	router.HandleFunc("/appointment/create/{dentist}", appointmentCreatePart2Handler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc(`/appointment/create/{dentist}/{date:\d{4}-\d{2}-\d{2}}/{session:[0-9]+}`, appointmentCreateConfirmHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc("/appointment/edit/{id:[0-9]+}", appointmentEditHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc(`/appointment/edit/{id:[0-9]+}/{dentist}/{date:\d{4}-\d{2}-\d{2}}/{session:[0-9]+}`, appointmentEditConfirmHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc("/appointment/delete/{id:[0-9]+}", appointmentDeleteHandler(&userList, &appointmentSessionList, &appointmentTree))

	// User
	router.HandleFunc("/users", userListHandler(&userList))
	router.HandleFunc("/user/edit/{username}", userEditHandler(&userList))
	router.HandleFunc("/user/delete/{username}", userDeleteHandler(&userList))

	// Admin
	router.HandleFunc("/sessions", sessionListHandler(&userList))

	http.ListenAndServe(portNum, router)
}

func indexHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
				Error.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		myUser := getUser(res, req, userList)
		if alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/appointments", http.StatusSeeOther)
			return
		}

		ViewData := struct {
			LoggedInUser *User
			PageTitle    string
		}{
			myUser,
			"Central City Dentist Clinic",
		}

		tpl.ExecuteTemplate(res, "index.gohtml", ViewData)
	}
}

func landingHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
				Error.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()
		myUser, authFail, httpStatusNum := authenticationCheck(res, req, userList, false)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		ViewData := struct {
			LoggedInUser *User
			PageTitle    string
		}{
			myUser,
			"Homepage",
		}

		tpl.ExecuteTemplate(res, "landing.gohtml", ViewData)
	}
}
