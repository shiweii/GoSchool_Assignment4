package main

import (
	"html/template"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	app "github.com/shiweii/appointment"
	bst "github.com/shiweii/binarysearchtree"
	dll "github.com/shiweii/doublylinkedlist"
	"github.com/shiweii/logger"
	"github.com/shiweii/user"
	util "github.com/shiweii/utility"
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
	defer func() {
		if err := recover(); err != nil {
			logger.Panic.Println(err)
		}
	}()

	// Go routine to verify checksum every 10 minutes
	go util.VerifyCheckSum()

	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/*"))

	// Go Routine to perform encryption if file was left decrypted due to panic
	util.CheckEncryption()
}

func main() {
	logger.Info.Println("[Server Start]")

	// Channel to detect ctrl-c and exit the server gracefully
	sigchld := make(chan os.Signal, 1)
	signal.Notify(sigchld, os.Interrupt)
	go func() {
		<-sigchld
		logger.Info.Println("[Server Stop]")
		logger.CloseLogger()
		os.Exit(0)
	}()

	// Initialize new doubly linked-list and binary search tree
	var (
		appointmentTree        = app.BinarySearchTree{BinarySearchTree: bst.New()}
		userList               = user.DoublyLinkedList{DoublyLinkedList: dll.New()}
		appointmentSessionList = dll.New()
	)

	// Initialize Sample Data
	appointmentSessionList.Add(app.AppSession{Num: 1, StartTime: "09:00", EndTime: "10:00", Available: true})
	appointmentSessionList.Add(app.AppSession{Num: 2, StartTime: "10:00", EndTime: "11:00", Available: true})
	appointmentSessionList.Add(app.AppSession{Num: 3, StartTime: "11:00", EndTime: "12:00", Available: true})
	appointmentSessionList.Add(app.AppSession{Num: 4, StartTime: "13:00", EndTime: "14:00", Available: true})
	appointmentSessionList.Add(app.AppSession{Num: 5, StartTime: "14:00", EndTime: "15:00", Available: true})
	appointmentSessionList.Add(app.AppSession{Num: 6, StartTime: "15:00", EndTime: "16:00", Available: true})
	appointmentSessionList.Add(app.AppSession{Num: 7, StartTime: "16:00", EndTime: "17:00", Available: true})

	users := user.GetEncryptedUserData()
	for _, userObj := range users {
		userList.Add(userObj)
	}
	userList.InsertionSort()

	appointments := app.GetAppointmentData()
	for _, v := range appointments {
		appointment := app.New(v.ID, userList.FindByUsername(v.Patient.(string)), userList.FindByUsername(v.Dentist.(string)), v.Date, v.Session)
		appointmentTree.Add(v.Date, appointment)
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
	router.HandleFunc(`/appointment/create/{dentist}/{date:\d{4}-\d{2}-\d{2}}/{session:[1-7]+}`, appointmentCreateConfirmHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc("/appointment/edit/{id:[0-9]+}", appointmentEditHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc(`/appointment/edit/{id:[0-9]+}/{dentist}/{date:\d{4}-\d{2}-\d{2}}/{session:[1-7]+}`, appointmentEditConfirmHandler(&userList, &appointmentSessionList, &appointmentTree))
	router.HandleFunc("/appointment/delete/{id:[0-9]+}", appointmentDeleteHandler(&userList, &appointmentSessionList, &appointmentTree))

	// User
	router.HandleFunc("/users", userListHandler(&userList))
	router.HandleFunc("/user/edit/{username}", userEditHandler(&userList))
	router.HandleFunc("/user/delete/{username}", userDeleteHandler(&userList))

	// Admin
	router.HandleFunc("/sessions", sessionListHandler(&userList))

	if err := http.ListenAndServeTLS(util.GetEnvVar("PORT"), util.GetEnvVar("SSL_CERT"), util.GetEnvVar("SSL_KEY"), router); err != nil {
		logger.Fatal.Fatalln("ListenAndServe: ", err)
	}
}
