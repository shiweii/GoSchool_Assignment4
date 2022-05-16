// Package binarysearchtree implements a binary search tree.
package binarysearchtree

import (
	"errors"
	"time"

	"github.com/shiweii/logger"
)

// BinaryNode is an element within the binary search tree.
type BinaryNode struct {
	ID      int
	Dentist interface{}
	Patient interface{}
	Date    string
	Session int
	left    *BinaryNode
	right   *BinaryNode
}

// BinarySearchTree  holds elements of the binary search tree.
type BinarySearchTree struct {
	root *BinaryNode
}

// New will return a newly created instance of a binary search tree.
func New() *BinarySearchTree {
	bst := &BinarySearchTree{nil}
	return bst
}

// Add wrapper function to added new element into the binary search tree.
func (bst *BinarySearchTree) Add(id int, date string, session int, dentist interface{}, patient interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logger.Panic.Printf("panic, recovered value: %v\n", r)
		}
	}()
	bst.insertNode(&bst.root, date, session, dentist, patient, id)
}

// insertNode inserts a new binary node into the binary search tree.
func (bst *BinarySearchTree) insertNode(t **BinaryNode, date string, session int, dentist interface{}, patient interface{}, id int) {
	if (*t) == nil {
		newNode := &BinaryNode{id, dentist, patient, date, session, nil, nil}
		*t = newNode
	} else {
		if date < (*t).Date {
			bst.insertNode(&(*t).left, date, session, dentist, patient, id) // de-referencing
		} else {
			bst.insertNode(&(*t).right, date, session, dentist, patient, id) // de-referencing
		}
	}
}

// GetAppointmentByID returns binary node based on id field
func (bst *BinarySearchTree) GetAppointmentByID(id int) *BinaryNode {
	var result *BinaryNode
	SearchAppointmentByID(bst.root, id, &result)
	return result
}

// SearchAppointmentByID performs InOrder Traversal to search for an element based on application ID
func SearchAppointmentByID(t *BinaryNode, id int, result **BinaryNode) {
	if t != nil {
		SearchAppointmentByID(t.left, id, result)
		if t.ID == id {
			*result = t
		}
		SearchAppointmentByID(t.right, id, result)
	}
}

// findSuccessor find and return binary nodes next smaller value
func (bst *BinarySearchTree) findSuccessor(t *BinaryNode) *BinaryNode {
	for t.right != nil { // Find node on extreme right
		t = t.right
	}
	return t
}

// removeNode removes a node from the binary search tree base on the follow cases
// Case 1, node to be deleted has 0 child (is a leaf)
// Case 2, node to be deleted has 1 child
// Case 3, node to be deleted has 2 children
func (bst *BinarySearchTree) removeNode(t **BinaryNode, removeNode *BinaryNode) (*BinaryNode, error) {
	if *t == nil {
		return nil, errors.New("error: tree is empty")
	} else if removeNode.Date < (*t).Date {
		(*t).left, _ = bst.removeNode(&(*t).left, removeNode)
	} else if removeNode.Date > (*t).Date {
		(*t).right, _ = bst.removeNode(&(*t).right, removeNode)
	} else {
		if (*t).left == nil {
			return (*t).right, nil
		} else if (*t).right == nil {
			return (*t).left, nil
		} else { // 3rd case of 2 children
			*t = bst.findSuccessor((*t).left)
			removeNode = *t
			(*t).left, _ = bst.removeNode(&(*t).left, removeNode)
		}
	}
	return *t, nil
}

// Remove wrapper function to remove application from the binary search tree.
func (bst *BinarySearchTree) Remove(removeNode *BinaryNode) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Panic.Printf("panic, recovered value: %v\n", r)
		}
	}()
	bst.root, _ = bst.removeNode(&bst.root, removeNode)
	return nil
}

// GetSize returns the number of nodes stored in the binary search tree.
func (bst *BinarySearchTree) GetSize() int {
	return size(bst.root)
}

// size performs recursive call to get the total number of nodes in the binary search tree.
func size(node *BinaryNode) int {
	if node == nil {
		return 0
	} else {
		return size(node.left) + 1 + size(node.right)
	}
}

// Contains searches all elements in the binary search tree for any matches.
func (bst *BinarySearchTree) Contains(date string, session int, dentist interface{}, patient interface{}) bool {
	newNode := BinaryNode{0, dentist, patient, date, session, nil, nil}
	return containsTraversal(bst.root, newNode)
}

// containsTraversal performs InOrder Traversal to match any elements stored in the binary search tree.
func containsTraversal(t *BinaryNode, n BinaryNode) bool {
	if t != nil {
		containsTraversal(t.left, n)
		if t.Patient == n.Patient && t.Dentist == n.Dentist && t.Date == n.Date && t.Session == n.Session {
			return true
		}
		containsTraversal(t.right, n)
	}
	return false
}

// GetAppointmentByDate returns a list of elements based on date field.
func (bst *BinarySearchTree) GetAppointmentByDate(date, role string, searchInterface interface{}) []*BinaryNode {
	defer func() {
		if r := recover(); r != nil {
			logger.Panic.Printf("panic, recovered value: %v\n", r)
		}
	}()
	var list []*BinaryNode
	bst.searchAppointmentByDate(bst.root, date, role, searchInterface, &list)
	return list
}

// searchAppointmentByDate performs a binary search on the binary search tree.
func (bst *BinarySearchTree) searchAppointmentByDate(t *BinaryNode, date, role string, searchInterface interface{}, list *[]*BinaryNode) []*BinaryNode {
	if t == nil {
		return nil
	} else {
		if t.Date == date {
			if t.right != nil {
				if role == "dentist" {
					if t.Dentist == searchInterface {
						*list = append(*list, t)
					}
				}
				if role == "patient" {
					if t.Patient == searchInterface {
						*list = append(*list, t)
					}
				}
				return bst.searchAppointmentByDate(t.right, date, role, searchInterface, list)
			} else {
				if role == "dentist" {
					if t.Dentist == searchInterface {
						*list = append(*list, t)
					}
				}
				if role == "patient" {
					if t.Patient == searchInterface {
						*list = append(*list, t)
					}
				}
				return *list
			}
		} else {
			if t.Date > date {
				return bst.searchAppointmentByDate(t.left, date, role, searchInterface, list)
			} else {
				return bst.searchAppointmentByDate(t.right, date, role, searchInterface, list)
			}
		}
	}
}

// GetAllAppointments returns all elements based on user role.
func (bst *BinarySearchTree) GetAllAppointments(searchInterface interface{}, role string) []*BinaryNode {
	var list []*BinaryNode
	oldDate := time.Now().AddDate(-100, 0, 0)
	bst.searchAppointments(bst.root, oldDate.Format("2006-01-02"), searchInterface, role, &list)
	return list
}

// GetUpComingAppointments returns all elements based on user role.
// Only return all elements which date are grater than time.Now()
func (bst *BinarySearchTree) GetUpComingAppointments(searchInterface interface{}, role string) []*BinaryNode {
	var list []*BinaryNode
	currentTime := time.Now()
	bst.searchAppointments(bst.root, currentTime.Format("2006-01-02"), searchInterface, role, &list)
	return list
}

// containsTraversal performs InOrder Traversal to illiterate through the binary search tree.
func (bst *BinarySearchTree) searchAppointments(t *BinaryNode, date string, searchInterface interface{}, role string, list *[]*BinaryNode) []*BinaryNode {
	if t != nil {
		bst.searchAppointments(t.left, date, searchInterface, role, list)
		if t.Date >= date {
			if role == "dentist" {
				if t.Dentist == searchInterface {
					*list = append(*list, t)
				}
			} else if role == "patient" {
				if t.Patient == searchInterface {
					*list = append(*list, t)
				}
			} else {
				*list = append(*list, t)
			}
		}
		bst.searchAppointments(t.right, date, searchInterface, role, list)
	}
	return *list
}

// SearchAllByField returns all elements based on selected field.
func (bst *BinarySearchTree) SearchAllByField(field string, value interface{}, channel chan []*BinaryNode) {
	var list []*BinaryNode
	bst.searchInOrderTraversal(bst.root, field, value, &list)
	channel <- list
}

// SearchAllByField performs and return elements using InOrder Traversal to illiterate through the binary search tree.
func (bst *BinarySearchTree) searchInOrderTraversal(t *BinaryNode, field string, value interface{}, list *[]*BinaryNode) []*BinaryNode {
	if t != nil {
		bst.searchInOrderTraversal(t.left, field, value, list)
		switch field {
		case "date":
			if t.Date == value {
				*list = append(*list, t)
			}
		case "patient":
			if t.Patient == value {
				*list = append(*list, t)
			}
		case "dentist":
			if t.Dentist == value {
				*list = append(*list, t)
			}
		case "session":
			if t.Session == value.(int) {
				*list = append(*list, t)
			}
		}
		bst.searchInOrderTraversal(t.right, field, value, list)
	}
	return *list
}
