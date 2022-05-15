package main

import (
	"fmt"
	"net/http"
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

func indexHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Panic.Println(err)
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

		err := tpl.ExecuteTemplate(res, "index.gohtml", ViewData)
		if err != nil {
			logger.Info.Printf("indexHandler: %v", err)
		}
	}
}

func signupHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
				logger.Panic.Println(err)
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
			LoggedInUser         *user.User
			PageTitle            string
			ValidateFirstName    bool
			ValidateLastName     bool
			ValidateUserName     bool
			UserNameErrorMsg     string
			ValidatePassword     bool
			ValidateMobileNumber bool
		}{
			nil,
			"Sign Up",
			true,
			true,
			true,
			"",
			true,
			true,
		}

		// process form submission
		if req.Method == http.MethodPost {
			// get form values
			inputUserName := strings.TrimSpace(req.FormValue("username"))
			inputPassword := strings.TrimSpace(req.FormValue("password"))
			inputFirstName := strings.TrimSpace(req.FormValue("firstname"))
			inputLastName := strings.TrimSpace(req.FormValue("lastname"))
			inputMobile := strings.TrimSpace(req.FormValue("mobileNum"))

			logger.Info.Printf("signupHandler: Username: %v, FirstName: %v, LastName: %v, MobileNumber: %v", inputUserName, inputFirstName, inputLastName, inputMobile)

			//Validate Fields
			if validator.IsEmpty(inputFirstName) || !validator.IsValidName(inputFirstName) {
				ViewData.ValidateFirstName = false
			}
			if validator.IsEmpty(inputFirstName) || !validator.IsValidName(inputLastName) {
				ViewData.ValidateLastName = false
			}
			if validator.IsEmpty(inputUserName) || !validator.IsValidUsername(inputUserName) {
				ViewData.ValidateUserName = false
				ViewData.UserNameErrorMsg = "Please enter a valid username."
			}
			if validator.IsEmpty(inputPassword) || !validator.IsValidPassword(inputPassword) {
				ViewData.ValidatePassword = false
			}
			if validator.IsEmpty(inputMobile) || !validator.IsMobileNumber(inputMobile) {
				ViewData.ValidateMobileNumber = false
			}

			// If all validations are true
			if ViewData.ValidateFirstName && ViewData.ValidateLastName && ViewData.ValidateUserName && ViewData.ValidatePassword && ViewData.ValidateMobileNumber {
				// check if username exist/ taken
				userItf := (**userList).FindByUsername(inputUserName)
				if userItf != nil {
					ViewData.ValidateUserName = false
					ViewData.UserNameErrorMsg = "Please enter a valid username."
				}

				if ViewData.ValidateUserName {
					// create session
					id := uuid.NewV4()
					myCookie := &http.Cookie{
						Name:  "myCookie",
						Value: id.String(),
					}
					http.SetCookie(res, myCookie)
					mapSessions[myCookie.Value] = inputUserName

					bPassword, err := bcrypt.GenerateFromPassword([]byte(inputPassword), bcrypt.MinCost)
					if err != nil {
						logger.Info.Printf("signupHandler: %v", err)
						http.Error(res, "Internal server error", http.StatusInternalServerError)
						return
					}

					myUser.Username = inputUserName
					myUser.Password = string(bPassword)
					myUser.FirstName = inputFirstName
					myUser.LastName = inputLastName
					myUser.Role = enumPatient

					mobileNum, _ := strconv.Atoi(inputMobile)
					myUser.MobileNumber = mobileNum

					// Add into linklist and JSON
					err = (**userList).Add(&myUser)
					if err != nil {
						logger.Error.Println(err)
					} else {
						(**userList).InsertionSort()
						user.AddUserDate(&myUser)
					}

					// redirect to patient landing page
					http.Redirect(res, req, "/appointments", http.StatusSeeOther)
					return
				}
			}
		}
		err := tpl.ExecuteTemplate(res, "signup.gohtml", ViewData)
		if err != nil {
			logger.Info.Printf("signupHandler: %v", err)
		}
	}
}

func loginHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
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

			var (
				userItf interface{}
				userObj *user.User
			)

			// Retrieve form input and remove empty spaces
			inputUserName := strings.TrimSpace(req.FormValue("username"))
			inputPassword := strings.TrimSpace(req.FormValue("password"))

			//Validate Fields
			if validator.IsEmpty(inputUserName) || !validator.IsValidUsername(inputUserName) {
				ViewData.LoginFail = true
			}

			// check if user exist with username
			if !ViewData.LoginFail {
				userItf = (**userList).FindByUsername(inputUserName)
				if userItf == nil {
					ViewData.LoginFail = true
					logger.Info.Println("loginHandler: Login fail. user:", inputUserName)
				}
			}

			// Check if user is deleted
			if !ViewData.LoginFail {
				userObj = userItf.(*user.User)
				if userObj.IsDeleted {
					ViewData.LoginFail = true
					logger.Info.Println("loginHandler: Login fail. user:", userObj.Username)
				}
			}

			// Matching of password entered
			if !ViewData.LoginFail {
				err := bcrypt.CompareHashAndPassword([]byte(userObj.Password), []byte(inputPassword))
				if err != nil {
					ViewData.LoginFail = true
					logger.Info.Println("loginHandler: Login fail. user:", userObj.Username)
				}
			}

			if !ViewData.LoginFail {
				id := uuid.NewV4()
				myCookie := &http.Cookie{
					Name:    "myCookie",
					Expires: time.Now().AddDate(0, 0, 1),
					Value:   id.String(),
				}
				go killOtherSession(myCookie)
				logger.Info.Printf("loginHandler: Login successful. user:%v", userObj.Username)
				http.SetCookie(res, myCookie)
				mapSessions[myCookie.Value] = userObj.Username
				http.Redirect(res, req, "/appointments", http.StatusSeeOther)
				return
			}
		}
		err := tpl.ExecuteTemplate(res, "login.gohtml", ViewData)
		if err != nil {
			logger.Error.Printf("loginHandler: %v", err)
		}
	}
}

func logoutHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !alreadyLoggedIn(req, userList) {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}
		myCookie, _ := req.Cookie("myCookie")
		// Get username
		username, _ := mapSessions[myCookie.Value]
		logger.Info.Println("logoutHandler: Logout... user:", username)
		// delete the session
		delete(mapSessions, myCookie.Value)
		// Expire the cookie
		myCookie = &http.Cookie{
			Path:    "/",
			Name:    "myCookie",
			MaxAge:  -1,
			Expires: time.Now().Add(-100 * time.Hour),
		}
		http.SetCookie(res, myCookie)
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func appointmentListHandler(userList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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
				appointments = (**appointmentTree).GetUpComingAppointments(myUser, enumPatient)
			} else {
				appointments = (**appointmentTree).GetAllAppointments(myUser, enumPatient)
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

			// Data conversation
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
		err := tpl.ExecuteTemplate(res, "appointmentList.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentSearchHandler(userList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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
		err := tpl.ExecuteTemplate(res, "appointmentSearch.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentCreateHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

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

		err := tpl.ExecuteTemplate(res, "appointmentCreate_step1.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentCreatePart2Handler(userList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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
			inputDate := req.FormValue("appDate")
			appointmentDate, err := time.Parse("2006-01-02", inputDate)
			if err == nil {
				ViewData.Sessions = app.GetDentistAvailability(appointmentSessionList, appointmentTree, appointmentDate, ViewData.Dentist)
				ViewData.SelectedDate = appointmentDate.Format("2006-01-02")
			}
		}
		err := tpl.ExecuteTemplate(res, "appointmentCreate_step2.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentCreateConfirmHandler(userList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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
			var id = util.GenerateID()
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

		err = tpl.ExecuteTemplate(res, "appointmentCreateConfirm.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentEditHandler(userList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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
			inputDate := req.FormValue("appDate")
			inputDentist := req.FormValue("appDentist")
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

		err := tpl.ExecuteTemplate(res, "appointmentEdit.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentEditConfirmHandler(userList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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

		// Decode and retrieve URL query string
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
			"Confirm Appointment Change",
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
			editedAppointment := app.NewAppointment(appointment.ID, appointment.Patient.(*user.User).Username, appointment.Dentist.(*user.User).Username, appointment.Date, appointment.Session)
			// If there are no changes made to appointment date, update dentist and/or session value
			if appointment.Date == dateReq {
				if appointment.Dentist.(*user.User).Username != dentist.Username {
					appointment.Dentist = dentist
					editedAppointment.Dentist = dentist.Username
				}
				if appointment.Session != session {
					appointment.Session = session
					editedAppointment.Session = session
				}
				ViewData.Successful = true
				app.UpdateAppointmentData(oldAppointment, editedAppointment)
			} else {
				// If changes made to appointment date, added a new appointment and delete old appointment
				var id = util.GenerateID()
				(**appointmentTree).Add(id, dateReq, session, dentist, appointment.Patient)
				appointmentData := app.NewAppointment(id, appointment.Patient.(*user.User).Username, dentist.Username, dateReq, session)
				app.AddAppointmentData(appointmentData)
				err := (**appointmentTree).Remove(appointment)
				if err != nil {
					logger.Error.Println(err)
				} else {
					app.DeleteAppointmentData(oldAppointment.ID)
					ViewData.Successful = true
				}
			}
		}
		err := tpl.ExecuteTemplate(res, "appointmentEditConfirm.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func appointmentDeleteHandler(userList, appointmentSessionList **dll.DoublyLinkedList, appointmentTree **bst.BinarySearchTree) http.HandlerFunc {
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
			err := (**appointmentTree).Remove(appointment)
			if err != nil {
				logger.Error.Println(err)
			} else {
				app.DeleteAppointmentData(appointment.ID)
				ViewData.Successful = true
			}
		}

		err := tpl.ExecuteTemplate(res, "appointmentDelete.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func userListHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
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

		err := tpl.ExecuteTemplate(res, "userList.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func userEditHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
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

		var selectedUser *user.User
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

		ret := (**userList).FindByUsername(username)
		if ret == nil {
			err := tpl.ExecuteTemplate(res, "userEdit.gohtml", ViewData)
			if err != nil {
				logger.Error.Println(err)
			}
			return
		}

		selectedUser = ret.(*user.User)
		ViewData.UserData = selectedUser
		copyUser := user.NewUser(selectedUser.Username, selectedUser.Password, selectedUser.Role, selectedUser.FirstName, selectedUser.LastName, selectedUser.MobileNumber)

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
				if c := strings.Compare(inputFirstName, selectedUser.FirstName); c != 0 {
					selectedUser.FirstName = inputFirstName
					edited = true
				}
			}
			// Validate last name input
			if validator.IsEmpty(inputLastName) || !validator.IsValidName(inputLastName) {
				ViewData.ValidateLastName = false
			}
			if ViewData.ValidateLastName {
				if c := strings.Compare(inputLastName, selectedUser.LastName); c != 0 {
					selectedUser.LastName = inputLastName
					edited = true
				}
			}
			// Validate mobile number input
			if selectedUser.Role == enumPatient {
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
			}
			// Change Password
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

			checkboxInput := req.FormValue("deleteChkBox")
			deleteChkBox, err := strconv.ParseBool(checkboxInput)
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
			logger.Error.Println(err)
		}
	}
}

func userDeleteHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ede.CheckEncryption(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
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
			ViewData.ErrorDeleteMsg = "Error deleting user: " + username + " user does not exist."
		} else {
			// Soft delete user
			retUser.IsDeleted = true
			ViewData.Successful = true
			// Delete user for Linked-list
			user.DeleteUserData(retUser)
			// Remove user for sesscion if active
		}
		ViewData.Users = (**userList).GetList()

		err := tpl.ExecuteTemplate(res, "userList.gohtml", ViewData)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}

func sessionListHandler(userList **dll.DoublyLinkedList) http.HandlerFunc {
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

		var sessions []SessionStruct

		ViewData := struct {
			LoggedInUser *user.User
			PageTitle    string
			CurrentPage  string
			Sessions     []SessionStruct
		}{
			myUser,
			"Manage Session",
			"MS",
			sessions,
		}

		for k, v := range mapSessions {
			userObj := (**userList).FindByUsername(v).(*user.User)
			ViewData.Sessions = append(ViewData.Sessions, SessionStruct{SessionID: k, Username: v, Role: userObj.Role})
		}

		// Process form submission
		if req.Method == http.MethodPost {
			err := req.ParseForm()
			if err != nil {
				logger.Error.Println(err)
			}
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
			logger.Error.Println(err)
		}
	}
}
