package user

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	ede "github.com/shiweii/encryptdecrypt"
	util "github.com/shiweii/utility"
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

func NewUser(username, password, role, firstName, lastName string, mobileNumber int) *User {
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

func GetEncryptedUserData() []*User {
	ede.DecryptFile(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	users := getUserData()
	ede.EncryptFile(util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
	return users
}

func getUserData() []*User {
	var users []*User
	JSONData, _ := ioutil.ReadFile(util.GetEnvVar("USER_DATA"))
	err := json.Unmarshal(JSONData, &users)
	if err != nil {
		fmt.Println(err)
	}
	return users
}

func AddUserDate(u *User) {
	ede.DecryptFile(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	var users []*User
	users = getUserData()
	users = append(users, u)
	JSONData, _ := json.MarshalIndent(users, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("USER_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
	ede.EncryptFile(util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
}

func UpdateUserData(oldUser *User, newUser *User) {
	ede.DecryptFile(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	var users []*User = getUserData()
	for k, v := range users {
		if reflect.DeepEqual(v, oldUser) {
			users[k] = newUser
		}
	}
	JSONData, _ := json.MarshalIndent(users, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("USER_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
	ede.EncryptFile(util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
}

func DeleteUserData(delUser *User) {
	ede.DecryptFile(util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	var users []*User = getUserData()
	for k, v := range users {
		if v.Username == delUser.Username {
			users[k] = delUser
		}
	}
	JSONData, _ := json.MarshalIndent(users, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("USER_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
	ede.EncryptFile(util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
}

func GetDentistList(users []interface{}) []*User {
	var dentists []*User
	for _, v := range users {
		user := v.(*User)
		if user.Role == "dentist" {
			dentists = append(dentists, user)
		}
	}
	return dentists
}
