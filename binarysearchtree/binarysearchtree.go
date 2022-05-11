package binarysearchtree

import (
	"errors"
	"fmt"
	"time"
)

type BinaryNode struct {
	ID      int
	Dentist interface{}
	Patient interface{}
	Date    string
	Session int
	left    *BinaryNode
	right   *BinaryNode
}

type BinarySearchTree struct {
	root *BinaryNode
}

func New() *BinarySearchTree {
	bst := &BinarySearchTree{nil}
	return bst
}

func (bst *BinarySearchTree) Add(id int, date string, session int, dentist interface{}, patient interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic, recovered value: %v\n", r)
		}
	}()
	bst.insertNode(&bst.root, date, session, dentist, patient, id)
}

func (bst *BinarySearchTree) insertNode(t **BinaryNode, date string, session int, dentist interface{}, patient interface{}, id int) {
	if (*t) == nil {
		newNode := &BinaryNode{id, dentist, patient, date, session, nil, nil}
		(*t) = newNode
	} else {
		if date < (*t).Date {
			bst.insertNode(&(*t).left, date, session, dentist, patient, id) // dereferencing
		} else {
			bst.insertNode(&(*t).right, date, session, dentist, patient, id) // dereferencing
		}
	}
}

func (bst *BinarySearchTree) GetAppointmentByID(id int) *BinaryNode {
	var result *BinaryNode
	SearchAppointmentByID(bst.root, id, &result)
	return result
}

func SearchAppointmentByID(t *BinaryNode, id int, result **BinaryNode) {
	if t != nil {
		SearchAppointmentByID(t.left, id, result)
		if t.ID == id {
			*result = t
		}
		SearchAppointmentByID(t.right, id, result)
	}
}

func (bst *BinarySearchTree) findSuccessor(t *BinaryNode) *BinaryNode {
	for t.right != nil { // Find node on extreme right
		t = t.right
	}
	return t
}

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
			(*t) = bst.findSuccessor((*t).left)
			removeNode = (*t)
			(*t).left, _ = bst.removeNode(&(*t).left, removeNode)
		}
	}
	return *t, nil
}

func (bst *BinarySearchTree) Remove(removeNode *BinaryNode) error {
	bst.root, _ = bst.removeNode(&bst.root, removeNode)
	return nil
}

func (bst *BinarySearchTree) GetSize() int {
	return size(bst.root)
}

func size(node *BinaryNode) int {
	if node == nil {
		return 0
	} else {
		return (size(node.left) + 1 + size(node.right))
	}
}

func (bst *BinarySearchTree) Contains(date string, session int, dentist interface{}, patient interface{}) bool {
	newNode := BinaryNode{0, dentist, patient, date, session, nil, nil}
	return containsTraversal(bst.root, newNode)
}

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

func (bst *BinarySearchTree) GetAppointmentByDate(date, role string, searchInterface interface{}) []*BinaryNode {
	list := []*BinaryNode{}
	bst.searchAppointmentByDate(bst.root, date, role, searchInterface, &list)
	return list
}

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

func (bst *BinarySearchTree) GetAllAppointments(searchInterface interface{}, role string) []*BinaryNode {
	list := []*BinaryNode{}
	oldDate := time.Now().AddDate(-100, 0, 0)
	bst.searchAppointments(bst.root, oldDate.Format("2006-01-02"), searchInterface, role, &list)
	return list
}

func (bst *BinarySearchTree) GetUpComingAppointments(searchInterface interface{}, role string) []*BinaryNode {
	list := []*BinaryNode{}
	currentTime := time.Now()
	bst.searchAppointments(bst.root, currentTime.Format("2006-01-02"), searchInterface, role, &list)
	return list
}

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

func (bst *BinarySearchTree) SearchAllByField(field string, value interface{}, channel chan []*BinaryNode) {
	list := []*BinaryNode{}
	bst.searchInOrderTraversal(bst.root, field, value, &list)
	channel <- list
}

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
