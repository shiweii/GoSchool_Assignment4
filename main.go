package main

import (
	bst "GoSchool_Assignment4/binarysearchtree"
	dll "GoSchool_Assignment4/doublylinkedlist"
	util "GoSchool_Assignment4/utility"
	"html/template"
	"log"
	"net/http"
	"os"
)

var (
	tpl         *template.Template
	mapSessions = map[string]string{}
	fm          = template.FuncMap{
		"addOne":           util.AddOne,
		"getDay":           util.GetDay,
		"formatDate":       util.FormatDate,
		"firstCharToUpper": util.FirstCharToUpper,
		"urlEncode":        util.UrlEncode,
	}
)

func init() {
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/*"))
}

func main() {

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

	// Handler functions
	http.HandleFunc("/", indexHandler(&userList))
	http.HandleFunc("/signup", signupHandler(&userList))
	http.HandleFunc("/login", loginHandler(&userList))
	http.HandleFunc("/logout", logoutHandler(&userList))
	http.HandleFunc("/landing", landingHandler(&userList))
	http.Handle("/favicon.ico", http.NotFoundHandler())

	// Appointment
	http.HandleFunc("/appointments", appointmentListHandler(&userList, &appointmentSessionList, &appointmentTree))
	http.HandleFunc("/appointment_create", appointmentCreateHandler(&userList))
	http.HandleFunc("/appointment_create_2", appointmentCreatePart2Handler(&userList, &appointmentSessionList, &appointmentTree))
	http.HandleFunc("/appointment_create_confirm", appointmentCreateConfirmHandler(&userList, &appointmentSessionList, &appointmentTree))
	http.HandleFunc("/appointment_edit", appointmentEditHandler(&userList, &appointmentSessionList, &appointmentTree))
	http.HandleFunc("/appointment_edit_confirm", appointmentEditConfirmHandler(&userList, &appointmentSessionList, &appointmentTree))
	http.HandleFunc("/appointment_delete", appointmentDeleteHandler(&userList, &appointmentSessionList, &appointmentTree))
	http.HandleFunc("/appointments_search", appointmentSearchHandler(&userList, &appointmentSessionList, &appointmentTree))

	// User
	http.HandleFunc("/users", usersHandler(&userList))
	http.HandleFunc("/user_edit", userEditHandler(&userList))
	http.HandleFunc("/user_delete", userDeleteHandler(&userList))

	// Admin
	http.HandleFunc("/sessions", sessionListHandler(&userList))

	// Start Server
	http.ListenAndServe(":5221", nil)

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
			http.Redirect(res, req, "/landing", http.StatusSeeOther)
		}
		tpl.ExecuteTemplate(res, "index.gohtml", myUser)
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
		tpl.ExecuteTemplate(res, "landing.gohtml", myUser)
	}
}
