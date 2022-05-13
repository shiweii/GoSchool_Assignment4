package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	app "github.com/shiweii/appointment"
	bst "github.com/shiweii/binarysearchtree"
	dll "github.com/shiweii/doublylinkedlist"
	ede "github.com/shiweii/encryptdecrypt"
	"github.com/shiweii/logger"
	"github.com/shiweii/user"
	util "github.com/shiweii/utility"
	"github.com/shiweii/validator"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	enumPatient  = "patient"
	enumAdmin    = "admin"
	enumDentist  = "dentist"
	enumUpcoming = "upcoming"
)

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
			LoggedInUser *user.User
			PageTitle    string
		}{
			myUser,
			"Central City Dentist Clinic",
		}

		tpl.ExecuteTemplate(res, "index.gohtml", ViewData)
	}
}

func signupHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
				var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
				Error.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		if alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}

		var myUser user.User

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
		}{
			nil,
			"Sign Up",
		}
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
				user.AddUserDate(&myUser)
			}
			// redirect to patient landing page
			http.Redirect(res, req, "/appointments", http.StatusSeeOther)
			return

		}
		tpl.ExecuteTemplate(res, "signup.gohtml", ViewData)
	}
}

func loginHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		if alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
			LoginFail    bool
		}{
			nil,
			"Login",
			false,
		}

		// process form submission
		if req.Method == http.MethodPost {
			username := req.FormValue("username")
			password := req.FormValue("password")

			var userObj *user.User

			// check if user exist with username
			userItf := (**userList).FindByUsername(username)
			if userItf == nil {
				ViewData.LoginFail = true
				logger.Info.Println("Login fail. user:", username)
			}

			// Check if user is deleted
			if !ViewData.LoginFail {
				userObj = userItf.(*user.User)
				if userObj.IsDeleted {
					ViewData.LoginFail = true
					logger.Info.Println("Login fail. user:", username)
				}
			}

			// Matching of password entered
			if !ViewData.LoginFail {
				err := bcrypt.CompareHashAndPassword([]byte(userObj.Password), []byte(password))
				if err != nil {
					ViewData.LoginFail = true
					logger.Info.Println("Login fail. user:", username)
				}
			}

			if !ViewData.LoginFail {
				id := uuid.NewV4()
				myCookie := &http.Cookie{
					Name:  "myCookie",
					Value: id.String(),
				}
				http.SetCookie(res, myCookie)
				mapSessions[myCookie.Value] = userObj.Username
				http.Redirect(res, req, "/appointments", http.StatusSeeOther)
				return
			}
			logger.Info.Println("Login successful. user:", username)
		}
		tpl.ExecuteTemplate(res, "login.gohtml", ViewData)
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
		dentists := user.GetDentistList(users)

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Appointments []*bst.BinaryNode
			Sessions     []interface{}
			Dentists     []*user.User
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
				ViewData.Appointments = app.GetDuplicate(result, filterCount)
			}
		}
		tpl.ExecuteTemplate(res, "appointmentList.gohtml", ViewData)
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

		var sessionList []app.AppointmentSession
		timeNow := time.Now()
		users := (**userList).GetList()
		dentists := user.GetDentistList(users)

		ViewData := struct {
			LoggedInUser    *user.User
			PageTitle       string
			CurrentPage     string
			Dentist         *user.User
			Dentists        []*user.User
			TodayDate       string
			DentistsSession []app.AppointmentSession
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
				ViewData.Dentist = dentist.(*user.User)
				ViewData.DentistsSession = app.GetDentistAvailability(appointmentSessionList, appointmentTree, appointmentDate, ViewData.Dentist)
				ViewData.SelectedDate = appointmentDate.Format("2006-01-02")
			}
		}
		tpl.ExecuteTemplate(res, "appointmentSearch.gohtml", ViewData)
	}
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
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Dentists     []*user.User
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			user.GetDentistList(users),
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

		var sessionList []app.AppointmentSession
		dt := time.Now()

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Dentist      *user.User
			TodayDate    string
			Sessions     []app.AppointmentSession
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
			ViewData.Dentist = dentist.(*user.User)
		}

		// Process form submission
		if req.Method == http.MethodPost {
			inputDate := req.FormValue("apptDate")
			appointmentDate, err := time.Parse("2006-01-02", inputDate)
			if err == nil {
				ViewData.Sessions = app.GetDentistAvailability(appointmentSessionList, appointmentTree, appointmentDate, ViewData.Dentist)
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

		session := (**appointmentSessionList).Get(ses).(app.AppointmentSession)
		dentist := (**userList).FindByUsername(dentistReq)

		ViewData := struct {
			LoggedInUser  *user.User
			PageTitle     string
			CurrentPage   string
			Dentist       *user.User
			Date          string
			StartTime     string
			EndTime       string
			Successful    bool
			FormSubmitted bool
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			dentist.(*user.User),
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
			go app.CreateNewAppointment(id, appointmentDate.Format("2006-01-02"), session.Num, dentist, myUser, appointmentTree, chn)
			successful := <-chn
			if successful {
				appointment := app.NewAppointment(id, myUser.Username, dentist.(*user.User).Username, appointmentDate.Format("2006-01-02"), session.Num)
				app.AddAppointmentData(appointment)
				ViewData.Successful = true
			}
			ViewData.FormSubmitted = true
		}

		tpl.ExecuteTemplate(res, "appointmentCreateConfirm.gohtml", ViewData)
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

		var sessionList []app.AppointmentSession
		dt := time.Now()
		vars := mux.Vars(req)
		appointmentReq := vars["id"]
		appointmentID, _ := strconv.Atoi(appointmentReq)
		appointment := (**appointmentTree).GetAppointmentByID(appointmentID)
		sessions := (**appointmentSessionList).GetList()
		selectedDentist := appointment.Dentist.(*user.User)
		users := (**userList).GetList()
		dentists := user.GetDentistList(users)

		ViewData := struct {
			LoggedInUser    *user.User
			PageTitle       string
			CurrentPage     string
			Appointment     *bst.BinaryNode
			Dentists        []*user.User
			DentistsSession []app.AppointmentSession
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
				dentist := ret.(*user.User)

				schedule := (**appointmentTree).GetAppointmentByDate(appointmentDate.Format("2006-01-02"), dentist.Role, dentist)
				retSessionList := (**appointmentSessionList).GetList()
				for _, v := range retSessionList {
					session := v.(app.AppointmentSession)
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

		dentist := (**userList).FindByUsername(dentistReq).(*user.User)
		appointment := (**appointmentTree).GetAppointmentByID(appointmentID)
		sessions := (**appointmentSessionList).GetList()

		oldDentist := appointment.Dentist.(*user.User)
		oldDate := appointment.Date
		oldSession := appointment.Session

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Appointment  *bst.BinaryNode
			Dentist      *user.User
			Date         string
			Session      int
			OldDentist   *user.User
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
			oldAppointment := app.NewAppointment(appointment.ID, appointment.Patient.(*user.User).Username, appointment.Dentist.(*user.User).Username, appointment.Date, appointment.Session)
			editiedAppointment := app.NewAppointment(appointment.ID, appointment.Patient.(*user.User).Username, appointment.Dentist.(*user.User).Username, appointment.Date, appointment.Session)
			// If there are no changes made to appointment date, update dentist and/or session value
			if appointment.Date == dateReq {
				if appointment.Dentist.(*user.User).Username != dentist.Username {
					appointment.Dentist = dentist
					editiedAppointment.Dentist = dentist.Username
				}
				if appointment.Session != session {
					appointment.Session = session
					editiedAppointment.Session = session
				}
				ViewData.Successful = true
				app.UpdateAppointmentData(oldAppointment, editiedAppointment)
			} else {
				// If changes made to appointment date, added a new appointment and delete old appointment
				var id int = util.GenerateID()
				(**appointmentTree).Add(id, dateReq, session, dentist, appointment.Patient)
				appointmentData := app.NewAppointment(id, appointment.Patient.(*user.User).Username, dentist.Username, dateReq, session)
				app.AddAppointmentData(appointmentData)
				(**appointmentTree).Remove(appointment)
				app.DeleteAppointmentData(oldAppointment.ID)
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
			LoggedInUser *user.User
			CurrentPage  string
			Appointment  *bst.BinaryNode
			Sessions     []interface{}
			Successful   bool
		}{
			"Cancel Appointment",
			myUser,
			"MA",
			appointment,
			sessions,
			false,
		}

		// Process form submission
		if req.Method == http.MethodPost {
			(**appointmentTree).Remove(appointment)
			app.DeleteAppointmentData(appointment.ID)
			ViewData.Successful = true
		}

		tpl.ExecuteTemplate(res, "appointmentDelete.gohtml", ViewData)
	}
}

func userListHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		myUser, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		users := (**userList).GetList()

		ViewData := struct {
			LoggedInUser   *user.User
			PageTitle      string
			CurrentPage    string
			Users          []interface{}
			Successful     bool
			ErrorDelete    bool
			ErrorDeleteMsg string
		}{
			myUser,
			"Manage Users",
			"MU",
			users,
			false,
			false,
			"",
		}

		tpl.ExecuteTemplate(res, "userList.gohtml", ViewData)
	}
}

func userEditHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
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
		username := vars["username"]

		if myUser.Role == enumPatient {
			if username != myUser.Username {
				http.Redirect(res, req, "/", http.StatusUnauthorized)
				return
			}
		}

		ret := (**userList).FindByUsername(username)
		selectedUser := ret.(*user.User)
		copyUser := user.NewUser(selectedUser.Username, selectedUser.Password, selectedUser.Role, selectedUser.FirstName, selectedUser.LastName, selectedUser.MobileNumber)

		ViewData := struct {
			LoggedInUser         *user.User
			PageTitle            string
			CurrentPage          string
			UserData             *user.User
			ValidateFirstName    bool
			ValidateLastName     bool
			ValidateMobileNumber bool
			ValidatePassword     bool
			Successful           bool
		}{
			myUser,
			"Edit User Information",
			"",
			selectedUser,
			true,
			true,
			true,
			true,
			false,
		}

		if ViewData.LoggedInUser.Role == enumAdmin {
			ViewData.CurrentPage = "MU"
		}

		// process form submission
		if req.Method == http.MethodPost {
			var edited bool = false
			// Validate first name input
			inputFirstName := req.FormValue("firstName")
			if validator.IsEmpty(inputFirstName) || !validator.IsAlphabet(inputFirstName) {
				ViewData.ValidateFirstName = false
			}
			if ViewData.ValidateFirstName {
				if c := strings.Compare(inputFirstName, selectedUser.FirstName); c != 0 {
					selectedUser.FirstName = inputFirstName
					edited = true
				}
			}
			// Validate last name input
			inputLastName := req.FormValue("lastName")
			if validator.IsEmpty(inputLastName) || !validator.IsAlphabet(inputLastName) {
				ViewData.ValidateLastName = false
			}
			if ViewData.ValidateLastName {
				if c := strings.Compare(inputLastName, selectedUser.LastName); c != 0 {
					selectedUser.LastName = inputLastName
					edited = true
				}
			}
			// Validate mobile number input
			inputMobile := req.FormValue("mobileNum")
			if validator.IsEmpty(inputMobile) || !validator.IsMobileNumber(inputMobile) {
				ViewData.ValidateMobileNumber = false
			}
			if ViewData.ValidateMobileNumber {
				mobileNumber, _ := strconv.Atoi(inputMobile)
				if mobileNumber != selectedUser.MobileNumber {
					selectedUser.MobileNumber = mobileNumber
					edited = true
				}
			}
			// Change Password
			inputPassword := req.FormValue("password")
			if len(inputPassword) > 0 {
				// Matching of password entered
				err := bcrypt.CompareHashAndPassword([]byte(selectedUser.Password), []byte(inputPassword))
				// Different password
				if err != nil {
					bPassword, err := bcrypt.GenerateFromPassword([]byte(inputPassword), bcrypt.MinCost)
					if err != nil {
						http.Error(res, "Internal server error", http.StatusInternalServerError)
						return
					} else {
						selectedUser.Password = string(bPassword)
						edited = true
					}
				}
			}

			chkboxInput := req.FormValue("deleteChkBox")
			deleteChkBox, err := strconv.ParseBool(chkboxInput)
			if err != nil {
				deleteChkBox = false
			}

			// Validation completed
			if ViewData.ValidateFirstName && ViewData.ValidateLastName && ViewData.ValidateMobileNumber {
				if edited {
					user.UpdateUserData(copyUser, selectedUser)
				}
				if deleteChkBox && !selectedUser.IsDeleted {
					selectedUser.IsDeleted = true
					user.DeleteUserData(selectedUser)
				}
				if !deleteChkBox && selectedUser.IsDeleted {
					selectedUser.IsDeleted = false
					user.DeleteUserData(selectedUser)
				}
				ViewData.Successful = true
			}
		}

		err := tpl.ExecuteTemplate(res, "userEdit.gohtml", ViewData)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func userDeleteHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
				var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
				Error.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		myUser, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		vars := mux.Vars(req)
		username := vars["username"]

		var users []interface{}

		ViewData := struct {
			LoggedInUser   *user.User
			PageTitle      string
			CurrentPage    string
			Users          []interface{}
			Successful     bool
			ErrorDelete    bool
			ErrorDeleteMsg string
		}{
			myUser,
			"Manage Users",
			"MU",
			users,
			false,
			false,
			"",
		}

		retUser := (**userList).FindByUsername(username).(*user.User)
		if retUser == nil {
			ViewData.ErrorDelete = true
			ViewData.ErrorDeleteMsg = "Error deleteing user: " + username + " user does not exist."
		} else {
			// Soft delete user
			retUser.IsDeleted = true
			ViewData.Successful = true
			user.DeleteUserData(retUser)
		}
		ViewData.Users = (**userList).GetList()

		tpl.ExecuteTemplate(res, "userList.gohtml", ViewData)
	}
}

func sessionListHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		myUser, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
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

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Sessions     []SessionStrcut
		}{
			myUser,
			"Manage Session",
			"MS",
			sessions,
		}

		for k, v := range mapSessions {
			user := (**userList).FindByUsername(v).(*user.User)
			ViewData.Sessions = append(ViewData.Sessions, SessionStrcut{SessionID: k, Username: v, Role: user.Role})
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

		err := tpl.ExecuteTemplate(res, "sessions.gohtml", ViewData)
		if err != nil {
			log.Fatalln(err)
		}
	}
}