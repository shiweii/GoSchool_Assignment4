// Package doublylinkedlist implements a doubly linked list.
package doublylinkedlist

import (
	"errors"
	"fmt"
	"reflect"
)

// Node is an element of a linked list.
type Node struct {
	Value    interface{}
	Previous *Node
	Next     *Node
}

// DoublyLinkedList represents a doubly linked list.
type DoublyLinkedList struct {
	head *Node
	tail *Node
	size int
}

// New will return a newly created instance of a doubly linked list.
func New() *DoublyLinkedList {
	list := &DoublyLinkedList{nil, nil, 0}
	return list
}

// GetHeadNode returns the head node of doubly linked list.
func (list *DoublyLinkedList) GetHeadNode() *Node {
	return list.head
}

// GetTailNode returns the tail node of doubly linked list.
func (list *DoublyLinkedList) GetTailNode() *Node {
	return list.tail
}

// GetSize return Size of linked list.
func (list *DoublyLinkedList) GetSize() int {
	return list.size
}

// Add appends an interface to the end of the linked list.
func (list *DoublyLinkedList) Add(elm interface{}) error {
	newNode := &Node{
		Value:    elm,
		Previous: nil,
		Next:     nil,
	}
	if list.head == nil {
		list.head = newNode
		list.tail = newNode
	} else {
		currentNode := list.head
		for currentNode.Next != nil {
			currentNode = currentNode.Next
		}
		currentNode.Next = newNode
		newNode.Previous = currentNode
		list.tail = newNode
	}
	list.size++
	return nil
}

// Remove Wrapper function to remove element from the linked list.
func (list *DoublyLinkedList) Remove(elm interface{}) (interface{}, error) {
	var index = 1
	currentNode := list.head

	for currentNode != nil {
		if reflect.DeepEqual(currentNode.Value, elm) {
			return list.RemoveNode(index)
		}
		currentNode = currentNode.Next
		index++
	}
	return nil, nil
}

// RemoveNode removes the element at the given index from the linked list.
func (list *DoublyLinkedList) RemoveNode(index int) (interface{}, error) {
	if list.head == nil {
		return "", errors.New("empty linked list")
	}

	fmt.Println(list.size)

	if index < 1 || index > list.size {
		return "", errors.New("invalid index position")
	}

	var item interface{}

	if index == 1 {
		item = list.head.Value
		list.head = list.head.Next
	} else if index == list.size {
		item = list.tail.Value
		list.tail = list.tail.Previous
		list.tail.Next = nil
	} else {
		currentNode := list.head
		prevNode := list.head
		for i := 1; i <= index-1; i++ {
			prevNode = currentNode
			currentNode = currentNode.Next
		}
		item = currentNode.Value
		prevNode.Next = currentNode.Next
		currentNode.Next.Previous = prevNode
	}
	list.size--
	return item, nil
}

// GetList returns all elements in the linked list.
func (list *DoublyLinkedList) GetList() []interface{} {
	var values []interface{}
	currentNode := list.head
	if currentNode == nil {
		return values
	}
	values = append(values, currentNode.Value)
	for currentNode.Next != nil {
		currentNode = currentNode.Next
		values = append(values, currentNode.Value)
	}
	return values
}

// Get returns the element at index.
func (list *DoublyLinkedList) Get(index int) interface{} {
	var value interface{}
	currentNode := list.head
	//if currentNode == nil {
	//	return "", errors.New("linked list is empty")
	//}
	if index == 1 {
		value = currentNode.Value
	} else {
		for i := 1; i <= index-1; i++ {
			currentNode = currentNode.Next
		}
		value = currentNode.Value
	}
	return value
}

// Clear removes all elements from the list.
func (list *DoublyLinkedList) Clear() {
	list.head = nil
	list.tail = nil
	list.size = 0
}
