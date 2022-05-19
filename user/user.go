// Package user is for user data storage and manipulation.
// User data are read and write to data/user.json file. The file is encrypted by default
package user

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/shiweii/cryptography"
	dll "github.com/shiweii/doublylinkedlist"
	util "github.com/shiweii/utility"
)

// User struct stores user data.
type User struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	MobileNumber int    `json:"mobileNumber,omitempty"`
	IsDeleted    bool   `json:"isDeleted,omitempty"`
}

// DoublyLinkedList extends from doublylinkedlist package for user related processing.
type DoublyLinkedList struct {
	*dll.DoublyLinkedList
}

// New will return a newly created instance of a user.
func New(username, password, role, firstName, lastName string, mobileNumber int) *User {
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

// GetEncryptedUserData will perform decryption and encryption on user JSON file.
func GetEncryptedUserData() []*User {
	cryptography.DecryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	users := getUserData()
	cryptography.EncryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
	return users
}

// getUserData open, read and unmarshal user data from JSON file.
func getUserData() []*User {
	var users []*User
	JSONData, _ := ioutil.ReadFile(util.GetEnvVar("USER_DATA"))
	err := json.Unmarshal(JSONData, &users)
	if err != nil {
		fmt.Println(err)
	}
	return users
}

// AddUserDate will decrypt, open, marshal, append and encrypt new user data into JSON file.
func AddUserDate(u *User) {
	cryptography.DecryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	var users []*User
	users = getUserData()
	users = append(users, u)
	JSONData, _ := json.MarshalIndent(users, "", " ")
	err := ioutil.WriteFile(util.GetEnvVar("USER_DATA"), JSONData, 0644)
	if err != nil {
		fmt.Println(err)
	}
	cryptography.EncryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
}

// UpdateUserData will decrypt, open, marshal, update and encrypt matching user data into JSON file.
func UpdateUserData(oldUser *User, newUser *User) {
	cryptography.DecryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	var users = getUserData()
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
	cryptography.EncryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
}

// DeleteUserData will decrypt, open, marshal, delete  and encrypt user data using username from JSON file.
func DeleteUserData(delUser *User) {
	cryptography.DecryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA_ENCRYPT"), util.GetEnvVar("USER_DATA"))
	var users = getUserData()
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
	cryptography.EncryptFile(util.GetEnvVar("KEY"), util.GetEnvVar("USER_DATA"), util.GetEnvVar("USER_DATA_ENCRYPT"))
}

// GetDentistList returns all elements in the linked list.
func (list *DoublyLinkedList) GetDentistList() []*User {
	var values []*User
	currentNode := list.GetHeadNode()
	for currentNode != nil {
		if currentNode.Value.(*User).Role == "dentist" {
			values = append(values, currentNode.Value.(*User))
		}
		currentNode = currentNode.Next
	}
	return values
}

// InsertionSort Sort elements using insertion sort using username as the majority of searches uses username
func (list *DoublyLinkedList) InsertionSort() {
	// Get first node
	var front = list.GetHeadNode()
	var back *dll.Node = nil
	for front != nil {
		// Get next node
		back = front.Next
		// Update node value when consecutive nodes are not sort
		for back != nil && back.Previous != nil {
			if back.Value.(*User).Username < back.Previous.Value.(*User).Username {
				// Modified node data
				list.swapData(back, back.Previous)
			}
			// Visit to previous node
			back = back.Previous
		}
		// Visit to next node
		front = front.Next
	}
}

// swapData swaps dara between two nodes
func (list *DoublyLinkedList) swapData(first, second *dll.Node) {
	value := first.Value
	first.Value = second.Value
	second.Value = value
}

// FindByUsername iterates and return element from sorted linked link by username.
func (list *DoublyLinkedList) FindByUsername(username string) *User {
	if len(username) > 0 {
		return list.recursiveBinarySearchByUsername(list.GetHeadNode(), list.GetTailNode(), username, list.GetSize())
	}
	return nil
}

// middleNode return the middle element within a given range of elements.
func middleNode(start *dll.Node, mid int) *dll.Node {
	if start == nil {
		return nil
	}
	for i := 1; i < mid; i++ {
		start = start.Next
	}
	return start
}

// recursiveBinarySearchByUsername performs recursive binary search on sorted linked list.
func (list *DoublyLinkedList) recursiveBinarySearchByUsername(firstNode *dll.Node, lastNode *dll.Node, value string, size int) *User {
	if firstNode == nil || lastNode == nil {
		return nil
	}
	if firstNode.Value.(*User).Username > lastNode.Value.(*User).Username {
		return nil
	} else {
		mid := size / 2
		midNode := middleNode(firstNode, mid)
		if midNode.Value.(*User).Username == value {
			return midNode.Value.(*User)
		} else {
			if value < midNode.Value.(*User).Username {
				return list.recursiveBinarySearchByUsername(firstNode, midNode.Previous, value, mid)
			} else {
				return list.recursiveBinarySearchByUsername(midNode.Next, lastNode, value, mid)
			}
		}
	}
}

// SearchByMobileNumber iterates and return element from sorted linked link by mobile number.
func (list *DoublyLinkedList) SearchByMobileNumber(mobileNum int) *User {
	//var ret interface{}
	ret := list.recursiveSeqSearchByMobileNumber(list.GetHeadNode(), mobileNum)
	return ret
}

// recursiveSeqSearchByMobileNumber performs recursive sequential search on linked list.
func (list *DoublyLinkedList) recursiveSeqSearchByMobileNumber(node *dll.Node, value int) *User {
	if node == nil {
		return nil
	} else {
		if node.Value.(*User).MobileNumber == value {
			return node.Value.(*User)
		}
		return list.recursiveSeqSearchByMobileNumber(node.Next, value)
	}
}
