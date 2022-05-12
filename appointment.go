package main

import (
	bst "GoSchool_Assignment4/binarysearchtree"
	dll "GoSchool_Assignment4/doublylinkedlist"
	util "GoSchool_Assignment4/utility"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Appointment struct {
	ID      int    `json:"id"`
	Patient string `json:"patient"`
	Dentist string `json:"dentist"`
	Date    string `json:"date"`
	Session int    `json:"session"`
}

type AppointmentSession struct {
	Num       int
	StartTime string
	EndTime   string
	Available bool
}

func newAppointment(id int, pateint, dentist, date string, session int) *Appointment {
	return &Appointment{
		ID:      id,
		Patient: pateint,
		Dentist: dentist,
		Date:    date,
		Session: session,
	}
}

func getAppointmentData() []*Appointment {
	var appointments []*Appointment
	JSONData, _ := ioutil.ReadFile(util.GetEnvVar("APPOINTMENT_DATA"))
	err := json.Unmarshal(JSONData, &appointments)
	if err != nil {
		fmt.Println(err)
	}
	return appointments
}

func addAppointmentData(a *Appointment) {
	var appointments []*Appointment
	appointments = getAppointmentData()
	appointments = append(appointments, a)
	JSONData, _ := json.MarshalIndent(appointments, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("APPOINTMENT_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func updateAppointmentData(oldAppointment *Appointment, editiedAppointment *Appointment) {
	var appointments []*Appointment = getAppointmentData()
	for k, v := range appointments {
		if reflect.DeepEqual(v, oldAppointment) {
			appointments[k] = editiedAppointment
		}
	}
	JSONData, _ := json.MarshalIndent(appointments, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("APPOINTMENT_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func deleteAppointmentData(id int) {
	var appointments []*Appointment
	var idx int
	appointments = getAppointmentData()
	for k, v := range appointments {
		if v.ID == id {
			idx = k
		}
	}
	appointments = remove(appointments, idx)
	JSONData, _ := json.MarshalIndent(appointments, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("APPOINTMENT_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func remove(slice []*Appointment, s int) []*Appointment {
	return append(slice[:s], slice[s+1:]...)
}

func appointmentListHandler(userList, appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		var appointments []*bst.BinaryNode
		viewReq := req.FormValue("view")
		dt := time.Now()

		// If logged in as admin, display all appointments
		if myUser.Role == enumAdmin {
			appointments = (**appointmentTree).GetAllAppointments(nil, "")
		}

		// If logged in as patient, display appointments based on selection
		if myUser.Role == enumPatient {
			if len(viewReq) == 0 {
				viewReq = enumUpcoming
			}
			if viewReq == enumUpcoming {
				appointments = (**appointmentTree).GetUpComingAppointments(myUser, "patient")
			} else {
				appointments = (**appointmentTree).GetAllAppointments(myUser, "patient")
			}
		}

		sessions := (**appointmentSessionList).GetList()
		users := (**userList).GetList()
		dentists := getDentistList(users)

		ViewData := struct {
			LoggedInUser *User
			PageTitle    string
			CurrentPage  string
			Appointments []*bst.BinaryNode
			Sessions     []interface{}
			Dentists     []*User
			Option       string
			TodayDate    string
		}{
			myUser,
			"Appointments",
			"MA",
			appointments,
			sessions,
			dentists,
			viewReq,
			dt.Format("2006-01-02"),
		}

		// Process form submission
		if req.Method == http.MethodPost {
			inputDentist := req.FormValue("inputDentist")
			inputDate := req.FormValue("inputDate")
			inputPatientMobileNumber := req.FormValue("inputPatientMobileNumber")
			inputSession := req.FormValue("inputSession")

			// Data covertion
			dentist := (**userList).FindByUsername(inputDentist)
			appointmentDate, _ := time.Parse("2006-01-02", inputDate)
			appointmentSession, _ := strconv.Atoi(inputSession)
			patientMobileNumber, _ := strconv.Atoi(inputPatientMobileNumber)

			// If inputs are valid
			if !(dentist == nil && len(inputDate) == 0 && appointmentSession == 0 && len(inputPatientMobileNumber) == 0) {

				// Initialize channels
				chSearchDate := make(chan []*bst.BinaryNode)
				chSearchPatient := make(chan []*bst.BinaryNode)
				chSearchDentist := make(chan []*bst.BinaryNode)
				chSearchSession := make(chan []*bst.BinaryNode)
				filterCount := 0

				if dentist != nil {
					filterCount++
					go (**appointmentTree).SearchAllByField(enumDentist, dentist, chSearchDentist)
				}
				if len(inputDate) > 0 {
					filterCount++
					go (**appointmentTree).SearchAllByField("date", appointmentDate.Format("2006-01-02"), chSearchDate)
				}
				if len(inputPatientMobileNumber) > 0 {
					filterCount++
					patient := (**userList).SearchByMobileNumber(patientMobileNumber)
					go (**appointmentTree).SearchAllByField(enumPatient, patient, chSearchPatient)
				}
				if appointmentSession > 0 {
					filterCount++
					go (**appointmentTree).SearchAllByField("session", appointmentSession, chSearchSession)
				}

				var result []*bst.BinaryNode
				for i := 0; i < filterCount; i++ {
					select {
					case ret := <-chSearchDate:
						result = append(result, ret...)
					case ret2 := <-chSearchPatient:
						result = append(result, ret2...)
					case ret3 := <-chSearchDentist:
						result = append(result, ret3...)
					case ret4 := <-chSearchSession:
						result = append(result, ret4...)
					}
				}
				ViewData.Appointments = getDup(result, filterCount)
			}
		}
		tpl.ExecuteTemplate(res, "appointmentList.gohtml", ViewData)
	}
}

// Function to get duplicates from search results
func getDup(list []*bst.BinaryNode, count int) []*bst.BinaryNode {

	var temp []*bst.BinaryNode
	duplicate_frequency := make(map[*bst.BinaryNode]int)

	for _, item := range list {
		// check if the item/element exist in the duplicate_frequency map
		_, exist := duplicate_frequency[item]
		if exist {
			// increase counter by 1 if already in the map
			duplicate_frequency[item] += 1
		} else {
			// else start counting from 1
			duplicate_frequency[item] = 1
		}
	}
	for v, n := range duplicate_frequency {
		if n == count {
			temp = append(temp, v)
		}
	}
	return temp
}

func appointmentCreateHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		myUser, authFail, httpStatusNum := authenticationCheck(res, req, userList, false)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		users := (**userList).GetList()
		ViewData := struct {
			LoggedInUser *User
			PageTitle    string
			CurrentPage  string
			Dentists     []*User
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			getDentistList(users),
		}

		tpl.ExecuteTemplate(res, "appointmentCreate_step1.gohtml", ViewData)
	}
}

func appointmentCreatePart2Handler(userList, appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		var sessionList []AppointmentSession
		dt := time.Now()

		ViewData := struct {
			LoggedInUser *User
			PageTitle    string
			CurrentPage  string
			Dentist      *User
			TodayDate    string
			Sessions     []AppointmentSession
			SelectedDate string
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			nil,
			dt.Format("2006-01-02"),
			sessionList,
			"",
		}

		// Get data from query string
		vars := mux.Vars(req)
		dentistReq := vars["dentist"]
		dentist := (**userList).FindByUsername(dentistReq)

		if dentist != nil {
			ViewData.Dentist = dentist.(*User)
		}

		// Process form submission
		if req.Method == http.MethodPost {
			inputDate := req.FormValue("apptDate")
			appointmentDate, err := time.Parse("2006-01-02", inputDate)
			if err == nil {
				ViewData.Sessions = getDentistAvailability(appointmentSessionList, appointmentTree, appointmentDate, ViewData.Dentist)
				ViewData.SelectedDate = appointmentDate.Format("2006-01-02")
			}
		}
		tpl.ExecuteTemplate(res, "appointmentCreate_step2.gohtml", ViewData)
	}
}

func appointmentCreateConfirmHandler(userList, appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		// Get data from query string
		vars := mux.Vars(req)
		dentistReq := vars["dentist"]
		dateReq := vars["date"]
		sessionReq := vars["session"]

		// Data conversion
		appointmentDate, err := time.Parse("2006-01-02", dateReq)
		if err != nil {
			fmt.Println(err)
		}
		ses, _ := strconv.Atoi(sessionReq)

		session := (**appointmentSessionList).Get(ses).(AppointmentSession)
		dentist := (**userList).FindByUsername(dentistReq)

		ViewData := struct {
			LoggedInUser  *User
			PageTitle     string
			CurrentPage   string
			Dentist       *User
			Date          string
			StartTime     string
			EndTime       string
			Successful    bool
			FormSubmitted bool
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			dentist.(*User),
			appointmentDate.Format("2006-01-02"),
			session.StartTime,
			session.EndTime,
			false,
			false,
		}

		// Process form submission
		if req.Method == http.MethodPost {
			var id int = util.GenerateID()
			chn := make(chan bool)
			go createNewAppointment(id, appointmentDate.Format("2006-01-02"), session.Num, dentist, myUser, appointmentTree, chn)
			successful := <-chn
			if successful {
				appointment := newAppointment(id, myUser.Username, dentist.(*User).Username, appointmentDate.Format("2006-01-02"), session.Num)
				addAppointmentData(appointment)
				ViewData.Successful = true
			}
			ViewData.FormSubmitted = true
		}

		tpl.ExecuteTemplate(res, "appointmentCreateConfirm.gohtml", ViewData)
	}
}

// Run as Go routine to block users from booking the same dentist on the same date and session
func createNewAppointment(id int, date string, session int, dentist interface{}, pateint *User, appointmentTree **bst.BinarySearchTree, chn chan bool) {
	var sessionBooked bool = false
	// Check if appointment is booked
	appointments := (**appointmentTree).GetAppointmentByDate(date, enumDentist, dentist)
	for _, v := range appointments {
		if v.Session == session {
			sessionBooked = true
			chn <- false
		}
	}
	// If slot is not booked, proceed.
	if !sessionBooked {
		(**appointmentTree).Add(id, date, session, dentist, pateint)
		chn <- true
	}
}

func appointmentEditHandler(userList, appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		var sessionList []AppointmentSession
		dt := time.Now()
		vars := mux.Vars(req)
		appointmentReq := vars["id"]
		appointmentID, _ := strconv.Atoi(appointmentReq)
		appointment := (**appointmentTree).GetAppointmentByID(appointmentID)
		sessions := (**appointmentSessionList).GetList()
		selectedDentist := appointment.Dentist.(*User)
		users := (**userList).GetList()
		dentists := getDentistList(users)

		ViewData := struct {
			LoggedInUser    *User
			PageTitle       string
			CurrentPage     string
			Appointment     *bst.BinaryNode
			Dentists        []*User
			DentistsSession []AppointmentSession
			Sessions        []interface{}
			TodayDate       string
			SelectedDate    string
			SelectedDentist string
		}{
			myUser,
			"Change Appointment",
			"MA",
			appointment,
			dentists,
			sessionList,
			sessions,
			dt.Format("2006-01-02"),
			"",
			selectedDentist.Username,
		}

		// Process form submission
		if req.Method == http.MethodPost {
			inputDate := req.FormValue("apptDate")
			inputDentist := req.FormValue("apptDentist")
			appointmentDate, err := time.Parse("2006-01-02", inputDate)
			if err == nil {
				ret := (**userList).FindByUsername(inputDentist)
				dentist := ret.(*User)

				schedule := (**appointmentTree).GetAppointmentByDate(appointmentDate.Format("2006-01-02"), dentist.Role, dentist)
				retSessionList := (**appointmentSessionList).GetList()
				for _, v := range retSessionList {
					session := v.(AppointmentSession)
					for _, data := range schedule {
						if data.Session == session.Num {
							session.Available = false
						}
					}
					sessionList = append(sessionList, session)
				}
				ViewData.SelectedDentist = dentist.Username
				ViewData.SelectedDate = appointmentDate.Format("2006-01-02")
				ViewData.DentistsSession = sessionList
			}
		}

		tpl.ExecuteTemplate(res, "appointmentEdit.gohtml", ViewData)
	}
}

func appointmentEditConfirmHandler(userList, appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		// Decode and retrive URL query string
		vars := mux.Vars(req)
		appointmentReq := vars["id"]
		dentistReq := vars["dentist"]
		dateReq := vars["date"]
		sessionReq := vars["session"]

		// Date conversion
		appointmentID, _ := strconv.Atoi(appointmentReq)
		session, _ := strconv.Atoi(sessionReq)

		dentist := (**userList).FindByUsername(dentistReq).(*User)
		appointment := (**appointmentTree).GetAppointmentByID(appointmentID)
		sessions := (**appointmentSessionList).GetList()

		oldDentist := appointment.Dentist.(*User)
		oldDate := appointment.Date
		oldSession := appointment.Session

		ViewData := struct {
			LoggedInUser *User
			PageTitle    string
			CurrentPage  string
			Appointment  *bst.BinaryNode
			Dentist      *User
			Date         string
			Session      int
			OldDentist   *User
			OldDate      string
			OldSession   int
			Sessions     []interface{}
			Successful   bool
		}{
			myUser,
			"Confrim Appointment Change",
			"MA",
			appointment,
			dentist,
			dateReq,
			session,
			oldDentist,
			oldDate,
			oldSession,
			sessions,
			false,
		}

		// Process form submission
		if req.Method == http.MethodPost {
			oldAppointment := newAppointment(appointment.ID, appointment.Patient.(*User).Username, appointment.Dentist.(*User).Username, appointment.Date, appointment.Session)
			editiedAppointment := newAppointment(appointment.ID, appointment.Patient.(*User).Username, appointment.Dentist.(*User).Username, appointment.Date, appointment.Session)
			// If there are no changes made to appointment date, update dentist and/or session value
			if appointment.Date == dateReq {
				if appointment.Dentist.(*User).Username != dentist.Username {
					appointment.Dentist = dentist
					editiedAppointment.Dentist = dentist.Username
				}
				if appointment.Session != session {
					appointment.Session = session
					editiedAppointment.Session = session
				}
				ViewData.Successful = true
				updateAppointmentData(oldAppointment, editiedAppointment)
			} else {
				// If changes made to appointment date, added a new appointment and delete old appointment
				var id int = util.GenerateID()
				(**appointmentTree).Add(id, dateReq, session, dentist, appointment.Patient)
				appointmentData := newAppointment(id, appointment.Patient.(*User).Username, dentist.Username, dateReq, session)
				addAppointmentData(appointmentData)
				(**appointmentTree).Remove(appointment)
				deleteAppointmentData(oldAppointment.ID)
				ViewData.Successful = true
			}
		}
		tpl.ExecuteTemplate(res, "appointmentEditConfirm.gohtml", ViewData)
	}
}

func appointmentDeleteHandler(userList, appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		vars := mux.Vars(req)
		appointmentReq := vars["id"]
		appointmentID, _ := strconv.Atoi(appointmentReq)
		appointment := (**appointmentTree).GetAppointmentByID(appointmentID)
		sessions := (**appointmentSessionList).GetList()

		ViewData := struct {
			PageTitle    string
			LoggedInUser *User
			Appointment  *bst.BinaryNode
			Sessions     []interface{}
			Successful   bool
		}{
			"Cancel Appointment",
			myUser,
			appointment,
			sessions,
			false,
		}

		// Process form submission
		if req.Method == http.MethodPost {
			(**appointmentTree).Remove(appointment)
			deleteAppointmentData(appointment.ID)
			ViewData.Successful = true
		}

		tpl.ExecuteTemplate(res, "appointmentDelete.gohtml", ViewData)
	}
}

func appointmentSearchHandler(userList, appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		var sessionList []AppointmentSession
		timeNow := time.Now()
		users := (**userList).GetList()
		dentists := getDentistList(users)

		ViewData := struct {
			LoggedInUser    *User
			PageTitle       string
			CurrentPage     string
			Dentist         *User
			Dentists        []*User
			TodayDate       string
			DentistsSession []AppointmentSession
			SelectedDate    string
		}{
			myUser,
			"Search Available Appointment",
			"SAA",
			nil,
			dentists,
			timeNow.Format("2006-01-02"),
			sessionList,
			"",
		}

		// Process form submission
		if req.Method == http.MethodPost {
			inputDentist := req.FormValue("inputDentist")
			inputDate := req.FormValue("inputDate")

			// Data conversion
			dentist := (**userList).FindByUsername(inputDentist)
			appointmentDate, _ := time.Parse("2006-01-02", inputDate)

			// If valid dentist and input date is entered
			if !(dentist == nil || len(inputDate) == 0) {
				ViewData.Dentist = dentist.(*User)
				ViewData.DentistsSession = getDentistAvailability(appointmentSessionList, appointmentTree, appointmentDate, ViewData.Dentist)
				ViewData.SelectedDate = appointmentDate.Format("2006-01-02")
			}
		}
		tpl.ExecuteTemplate(res, "appointmentSearch.gohtml", ViewData)
	}
}

func getDentistAvailability(appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree, appointmentDate time.Time, Dentist *User) []AppointmentSession {
	var sessionList []AppointmentSession
	appointments := (**appointmentTree).GetAppointmentByDate(appointmentDate.Format("2006-01-02"), Dentist.Role, Dentist)
	retSessionList := (**appointmentSessionList).GetList()
	// Loop Session list and set dentist availability
	for _, v := range retSessionList {
		session := v.(AppointmentSession)
		for _, data := range appointments {
			if data.Session == session.Num {
				session.Available = false
			}
		}
		sessionList = append(sessionList, session)
	}
	return sessionList
}
