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
	// Go routine to verify .env checksum every 5 minutes
	go util.VerifyCheckSum()
	// Go Routine to perform encryption if file was left decrypted due to panic
	go ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
}

func main() {
	logger.Info.Println("Server Start...")

	sigchld := make(chan os.Signal, 1)
	signal.Notify(sigchld, os.Interrupt)
	go func() {
		<-sigchld
		logger.Info.Println("Server Stop...")
		logger.CloseLogger()
		os.Exit(0)
	}()

	// Initialize new doubly linked-list and binary search tree
	var (
		appointmentTree        = bst.New()
		userList               = dll.New()
		appointmentSessionList = dll.New()
	)

	// Initialize Sample Data
	appointmentSessionList.Add(app.AppointmentSession{Num: 1, StartTime: "09:00", EndTime: "10:00", Available: true})
	appointmentSessionList.Add(app.AppointmentSession{Num: 2, StartTime: "10:00", EndTime: "11:00", Available: true})
	appointmentSessionList.Add(app.AppointmentSession{Num: 3, StartTime: "11:00", EndTime: "12:00", Available: true})
	appointmentSessionList.Add(app.AppointmentSession{Num: 4, StartTime: "13:00", EndTime: "14:00", Available: true})
	appointmentSessionList.Add(app.AppointmentSession{Num: 5, StartTime: "14:00", EndTime: "15:00", Available: true})
	appointmentSessionList.Add(app.AppointmentSession{Num: 6, StartTime: "15:00", EndTime: "16:00", Available: true})
	appointmentSessionList.Add(app.AppointmentSession{Num: 7, StartTime: "16:00", EndTime: "17:00", Available: true})

	// Loading Data from JSON
	users := user.GetEncryptedUserData()
	for _, userObj := range users {
		userList.Add(userObj)
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

	err := http.ListenAndServe(util.GetEnvVar("PORT"), router)
	if err != nil {
		logger.Fatal.Fatalln(err)
	}
}
