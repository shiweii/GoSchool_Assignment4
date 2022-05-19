package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	app "github.com/shiweii/appointment"
	dll "github.com/shiweii/doublylinkedlist"
	"github.com/shiweii/logger"
	"github.com/shiweii/user"
	util "github.com/shiweii/utility"
	"github.com/shiweii/validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	enumPatient  = "patient"
	enumAdmin    = "admin"
	enumDentist  = "dentist"
	enumUpcoming = "upcoming"
)

func indexHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		if alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/appointments", http.StatusSeeOther)
			return
		}

		// Expire cookie if user's session was removed by admin
		cookie, err := req.Cookie(util.GetEnvVar("COOKIE_NAME"))
		if err == nil {
			username := mapSessions[cookie.Value]
			if username == "" {
				cookie = expireCookie()
				http.SetCookie(res, cookie)
			}
		}

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
		}{
			nil,
			"Central City Dentist Clinic",
		}

		if err := tpl.ExecuteTemplate(res, "index.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func signupHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				util.CheckEncryption()
				logger.Panic.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		if alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}

		ViewData := struct {
			LoggedInUser         *user.User
			PageTitle            string
			ValidateFirstName    bool
			ValidateLastName     bool
			ValidateUserName     bool
			UserNameTaken        bool
			ValidatePassword     bool
			ValidateMobileNumber bool
			InputUserName        string
			InputPassword        string
			InputFirstName       string
			InputLastName        string
			InputMobileNumber    string
		}{
			nil,
			"Sign Up",
			true,
			true,
			true,
			false,
			true,
			true,
			"",
			"",
			"",
			"",
			"",
		}

		// process form submission
		if req.Method == http.MethodPost {
			// get form values
			ViewData.InputUserName = strings.TrimSpace(req.FormValue("username"))
			ViewData.InputPassword = strings.TrimSpace(req.FormValue("password"))
			ViewData.InputFirstName = strings.TrimSpace(req.FormValue("firstname"))
			ViewData.InputLastName = strings.TrimSpace(req.FormValue("lastname"))
			ViewData.InputMobileNumber = strings.TrimSpace(req.FormValue("mobileNum"))

			logger.Trace.Printf("%v: Username: %v, FirstName: %v, LastName: %v, MobileNumber: %v", util.CurrFuncName(), ViewData.InputUserName, ViewData.InputFirstName, ViewData.InputLastName, ViewData.InputMobileNumber)

			//Validate Fields
			if validator.IsEmpty(ViewData.InputUserName) || !validator.IsValidUsername(ViewData.InputUserName) {
				ViewData.ValidateUserName = false
			} else {
				// check if username exist/ taken
				useObj := (*userList).FindByUsername(ViewData.InputUserName)
				if useObj != nil {
					ViewData.ValidateUserName = false
					ViewData.UserNameTaken = true
				}
			}
			if validator.IsEmpty(ViewData.InputPassword) || !validator.IsValidPassword(ViewData.InputPassword) {
				ViewData.ValidatePassword = false
			}
			if validator.IsEmpty(ViewData.InputFirstName) || !validator.IsValidName(ViewData.InputFirstName) {
				ViewData.ValidateFirstName = false
			}
			if validator.IsEmpty(ViewData.InputLastName) || !validator.IsValidName(ViewData.InputLastName) {
				ViewData.ValidateLastName = false
			}
			if validator.IsEmpty(ViewData.InputMobileNumber) || !validator.IsMobileNumber(ViewData.InputMobileNumber) {
				ViewData.ValidateMobileNumber = false
			}

			// If all validations are true
			if ViewData.ValidateFirstName && ViewData.ValidateLastName && ViewData.ValidateUserName && ViewData.ValidatePassword && ViewData.ValidateMobileNumber {
				var myUser user.User
				// create session
				cookie := createNewSecureCookie()
				http.SetCookie(res, cookie)
				mapSessions[cookie.Value] = ViewData.InputUserName

				bPassword, err := bcrypt.GenerateFromPassword([]byte(ViewData.InputPassword), bcrypt.MinCost)
				if err != nil {
					logger.Trace.Printf("%v: %v", util.CurrFuncName(), err)
					http.Error(res, "Internal server error", http.StatusInternalServerError)
					return
				}

				myUser.Username = ViewData.InputUserName
				myUser.Password = string(bPassword)
				myUser.FirstName = ViewData.InputFirstName
				myUser.LastName = ViewData.InputLastName
				myUser.Role = enumPatient
				mobileNum, _ := strconv.Atoi(ViewData.InputMobileNumber)
				myUser.MobileNumber = mobileNum

				// Add into linklist and JSON
				if err = (*userList).Add(&myUser); err != nil {
					logger.Error.Printf("%v: Error:", util.CurrFuncName(), err)
				} else {
					(*userList).InsertionSort()
					user.AddUserDate(&myUser)
				}
				// redirect to patient landing page
				http.Redirect(res, req, "/appointments", http.StatusSeeOther)
				return
			}
		}
		if err := tpl.ExecuteTemplate(res, "signup.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func loginHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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

			// Retrieve form input and remove empty spaces
			inputUserName := strings.TrimSpace(req.FormValue("username"))
			inputPassword := strings.TrimSpace(req.FormValue("password"))

			//Validate Fields
			if validator.IsEmpty(inputUserName) || !validator.IsValidUsername(inputUserName) {
				ViewData.LoginFail = true
			}

			// check if user exist with username
			if !ViewData.LoginFail {
				ViewData.LoggedInUser = (*userList).FindByUsername(inputUserName)
				if ViewData.LoggedInUser == nil {
					ViewData.LoginFail = true
					logger.Info.Printf("%v: Login fail. user: %v", util.CurrFuncName(), inputUserName)
				}
			}

			// Check if user is deleted
			if !ViewData.LoginFail {
				if ViewData.LoggedInUser.IsDeleted {
					ViewData.LoginFail = true
					logger.Info.Printf("%v: Login fail. user: %v", util.CurrFuncName(), ViewData.LoggedInUser.Username)
				}
			}

			// Matching of password entered
			if !ViewData.LoginFail {
				err := bcrypt.CompareHashAndPassword([]byte(ViewData.LoggedInUser.Password), []byte(inputPassword))
				if err != nil {
					ViewData.LoginFail = true
					logger.Info.Printf("%v: Login fail. user: %v", util.CurrFuncName(), ViewData.LoggedInUser.Username)
				}
			}

			if !ViewData.LoginFail {
				cookie := createNewSecureCookie()
				go terminateOtherSession(cookie.Value, ViewData.LoggedInUser.Username)
				http.SetCookie(res, cookie)
				mapSessions[cookie.Value] = ViewData.LoggedInUser.Username
				logger.Info.Printf("%v: Login successful. user:%v", util.CurrFuncName(), ViewData.LoggedInUser.Username)
				http.Redirect(res, req, "/appointments", http.StatusSeeOther)
				return
			}
		}
		if err := tpl.ExecuteTemplate(res, "login.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func logoutHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}
		cookie, err := req.Cookie(util.GetEnvVar("COOKIE_NAME"))
		if err == nil {
			// Get username
			username, _ := mapSessions[cookie.Value]
			logger.Info.Printf("%v: Logout... user [%v]", util.CurrFuncName(), username)
			// delete the session
			delete(mapSessions, cookie.Value)
			// Expire the cookie
			cookie = expireCookie()
			http.SetCookie(res, cookie)
		}
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func appointmentListHandler(userList *user.DoublyLinkedList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree *app.BinarySearchTree) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Appointments []*app.Appointment
			Sessions     []interface{}
			Dentists     []*user.User
			Option       string
			TodayDate    string
		}{
			myUser,
			"Appointments",
			"MA",
			nil,
			(**appointmentSessionList).GetList(),
			(*userList).GetDentistList(),
			"",
			time.Now().Format("2006-01-02"),
		}

		//var appointments []*bst.BinaryNode
		ViewData.Option = strings.TrimSpace(req.FormValue("view"))

		// If logged in as admin, display all appointments
		if myUser.Role == enumAdmin {
			ViewData.Appointments = (*appointmentTree).GetAllAppointments(nil, "")
		}

		// If logged in as patient, display appointments based on selection
		if myUser.Role == enumPatient {
			if len(ViewData.Option) == 0 {
				ViewData.Option = enumUpcoming
			}
			if ViewData.Option == enumUpcoming {
				ViewData.Appointments = (*appointmentTree).GetUpComingAppointments(myUser, enumPatient)
			} else {
				ViewData.Appointments = (*appointmentTree).GetAllAppointments(myUser, enumPatient)
			}
		}

		// Process form submission
		if req.Method == http.MethodPost {
			inputDentist := strings.TrimSpace(req.FormValue("inputDentist"))
			inputDate := strings.TrimSpace(req.FormValue("inputDate"))
			inputPatientMobileNumber := strings.TrimSpace(req.FormValue("inputPatientMobileNumber"))
			inputSession := strings.TrimSpace(req.FormValue("inputSession"))

			// Data conversation
			dentist := (*userList).FindByUsername(inputDentist)
			appointmentDate, _ := time.Parse("2006-01-02", inputDate)
			appointmentSession, _ := strconv.Atoi(inputSession)
			patientMobileNumber, _ := strconv.Atoi(inputPatientMobileNumber)

			// If inputs are valid
			if !(dentist == nil && len(inputDate) == 0 && appointmentSession == 0 && len(inputPatientMobileNumber) == 0) {

				// Initialize channels
				chSearchDate := make(chan []*app.Appointment)
				chSearchPatient := make(chan []*app.Appointment)
				chSearchDentist := make(chan []*app.Appointment)
				chSearchSession := make(chan []*app.Appointment)
				filterCount := 0

				if dentist != nil {
					filterCount++
					go (*appointmentTree).SearchAllByField(enumDentist, dentist, chSearchDentist)
				}
				if len(inputDate) > 0 {
					filterCount++
					go (*appointmentTree).SearchAllByField("date", appointmentDate.Format("2006-01-02"), chSearchDate)
				}
				if len(inputPatientMobileNumber) > 0 {
					filterCount++
					patient := (*userList).SearchByMobileNumber(patientMobileNumber)
					go (*appointmentTree).SearchAllByField(enumPatient, patient, chSearchPatient)
				}
				if appointmentSession > 0 {
					filterCount++
					go (*appointmentTree).SearchAllByField("session", appointmentSession, chSearchSession)
				}

				var result []*app.Appointment
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
		if err := tpl.ExecuteTemplate(res, "appointmentList.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentSearchHandler(userList *user.DoublyLinkedList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree *app.BinarySearchTree) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
			LoggedInUser    *user.User
			PageTitle       string
			CurrentPage     string
			Dentist         *user.User
			Dentists        []*user.User
			TodayDate       string
			DentistsSession []app.AppSession
			SelectedDate    string
			FormProcessed   bool
		}{
			myUser,
			"Search Available Appointment",
			"SAA",
			nil,
			userList.GetDentistList(),
			time.Now().Format("2006-01-02"),
			nil,
			"",
			false,
		}

		// Process form submission
		if req.Method == http.MethodPost {
			inputDentist := strings.TrimSpace(req.FormValue("inputDentist"))
			inputDate := strings.TrimSpace(req.FormValue("inputDate"))

			// Data conversion
			dentist := (*userList).FindByUsername(inputDentist)
			appointmentDate, _ := time.Parse("2006-01-02", inputDate)

			// If valid dentist and input date is entered
			if !(dentist == nil || len(inputDate) == 0) {
				ViewData.Dentist = dentist
				ViewData.DentistsSession = app.GetDentistAvailability(appointmentSessionList, appointmentTree, appointmentDate, ViewData.Dentist)
				ViewData.SelectedDate = appointmentDate.Format("2006-01-02")
			}
			ViewData.FormProcessed = true
		}
		if err := tpl.ExecuteTemplate(res, "appointmentSearch.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentCreateHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Dentists     []*user.User
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			(*userList).GetDentistList(),
		}

		if err := tpl.ExecuteTemplate(res, "appointmentCreate_step1.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentCreatePart2Handler(userList *user.DoublyLinkedList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree *app.BinarySearchTree) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Dentist      *user.User
			TodayDate    string
			Sessions     []app.AppSession
			SelectedDate string
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			nil,
			time.Now().Format("2006-01-02"),
			nil,
			"",
		}

		// Get data from query string
		vars := mux.Vars(req)
		dentistReq := vars["dentist"]
		ViewData.Dentist = (*userList).FindByUsername(dentistReq)

		// Process form submission
		if req.Method == http.MethodPost {
			inputDate := req.FormValue("appDate")
			appointmentDate, err := time.Parse("2006-01-02", inputDate)
			if err == nil {
				ViewData.Sessions = app.GetDentistAvailability(appointmentSessionList, appointmentTree, appointmentDate, ViewData.Dentist)
				ViewData.SelectedDate = appointmentDate.Format("2006-01-02")
			}
		}
		if err := tpl.ExecuteTemplate(res, "appointmentCreate_step2.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentCreateConfirmHandler(userList *user.DoublyLinkedList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree *app.BinarySearchTree) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
		dentistReq := vars["dentist"]
		dateReq := vars["date"]
		sessionReq := vars["session"]

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
			IsInputError  bool
		}{
			myUser,
			"Create New Appointment",
			"CNA",
			nil,
			"",
			"",
			"",
			false,
			false,
			false,
		}

		// Validating inputs
		ViewData.Dentist = (*userList).FindByUsername(dentistReq)
		if ViewData.Dentist == nil {
			ViewData.IsInputError = true
		}
		appointmentDate, err := time.Parse("2006-01-02", dateReq)
		if err != nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Error Parsing Data [%v]", util.CurrFuncName(), dateReq)
		}
		ses, err := strconv.Atoi(sessionReq)
		if err != nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Error Parsing Session Number [%v]", util.CurrFuncName(), sessionReq)
		}
		if ViewData.IsInputError {
			if err := tpl.ExecuteTemplate(res, "appointmentCreateConfirm.gohtml", ViewData); err != nil {
				logger.Error.Println(err)
			}
			return
		}

		logger.Trace.Printf("%v: Dentist [%v], Date [%v], Session [%v]", util.CurrFuncName(), dentistReq, dateReq, sessionReq)

		ViewData.Date = appointmentDate.Format("2006-01-02")
		session := (**appointmentSessionList).Get(ses).(app.AppSession)
		ViewData.StartTime = session.StartTime
		ViewData.EndTime = session.EndTime

		// Process form submission
		if req.Method == http.MethodPost {
			var id = util.GenerateID()
			chn := make(chan bool)
			go app.CreateNewAppointment(id, ViewData.Date, session.Num, ViewData.Dentist, myUser, appointmentTree, chn)
			successful := <-chn
			if successful {
				logger.Info.Printf("%v: Appointment created successfully.", util.CurrFuncName())
				logger.Trace.Printf("%v: Adding appointment data into JSON: id:[%v], username:[%v], dentist:[%v], date:[%v], session:[%v]", util.CurrFuncName(), id, myUser.Username, ViewData.Dentist.Username, ViewData.Date, session.Num)
				ViewData.Successful = true
			}
			ViewData.FormSubmitted = true
		}

		if err = tpl.ExecuteTemplate(res, "appointmentCreateConfirm.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentEditHandler(userList *user.DoublyLinkedList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree *app.BinarySearchTree) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
			LoggedInUser    *user.User
			PageTitle       string
			CurrentPage     string
			Appointment     *app.Appointment
			Dentists        []*user.User
			DentistsSession []app.AppSession
			Sessions        []interface{}
			TodayDate       string
			SelectedDate    string
			SelectedDentist string
			IsInputError    bool
		}{
			myUser,
			"Change Appointment",
			"MA",
			nil,
			(*userList).GetDentistList(),
			nil,
			(**appointmentSessionList).GetList(),
			time.Now().Format("2006-01-02"),
			"",
			"",
			false,
		}

		vars := mux.Vars(req)
		appointmentReq := vars["id"]

		appointmentID, _ := strconv.Atoi(appointmentReq)
		ViewData.Appointment = (*appointmentTree).GetAppointmentByID(appointmentID)
		if ViewData.Appointment == nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Application does not exist ID:[%v]", util.CurrFuncName(), appointmentID)
			if err := tpl.ExecuteTemplate(res, "appointmentEdit.gohtml", ViewData); err != nil {
				logger.Error.Println(err)
			}
			return
		}
		ViewData.SelectedDentist = ViewData.Appointment.Dentist.(*user.User).Username

		// Process form submission
		if req.Method == http.MethodPost {
			inputDate := strings.TrimSpace(req.FormValue("appDate"))
			inputDentist := strings.TrimSpace(req.FormValue("appDentist"))
			appointmentDate, err := time.Parse("2006-01-02", inputDate)
			if err == nil {
				dentist := (*userList).FindByUsername(inputDentist)
				var sessionList []app.AppSession
				schedule := (*appointmentTree).GetAppointmentByDate(appointmentDate.Format("2006-01-02"), dentist.Role, dentist)
				retSessionList := ViewData.Sessions
				for _, v := range retSessionList {
					session := v.(app.AppSession)
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

		if err := tpl.ExecuteTemplate(res, "appointmentEdit.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentEditConfirmHandler(userList *user.DoublyLinkedList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree *app.BinarySearchTree) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
		dentistReq := vars["dentist"]
		dateReq := vars["date"]
		sessionReq := vars["session"]

		ViewData := struct {
			LoggedInUser       *user.User
			PageTitle          string
			CurrentPage        string
			CurrentAppointment *app.Appointment
			OldDentist         *user.User
			OldDate            string
			OldSession         int
			EditedDentist      *user.User
			EditedDate         string
			EditedSession      int
			SessionList        []interface{}
			Successful         bool
			Unsuccessful       bool
			UnsuccessfulMsg    string
			IsInputError       bool
		}{
			myUser,
			"Confirm Appointment Change",
			"MA",
			nil,
			nil,
			"",
			0,
			nil,
			"",
			0,
			(**appointmentSessionList).GetList(),
			false,
			false,
			"",
			false,
		}

		// Validate Data
		// Validate Appointment
		appointmentID, err := strconv.Atoi(appointmentReq)
		if err != nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Error Parsing ID [%v]", util.CurrFuncName(), appointmentReq)
		}
		ViewData.CurrentAppointment = (*appointmentTree).GetAppointmentByID(appointmentID)
		if ViewData.CurrentAppointment == nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Application does not exist ID:[%v]", util.CurrFuncName(), appointmentID)
		}
		ViewData.OldDentist = ViewData.CurrentAppointment.Dentist.(*user.User)
		ViewData.OldDate = ViewData.CurrentAppointment.Date
		ViewData.OldSession = ViewData.CurrentAppointment.Session
		// Validate Dentist
		ViewData.EditedDentist = (*userList).FindByUsername(dentistReq)
		if ViewData.EditedDentist == nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Dentist [%v] not found", util.CurrFuncName(), dentistReq)
		}
		// Validate Date
		parsedDate, err := time.Parse("2006-01-02", dateReq)
		if err != nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Error Parsing Data [%v]", util.CurrFuncName(), dateReq)
		}
		ViewData.EditedDate = parsedDate.Format("2006-01-02")
		// Validate Session
		ViewData.EditedSession, err = strconv.Atoi(sessionReq)
		if err != nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Error Parsing Session Number [%v]", util.CurrFuncName(), sessionReq)
		}
		// If validation fail
		if ViewData.IsInputError {
			if err := tpl.ExecuteTemplate(res, "appointmentEditConfirm.gohtml", ViewData); err != nil {
				logger.Error.Println(err)
			}
			return
		}

		logger.Trace.Printf("%v: Application ID [%v], Dentist [%v], Date [%v], Session [%v]", util.CurrFuncName(), appointmentReq, dentistReq, dateReq, sessionReq)

		if req.Method == http.MethodPost {
			currentAppointment := app.New(ViewData.CurrentAppointment.ID, ViewData.CurrentAppointment.Patient.(*user.User).Username, ViewData.CurrentAppointment.Dentist.(*user.User).Username, ViewData.CurrentAppointment.Date, ViewData.CurrentAppointment.Session)
			newAppointment := app.New(ViewData.CurrentAppointment.ID, ViewData.CurrentAppointment.Patient.(*user.User).Username, ViewData.CurrentAppointment.Dentist.(*user.User).Username, ViewData.CurrentAppointment.Date, ViewData.CurrentAppointment.Session)
			// If there's no change to appointment date
			if ViewData.CurrentAppointment.Date == ViewData.EditedDate {
				// If there's a change in dentist
				if ViewData.CurrentAppointment.Dentist.(*user.User).Username != ViewData.EditedDentist.Username {
					ViewData.CurrentAppointment.Dentist = ViewData.EditedDentist
					newAppointment.Dentist = ViewData.EditedDentist.Username
				}
				// If there's a change in session
				if ViewData.CurrentAppointment.Session != ViewData.EditedSession {
					ViewData.CurrentAppointment.Session = ViewData.EditedSession
					newAppointment.Session = ViewData.EditedSession
				}
				ViewData.Successful = true
				// Update JSON
				app.UpdateAppointmentData(currentAppointment, newAppointment)
			} else {
				// If there's change to appointment date
				var id = util.GenerateID()
				chn := make(chan bool)
				go app.CreateNewAppointment(id, ViewData.EditedDate, ViewData.EditedSession, ViewData.EditedDentist, ViewData.CurrentAppointment.Patient.(*user.User), appointmentTree, chn)
				successful := <-chn
				if successful {
					if err := (*appointmentTree).DeleteAppointment(ViewData.CurrentAppointment); err != nil {
						logger.Error.Println(err)
						ViewData.Unsuccessful = true
						ViewData.UnsuccessfulMsg = "There's an error processing your transaction, please try again later."
					} else {
						ViewData.Successful = true
					}
				} else {
					ViewData.Unsuccessful = true
				}
			}
		}
		if err := tpl.ExecuteTemplate(res, "appointmentEditConfirm.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentDeleteHandler(userList *user.DoublyLinkedList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree *app.BinarySearchTree) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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
			PageTitle    string
			LoggedInUser *user.User
			CurrentPage  string
			Appointment  *app.Appointment
			Sessions     []interface{}
			Successful   bool
			IsInputError bool
		}{
			"Cancel Appointment",
			myUser,
			"MA",
			nil,
			nil,
			false,
			false,
		}

		vars := mux.Vars(req)
		appointmentReq := vars["id"]

		appointmentID, _ := strconv.Atoi(appointmentReq)
		ViewData.Appointment = (*appointmentTree).GetAppointmentByID(appointmentID)
		if ViewData.Appointment == nil {
			ViewData.IsInputError = true
			logger.Error.Printf("%v: Application does not exist ID:[%v]", util.CurrFuncName(), appointmentID)
			if err := tpl.ExecuteTemplate(res, "appointmentEdit.gohtml", ViewData); err != nil {
				logger.Error.Println(err)
			}
			return
		}

		ViewData.Sessions = (**appointmentSessionList).GetList()

		// Process form submission
		if req.Method == http.MethodPost {
			if err := (*appointmentTree).DeleteAppointment(ViewData.Appointment); err != nil {
				logger.Error.Printf("%v: Error: %v", util.CurrFuncName(), err)
			} else {
				ViewData.Successful = true
			}
		}
		if err := tpl.ExecuteTemplate(res, "appointmentDelete.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func userListHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		myUser, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

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
			(*userList).GetList(),
			false,
			false,
			"",
		}

		if err := tpl.ExecuteTemplate(res, "userList.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func userEditHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				util.CheckEncryption()
				logger.Panic.Println(err)
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
			nil,
			true,
			true,
			true,
			true,
			false,
		}
		if ViewData.LoggedInUser.Role == enumAdmin {
			ViewData.CurrentPage = "MU"
		}

		ViewData.UserData = (*userList).FindByUsername(username)
		if ViewData.UserData == nil {
			logger.Error.Printf("%v: User Not Found: %v", util.CurrFuncName(), username)
			if err := tpl.ExecuteTemplate(res, "userEdit.gohtml", ViewData); err != nil {
				logger.Error.Println(err)
			}
			return
		}

		copyUser := user.New(ViewData.UserData.Username, ViewData.UserData.Password, ViewData.UserData.Role, ViewData.UserData.FirstName, ViewData.UserData.LastName, ViewData.UserData.MobileNumber)

		// process form submission
		if req.Method == http.MethodPost {
			var edited = false

			inputFirstName := strings.TrimSpace(req.FormValue("firstName"))
			inputLastName := strings.TrimSpace(req.FormValue("lastName"))
			inputMobile := strings.TrimSpace(req.FormValue("mobileNum"))
			inputPassword := strings.TrimSpace(req.FormValue("password"))

			// Validate first name input
			if validator.IsEmpty(inputFirstName) || !validator.IsValidName(inputFirstName) {
				ViewData.ValidateFirstName = false
			}
			if ViewData.ValidateFirstName {
				if c := strings.Compare(inputFirstName, ViewData.UserData.FirstName); c != 0 {
					ViewData.UserData.FirstName = inputFirstName
					edited = true
				}
			}
			// Validate last name input
			if validator.IsEmpty(inputLastName) || !validator.IsValidName(inputLastName) {
				ViewData.ValidateLastName = false
			}
			if ViewData.ValidateLastName {
				if c := strings.Compare(inputLastName, ViewData.UserData.LastName); c != 0 {
					ViewData.UserData.LastName = inputLastName
					edited = true
				}
			}
			// Validate mobile number input
			if ViewData.UserData.Role == enumPatient {
				if validator.IsEmpty(inputMobile) || !validator.IsMobileNumber(inputMobile) {
					ViewData.ValidateMobileNumber = false
				}
				if ViewData.ValidateMobileNumber {
					mobileNumber, _ := strconv.Atoi(inputMobile)
					if mobileNumber != ViewData.UserData.MobileNumber {
						ViewData.UserData.MobileNumber = mobileNumber
						edited = true
					}
				}
			}
			// Change Password
			if len(inputPassword) > 0 {
				// Matching of password entered
				if err := bcrypt.CompareHashAndPassword([]byte(ViewData.UserData.Password), []byte(inputPassword)); err != nil {
					// Different password
					bPassword, err := bcrypt.GenerateFromPassword([]byte(inputPassword), bcrypt.MinCost)
					if err != nil {
						logger.Error.Printf("%v: Error:", util.CurrFuncName(), err)
					} else {
						ViewData.UserData.Password = string(bPassword)
						edited = true
					}
				}
			}

			checkboxInput := req.FormValue("deleteChkBox")
			deleteChkBox, err := strconv.ParseBool(checkboxInput)
			if err != nil {
				deleteChkBox = false
			}

			// Validation completed
			if ViewData.ValidateFirstName && ViewData.ValidateLastName && ViewData.ValidateMobileNumber {
				if edited {
					user.UpdateUserData(copyUser, ViewData.UserData)
				}
				if deleteChkBox && !ViewData.UserData.IsDeleted {
					ViewData.UserData.IsDeleted = true
					user.DeleteUserData(ViewData.UserData)
					// Remove user for session if logged in
					deleteSessionByUsername(ViewData.UserData.Username)
				}
				if !deleteChkBox && ViewData.UserData.IsDeleted {
					ViewData.UserData.IsDeleted = false
					user.DeleteUserData(ViewData.UserData)
				}
				ViewData.Successful = true
			}
		}

		if err := tpl.ExecuteTemplate(res, "userEdit.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func userDeleteHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				util.CheckEncryption()
				logger.Panic.Println(err)
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
			(*userList).GetList(),
			false,
			false,
			"",
		}

		userObj := (*userList).FindByUsername(username)
		if userObj == nil {
			ViewData.ErrorDelete = true
			ViewData.ErrorDeleteMsg = "Error deleting user: " + username + ", user does not exist."
			logger.Error.Printf("%v: User does not exist: %v", util.CurrFuncName(), username)
			if err := tpl.ExecuteTemplate(res, "userList.gohtml", ViewData); err != nil {
				logger.Error.Println(err)
			}
			return
		}

		// Soft delete user
		userObj.IsDeleted = true
		ViewData.Successful = true
		// Delete user from JSON
		user.DeleteUserData(userObj)
		// Remove user for session if logged in
		deleteSessionByUsername(userObj.Username)
		logger.Info.Printf("%v: User [%v] deleted successfully.", util.CurrFuncName(), username)

		if err := tpl.ExecuteTemplate(res, "userList.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}

func sessionListHandler(userList *user.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		myUser, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		type SessionStruct struct {
			SessionID string
			Username  string
			Role      string
		}

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Sessions     []SessionStruct
		}{
			myUser,
			"Manage Session",
			"MS",
			nil,
		}

		for k, v := range mapSessions {
			userObj := (*userList).FindByUsername(v)
			ViewData.Sessions = append(ViewData.Sessions, SessionStruct{SessionID: k, Username: v, Role: userObj.Role})
		}

		// Process form submission
		if req.Method == http.MethodPost {
			if err := req.ParseForm(); err != nil {
				logger.Error.Printf("%v: Error:", util.CurrFuncName(), err)
			} else {
				// Loop through form
				for key, values := range req.Form {
					// Value equals to session ID
					for _, sessionID := range values {
						// Key equals to checkbox group
						if key == "sessionsDel" {
							delete(mapSessions, sessionID)
						}
					}
				}
			}
			http.Redirect(res, req, "/sessions", http.StatusSeeOther)
		}

		if err := tpl.ExecuteTemplate(res, "sessions.gohtml", ViewData); err != nil {
			logger.Error.Println(err)
		}
	}
}
