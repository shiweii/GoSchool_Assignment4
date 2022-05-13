package main

import (
	"html/template"
	"net/http"
	"os"
	"os/signal"

	app "github.com/shiweii/appointment"
	bst "github.com/shiweii/binarysearchtree"
	dll "github.com/shiweii/doublylinkedlist"
	ede "github.com/shiweii/encryptdecrypt"
	"github.com/shiweii/logger"
	"github.com/shiweii/user"
	util "github.com/shiweii/utility"

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
	// Go routine to verify .env checksum every 5 munutes
	go util.VerifyCheckSum()
	// Go Routine to perform encryption if file was left decrypted due to panic
	go ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
}

func main() {
	logger.Info.Println("Server Start...")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		logger.Info.Println("Server Stop...")
		logger.CloseLogger()
		os.Exit(0)
	}()

	// Initialize new doubly linkedlist and binary search tree
	var (
		appointmentTree        = bst.New()
		userList               = dll.New()
		appointmentSessionList = dll.New()
	)

	// Initialize Sample Data
	appointmentSessionList.Add(app.AppointmentSession{1, "09:00", "10:00", true})
	appointmentSessionList.Add(app.AppointmentSession{2, "10:00", "11:00", true})
	appointmentSessionList.Add(app.AppointmentSession{3, "11:00", "12:00", true})
	appointmentSessionList.Add(app.AppointmentSession{4, "13:00", "14:00", true})
	appointmentSessionList.Add(app.AppointmentSession{5, "14:00", "15:00", true})
	appointmentSessionList.Add(app.AppointmentSession{6, "15:00", "16:00", true})
	appointmentSessionList.Add(app.AppointmentSession{7, "16:00", "17:00", true})

	// Loading Data from JSON
	users := user.GetEncryptedUserData()
	for _, user := range users {
		userList.Add(user)
	}
	userList.InsertionSort()

	appointments := app.GetAppointmentData()
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

	http.ListenAndServe(util.GetEnvVar("PORT"), router)
}
