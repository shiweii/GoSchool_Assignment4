package main

import (
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
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	MobileNumber int    `json:"mobileNumber,omitempty"`
	IsDeleted    bool   `json:"isDeleted,omitempty"`
}

func newUser(username, password, role, firstName, lastName string, mobileNumber int) *User {
	return &User{
		Username:     username,
		Password:     password,
		Role:         role,
		FirstName:    firstName,
		LastName:     lastName,
		MobileNumber: mobileNumber,
		IsDeleted:    false,
	}
}

func usersHandler(userList **dll.DoublyLinkedlist) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		_, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		users := (**userList).GetList()

		ViewData := struct {
			Users          []interface{}
			Successful     bool
			ErrorDelete    bool
			ErrorDeleteMsg string
		}{
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
		selectedUser := ret.(*User)
		copyUser := newUser(selectedUser.Username, selectedUser.Password, selectedUser.Role, selectedUser.FirstName, selectedUser.LastName, selectedUser.MobileNumber)

		ViewData := struct {
			LoggedInUser         *User
			UserData             *User
			ValidateUsername     bool
			MessageUsername      string
			ValidateFirstName    bool
			ValidateLastName     bool
			ValidateMobileNumber bool
			ValidatePassword     bool
			Successful           bool
		}{
			myUser,
			selectedUser,
			true,
			"",
			true,
			true,
			true,
			true,
			false,
		}

		// process form submission
		if req.Method == http.MethodPost {
			var edited bool = false
			// Validate username input
			/*inputUsername := req.FormValue("username")
			if c := strings.Compare(inputUsername, selectedUser.Username); c != 0 {
				if len(inputUsername) == 0 {
					ViewData.ValidateUsername = false
					ViewData.MessageUsername = "Username is required."
				} else {
					ret := (**userList).FindByUsername(inputUsername)
					if ret != nil {
						ViewData.ValidateUsername = false
						ViewData.MessageUsername = "Enter username already exist, please enter another username"
					}
				}
				if ViewData.ValidateUsername {
					selectedUser.Username = inputUsername
					edited = true
				}
			}*/
			// Validate first name input
			inputFirstName := req.FormValue("firstName")
			if c := strings.Compare(inputFirstName, selectedUser.FirstName); c != 0 {
				if len(inputFirstName) == 0 {
					ViewData.ValidateFirstName = false
				}
				if ViewData.ValidateFirstName {
					selectedUser.FirstName = inputFirstName
					edited = true
				}
			}
			// Validate last name input
			inputLastName := req.FormValue("lastName")
			if c := strings.Compare(inputLastName, selectedUser.LastName); c != 0 {
				if len(inputLastName) == 0 {
					ViewData.ValidateLastName = false
				}
				if ViewData.ValidateLastName {
					selectedUser.LastName = inputLastName
					edited = true
				}
			}
			// Validate mobile number input
			inputMobile := req.FormValue("mobileNum")
			mobileNumber, _ := strconv.Atoi(inputMobile)
			if mobileNumber != selectedUser.MobileNumber {
				if len(inputMobile) == 0 {
					ViewData.ValidateMobileNumber = false
				}
				if !util.ValidateMobileNumber(mobileNumber) {
					ViewData.ValidateMobileNumber = false
				}
				if ViewData.ValidateMobileNumber {
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
			if ViewData.ValidateUsername && ViewData.ValidateFirstName && ViewData.ValidateLastName && ViewData.ValidateMobileNumber {
				if edited {
					updateUserData(copyUser, selectedUser)
					if myUser.Role == enumPatient {
						if copyUser.Username != selectedUser.Username {
							updateCookie(res, req, selectedUser.Username)
						}
					}

				}
				if deleteChkBox && !selectedUser.IsDeleted {
					selectedUser.IsDeleted = true
					deleteUserData(selectedUser)
				}
				if !deleteChkBox && selectedUser.IsDeleted {
					selectedUser.IsDeleted = false
					deleteUserData(selectedUser)
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
				var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)
				Error.Println(err)
				http.Redirect(res, req, "/", http.StatusInternalServerError)
				return
			}
		}()

		_, authFail, httpStatusNum := authenticationCheck(res, req, userList, true)
		if authFail {
			http.Redirect(res, req, "/", httpStatusNum)
			return
		}

		vars := mux.Vars(req)
		username := vars["username"]

		var users []interface{}

		ViewData := struct {
			Users          []interface{}
			Successful     bool
			ErrorDelete    bool
			ErrorDeleteMsg string
		}{
			users,
			false,
			false,
			"",
		}

		retUser := (**userList).FindByUsername(username).(*User)
		if retUser == nil {
			ViewData.ErrorDelete = true
			ViewData.ErrorDeleteMsg = "Error deleteing user: " + username + " user does not exist."
		} else {
			// Soft delete user
			retUser.IsDeleted = true
			ViewData.Successful = true
			deleteUserData(retUser)
		}
		ViewData.Users = (**userList).GetList()

		tpl.ExecuteTemplate(res, "userList.gohtml", ViewData)
	}
}

func getUserData() []*User {
	var users []*User
	JSONData, _ := ioutil.ReadFile(userData)
	err := json.Unmarshal(JSONData, &users)
	if err != nil {
		fmt.Println(err)
	}
	return users
}

func addUserDate(u *User) {
	var users []*User
	users = getUserData()
	users = append(users, u)
	JSONData, _ := json.MarshalIndent(users, "", " ")
	err := ioutil.WriteFile(userData, JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func updateUserData(oldUser *User, newUser *User) {
	var users []*User = getUserData()
	for k, v := range users {
		if reflect.DeepEqual(v, oldUser) {
			users[k] = newUser
		}
	}
	JSONData, _ := json.MarshalIndent(users, "", " ")
	err := ioutil.WriteFile(userData, JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func deleteUserData(delUser *User) {
	var users []*User = getUserData()
	for k, v := range users {
		if v.Username == delUser.Username {
			users[k] = delUser
		}
	}
	JSONData, _ := json.MarshalIndent(users, "", " ")
	err := ioutil.WriteFile(userData, JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func getDentistList(users []interface{}) []*User {
	var dentists []*User
	for _, v := range users {
		user := v.(*User)
		if user.Role == "dentist" {
			dentists = append(dentists, user)
		}
	}
	return dentists
}
