package doublylinkedlist

import (
	"errors"
	"fmt"
	"reflect"
)

type node struct {
	value    interface{}
	previous *node
	next     *node
}

type DoublyLinkedlist struct {
	head *node
	tail *node
	size int
}

// New will return a newly created instance of a doubly linked list
func New() *DoublyLinkedlist {
	list := &DoublyLinkedlist{nil, nil, 0}
	return list
}

// Get Size of linked list
func (list *DoublyLinkedlist) GetSize() int {
	return list.size
}

func (list *DoublyLinkedlist) Add(elm interface{}) error {
	newNode := &node{
		value:    elm,
		previous: nil,
		next:     nil,
	}
	if list.head == nil {
		list.head = newNode
		list.tail = newNode
	} else {
		currentNode := list.head
		for currentNode.next != nil {
			currentNode = currentNode.next
		}
		currentNode.next = newNode
		newNode.previous = currentNode
		list.tail = newNode
	}
	list.size++
	return nil
}

func (list *DoublyLinkedlist) Remove(elm interface{}) (interface{}, error) {
	var index int = 1
	currentNode := list.head

	for currentNode != nil {
		if reflect.DeepEqual(currentNode.value, elm) {
			return list.RemoveNode(index)
		}
		currentNode = currentNode.next
		index++
	}
	return nil, nil
}

func (list *DoublyLinkedlist) RemoveNode(index int) (interface{}, error) {
	if list.head == nil {
		return "", errors.New("empty linked list")
	}

	fmt.Println(list.size)

	if index < 1 || index > list.size {
		return "", errors.New("invalid index position")
	}

	var item interface{}

	if index == 1 {
		item = list.head.value
		list.head = list.head.next
	} else if index == list.size {
		item = list.tail.value
		list.tail = list.tail.previous
		list.tail.next = nil
	} else {
		currentNode := list.head
		prevNode := list.head
		for i := 1; i <= index-1; i++ {
			prevNode = currentNode
			currentNode = currentNode.next
		}
		item = currentNode.value
		prevNode.next = currentNode.next
		currentNode.next.previous = prevNode
	}
	list.size--
	return item, nil
}

func (list *DoublyLinkedlist) GetList() []interface{} {
	var values []interface{}
	currentNode := list.head
	if currentNode == nil {
		return values
	}
	values = append(values, currentNode.value)
	for currentNode.next != nil {
		currentNode = currentNode.next
		values = append(values, currentNode.value)
	}
	return values
}

func (list *DoublyLinkedlist) Get(index int) interface{} {
	var value interface{}
	currentNode := list.head
	//if currentNode == nil {
	//	return "", errors.New("linked list is empty")
	//}
	if index == 1 {
		value = currentNode.value
	} else {
		for i := 1; i <= index-1; i++ {
			currentNode = currentNode.next
		}
		value = currentNode.value
	}
	return value
}

func (list *DoublyLinkedlist) Clear() {
	list.head = nil
	list.tail = nil
	list.size = 0
}

func getFieldValue(itf interface{}, field string) interface{} {
	rfl := reflect.ValueOf(itf).Elem()
	value := rfl.FieldByName(field).Interface()
	return value
}

func (list *DoublyLinkedlist) FindByUsername(username string) interface{} {
	if len(username) > 0 {
		return list.recursiveBinarySearchByUsername(list.head, list.tail, username, list.size)
	}
	return nil
}

func (list *DoublyLinkedlist) recursiveBinarySearchByUsername(firstNode *node, lastNode *node, value string, size int) interface{} {
	if firstNode == nil || lastNode == nil {
		return nil
	}
	firstNodeVal := getFieldValue(firstNode.value, "Username").(string)
	lastNodeVal := getFieldValue(lastNode.value, "Username").(string)
	if firstNodeVal > lastNodeVal {
		return nil
	} else {
		mid := size / 2
		midNode := middleNode(firstNode, mid)
		midNodeVal := getFieldValue(midNode.value, "Username").(string)
		if midNodeVal == value {
			return midNode.value
		} else {
			if value < midNodeVal {
				return list.recursiveBinarySearchByUsername(firstNode, midNode.previous, value, mid)
			} else {
				return list.recursiveBinarySearchByUsername(midNode.next, lastNode, value, mid)
			}
		}
	}
}

func middleNode(start *node, mid int) *node {
	if start == nil {
		return nil
	}
	for i := 1; i < mid; i++ {
		start = start.next
	}
	return start
}

func (list *DoublyLinkedlist) PrintAllNodes() error {
	currentNode := list.head
	if currentNode == nil {
		fmt.Println("Linked list is empty.")
		return nil
	}
	fmt.Printf("%+v\n", currentNode.value)
	for currentNode.next != nil {
		currentNode = currentNode.next
		fmt.Printf("%+v\n", currentNode.value)
	}
	return nil
}

// Swap value of given node
func (list *DoublyLinkedlist) swapData(first, second *node) {
	value := first.value
	first.value = second.value
	second.value = value
}

// Sort elements using insertion sort
func (list *DoublyLinkedlist) InsertionSort() {
	// Get first node
	var front *node = list.head
	var back *node = nil
	for front != nil {
		// Get next node
		back = front.next
		// Update node value when consecutive nodes are not sort
		for back != nil && back.previous != nil {
			backVal := getFieldValue(back.value, "Username").(string)
			prevVal := getFieldValue(back.previous.value, "Username").(string)
			if backVal < prevVal {
				// Modified node data
				list.swapData(back, back.previous)
			}
			// Visit to previous node
			back = back.previous
		}
		// Visit to next node
		front = front.next
	}
}

func (list *DoublyLinkedlist) SearchByMobileNumber(mobileNum int) interface{} {
	//var ret interface{}
	ret := list.recursiveSeqSearchByMobileNumber(list.head, mobileNum)
	return ret
}

func (list *DoublyLinkedlist) recursiveSeqSearchByMobileNumber(node *node, value int) interface{} {
	if node == nil {
		return nil
	} else {
		phoneNum := getFieldValue(node.value, "MobileNumber").(int)
		if phoneNum == value {
			return node.value
		}
		return list.recursiveSeqSearchByMobileNumber(node.next, value)
	}
}
