// Package appointment is for application data storage and manipulation.
// Application data are read and write to data/application.json file.
package appointment

import (
	"encoding/json"
	"fmt"
	bst "github.com/shiweii/binarysearchtree"
	dll "github.com/shiweii/doublylinkedlist"
	"github.com/shiweii/logger"
	"github.com/shiweii/user"
	util "github.com/shiweii/utility"
	"io/ioutil"
	"reflect"
	"time"
)

// Appointment struct stores application data.
type Appointment struct {
	ID      int         `json:"id"`
	Dentist interface{} `json:"dentist"`
	Patient interface{} `json:"patient"`
	Date    string      `json:"date"`
	Session int         `json:"session"`
}

// AppSession struct stores application session data.
type AppSession struct {
	Num       int
	StartTime string
	EndTime   string
	Available bool
}

// BinarySearchTree Extends binarysearchtree package for application related processing.
type BinarySearchTree struct {
	*bst.BinarySearchTree
}

// New will return a newly created instance of an appointment.
func New(id int, patient, dentist interface{}, date string, session int) *Appointment {
	return &Appointment{
		ID:      id,
		Dentist: dentist,
		Patient: patient,
		Date:    date,
		Session: session,
	}
}

// GetAppointmentData will open, read and unmarshal application data from JSON file.
func GetAppointmentData() []*Appointment {
	var appointments []*Appointment
	JSONData, _ := ioutil.ReadFile(util.GetEnvVar("APPOINTMENT_DATA"))
	err := json.Unmarshal(JSONData, &appointments)
	if err != nil {
		fmt.Println(err)
	}
	return appointments
}

// AddAppointmentData will open, marshal and append new application data into JSON file.
func AddAppointmentData(a *Appointment) {
	var appointments []*Appointment
	appointments = GetAppointmentData()
	appointments = append(appointments, a)
	JSONData, _ := json.MarshalIndent(appointments, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("APPOINTMENT_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// UpdateAppointmentData will open, marshal and update matching application data into JSON file.
func UpdateAppointmentData(oldAppointment *Appointment, editedAppointment *Appointment) {
	var appointments = GetAppointmentData()
	for k, v := range appointments {
		if reflect.DeepEqual(v, oldAppointment) {
			appointments[k] = editedAppointment
		}
	}
	JSONData, _ := json.MarshalIndent(appointments, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("APPOINTMENT_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// DeleteAppointmentData will open, marshal and delete application data using application id from JSON file.
func DeleteAppointmentData(id int) {
	var appointments []*Appointment
	var idx int
	appointments = GetAppointmentData()
	for k, v := range appointments {
		if v.ID == id {
			idx = k
		}
	}
	appointments = append(appointments[:idx], appointments[idx+1:]...)
	JSONData, _ := json.MarshalIndent(appointments, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("APPOINTMENT_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// CreateNewAppointment run as Go routine to block users from booking the same dentist on the same date and session.
// Application Data will then be inserted into the binary search tree then append into JSON for persistence storage.
func CreateNewAppointment(id int, date string, session int, dentist *user.User, patient *user.User, appointmentTree *BinarySearchTree, chn chan bool) {
	var sessionBooked = false
	// Check if appointment is booked
	appointments := (*appointmentTree).GetAppointmentByDate(date, "dentist", dentist)
	for _, v := range appointments {
		if v.Session == session {
			sessionBooked = true
			chn <- false
		}
	}
	// If slot is not booked, proceed.
	if !sessionBooked {
		appointment := New(id, patient, dentist, date, session)
		appointmentTree.Add(date, appointment)
		appointmentJSON := New(id, patient.Username, dentist.Username, date, session)
		AddAppointmentData(appointmentJSON)
		chn <- true
	}
}

// DeleteAppointment deletes an appointment from both binary search tree and JSON
func (appBst *BinarySearchTree) DeleteAppointment(application *Appointment) error {
	var node *bst.BinaryNode
	appID := application.ID
	getBinaryNodeTraversal(appBst.GetRootNode(), application, &node)
	err := appBst.Remove(node)
	if err == nil {
		// Delete from JSON
		DeleteAppointmentData(appID)
	}
	return err
}

// getBinaryNodeTraversal traverse the binary search tree and return a binary node.
func getBinaryNodeTraversal(t *bst.BinaryNode, application *Appointment, node **bst.BinaryNode) *bst.BinaryNode {
	if t != nil {
		getBinaryNodeTraversal(t.Left, application, node)
		if reflect.DeepEqual(t.Data, application) {
			*node = t
		}
		getBinaryNodeTraversal(t.Right, application, node)
	}
	return nil
}

// GetAllAppointments returns all elements based on user role.
func (appBst *BinarySearchTree) GetAllAppointments(user *user.User, role string) []*Appointment {
	var list []*Appointment
	oldDate := time.Now().AddDate(-100, 0, 0)
	appBst.searchAppointments(appBst.GetRootNode(), oldDate.Format("2006-01-02"), user, role, &list)
	return list
}

// GetUpComingAppointments returns all elements based on user role.
// Only return all elements which date are grater than time.Now()
func (appBst *BinarySearchTree) GetUpComingAppointments(user *user.User, role string) []*Appointment {
	var list []*Appointment
	currentTime := time.Now()
	appBst.searchAppointments(appBst.GetRootNode(), currentTime.Format("2006-01-02"), user, role, &list)
	return list
}

// searchAppointments performs InOrder Traversal to illiterate through the binary search tree.
func (appBst *BinarySearchTree) searchAppointments(t *bst.BinaryNode, date string, searchUser *user.User, role string, list *[]*Appointment) []*Appointment {
	if t != nil {
		appBst.searchAppointments(t.Left, date, searchUser, role, list)
		if t.Key >= date {
			if role == "dentist" {
				if t.Data.(*Appointment).Dentist.(*user.User) == searchUser {
					*list = append(*list, t.Data.(*Appointment))
				}
			} else if role == "patient" {
				if t.Data.(*Appointment).Patient.(*user.User) == searchUser {
					*list = append(*list, t.Data.(*Appointment))
				}
			} else {
				*list = append(*list, t.Data.(*Appointment))
			}
		}
		appBst.searchAppointments(t.Right, date, searchUser, role, list)
	}
	return *list
}

// SearchAllByField returns all elements based on selected field.
func (appBst *BinarySearchTree) SearchAllByField(field string, value interface{}, channel chan []*Appointment) {
	var list []*Appointment
	appBst.searchInOrderTraversal(appBst.GetRootNode(), field, value, &list)
	channel <- list
}

// searchInOrderTraversal performs and return elements using InOrder Traversal to illiterate through the binary search tree.
func (appBst *BinarySearchTree) searchInOrderTraversal(t *bst.BinaryNode, field string, value interface{}, list *[]*Appointment) []*Appointment {
	if t != nil {
		appBst.searchInOrderTraversal(t.Left, field, value, list)
		switch field {
		case "date":
			if t.Key == value {
				*list = append(*list, t.Data.(*Appointment))
			}
		case "patient":
			if t.Data.(*Appointment).Patient.(*user.User) == value.(*user.User) {
				*list = append(*list, t.Data.(*Appointment))
			}
		case "dentist":
			if t.Data.(*Appointment).Dentist.(*user.User) == value.(*user.User) {
				*list = append(*list, t.Data.(*Appointment))
			}
		case "session":
			if t.Data.(*Appointment).Session == value.(int) {
				*list = append(*list, t.Data.(*Appointment))
			}
		}
		appBst.searchInOrderTraversal(t.Right, field, value, list)
	}
	return *list
}

// GetDuplicate get duplicates element in a slice
func GetDuplicate(list []*Appointment, count int) []*Appointment {

	var temp []*Appointment
	duplicateFrequency := make(map[*Appointment]int)

	for _, item := range list {
		// check if the item/element exist in the duplicate_frequency map
		_, exist := duplicateFrequency[item]
		if exist {
			// increase counter by 1 if already in the map
			duplicateFrequency[item] += 1
		} else {
			// else start counting from 1
			duplicateFrequency[item] = 1
		}
	}
	for v, n := range duplicateFrequency {
		if n == count {
			temp = append(temp, v)
		}
	}
	return temp
}

// GetDentistAvailability retrieve all dentist's appointment by date and set availability flag
func GetDentistAvailability(appointmentSessionList **dll.DoublyLinkedList, appointmentTree *BinarySearchTree, appointmentDate time.Time, Dentist *user.User) []AppSession {
	var sessionList []AppSession
	appointments := (*appointmentTree).GetAppointmentByDate(appointmentDate.Format("2006-01-02"), Dentist.Role, Dentist)
	retSessionList := (**appointmentSessionList).GetList()
	// Loop Session list and set dentist availability
	for _, v := range retSessionList {
		session := v.(AppSession)
		for _, data := range appointments {
			if data.Session == session.Num {
				session.Available = false
			}
		}
		sessionList = append(sessionList, session)
	}
	return sessionList
}

// GetAppointmentByDate returns a list of elements based on date field.
func (appBst *BinarySearchTree) GetAppointmentByDate(date, role string, searchUser *user.User) []*Appointment {
	defer func() {
		if r := recover(); r != nil {
			logger.Panic.Printf("panic, recovered value: %v\n", r)
		}
	}()
	var list []*Appointment
	appBst.searchAppointmentByDate(appBst.GetRootNode(), date, role, searchUser, &list)
	return list
}

// searchAppointmentByDate performs a binary search on the binary search tree.
func (appBst *BinarySearchTree) searchAppointmentByDate(t *bst.BinaryNode, date, role string, searchUser *user.User, list *[]*Appointment) []*Appointment {
	if t == nil {
		return nil
	} else {
		if t.Key == date {
			if t.Right != nil {
				if role == "dentist" {
					if t.Data.(*Appointment).Dentist.(*user.User) == searchUser {
						*list = append(*list, t.Data.(*Appointment))
					}
				}
				if role == "patient" {
					if t.Data.(*Appointment).Patient.(*user.User) == searchUser {
						*list = append(*list, t.Data.(*Appointment))
					}
				}
				return appBst.searchAppointmentByDate(t.Right, date, role, searchUser, list)
			} else {
				if role == "dentist" {
					if t.Data.(*Appointment).Dentist.(*user.User) == searchUser {
						*list = append(*list, t.Data.(*Appointment))
					}
				}
				if role == "patient" {
					if t.Data.(*Appointment).Patient.(*user.User) == searchUser {
						*list = append(*list, t.Data.(*Appointment))
					}
				}
				return *list
			}
		} else {
			if t.Key > date {
				return appBst.searchAppointmentByDate(t.Left, date, role, searchUser, list)
			} else {
				return appBst.searchAppointmentByDate(t.Right, date, role, searchUser, list)
			}
		}
	}
}

// GetAppointmentByID returns binary node based on id field
func (appBst *BinarySearchTree) GetAppointmentByID(id int) *Appointment {
	var result *Appointment
	SearchAppointmentByID(appBst.GetRootNode(), id, &result)
	return result
}

// SearchAppointmentByID performs InOrder Traversal to search for an element based on application ID
func SearchAppointmentByID(t *bst.BinaryNode, id int, result **Appointment) {
	if t != nil {
		SearchAppointmentByID(t.Left, id, result)
		if t.Data.(*Appointment).ID == id {
			*result = t.Data.(*Appointment)
		}
		SearchAppointmentByID(t.Right, id, result)
	}
}
