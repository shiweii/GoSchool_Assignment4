package appointment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	bst "github.com/shiweii/binarysearchtree"
	dll "github.com/shiweii/doublylinkedlist"
	"github.com/shiweii/user"
	util "github.com/shiweii/utility"
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

func NewAppointment(id int, pateint, dentist, date string, session int) *Appointment {
	return &Appointment{
		ID:      id,
		Patient: pateint,
		Dentist: dentist,
		Date:    date,
		Session: session,
	}
}

func GetAppointmentData() []*Appointment {
	var appointments []*Appointment
	JSONData, _ := ioutil.ReadFile(util.GetEnvVar("APPOINTMENT_DATA"))
	err := json.Unmarshal(JSONData, &appointments)
	if err != nil {
		fmt.Println(err)
	}
	return appointments
}

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

func UpdateAppointmentData(oldAppointment *Appointment, editiedAppointment *Appointment) {
	var appointments []*Appointment = GetAppointmentData()
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

func DeleteAppointmentData(id int) {
	var appointments []*Appointment
	var idx int
	appointments = GetAppointmentData()
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

// Function to get duplicates from search results
func GetDuplicate(list []*bst.BinaryNode, count int) []*bst.BinaryNode {

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

// Run as Go routine to block users from booking the same dentist on the same date and session
func CreateNewAppointment(id int, date string, session int, dentist interface{}, pateint *user.User, appointmentTree **bst.BinarySearchTree, chn chan bool) {
	var sessionBooked bool = false
	// Check if appointment is booked
	appointments := (**appointmentTree).GetAppointmentByDate(date, "dentist", dentist)
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

func GetDentistAvailability(appointmentSessionList **dll.DoublyLinkedlist, appointmentTree **bst.BinarySearchTree, appointmentDate time.Time, Dentist *user.User) []AppointmentSession {
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
