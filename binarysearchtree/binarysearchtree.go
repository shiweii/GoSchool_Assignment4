// Package binarysearchtree implements a binary search tree.
package binarysearchtree

import (
	"errors"
	"github.com/shiweii/logger"
)

// BinaryNode is an element within the binary search tree.
type BinaryNode struct {
	Key   string
	Data  interface{}
	Left  *BinaryNode
	Right *BinaryNode
}

// BinarySearchTree holds elements of the binary search tree.
type BinarySearchTree struct {
	root *BinaryNode
}

// New will return a newly created instance of a binary search tree.
func New() *BinarySearchTree {
	bst := &BinarySearchTree{nil}
	return bst
}

// GetRootNode will return the root node of the binary search tree.
func (bst *BinarySearchTree) GetRootNode() *BinaryNode {
	return bst.root
}

// Add wrapper function to added new element into the binary search tree.
func (bst *BinarySearchTree) Add(key string, data interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logger.Panic.Printf("panic, recovered value: %v\n", r)
		}
	}()
	bst.insertNode(&bst.root, key, data)
}

// insertNode inserts a new binary node into the binary search tree.
func (bst *BinarySearchTree) insertNode(t **BinaryNode, key string, data interface{}) {
	if (*t) == nil {
		newNode := &BinaryNode{key, data, nil, nil}
		*t = newNode
	} else {
		if key < (*t).Key {
			bst.insertNode(&(*t).Left, key, data) // de-referencing
		} else {
			bst.insertNode(&(*t).Right, key, data) // de-referencing
		}
	}
}

// Remove wrapper function to remove application from the binary search tree.
func (bst *BinarySearchTree) Remove(removeNode *BinaryNode) error {
	var err error
	bst.root, err = bst.removeNode(&bst.root, removeNode)
	return err
}

// RemoveNode removes a node from the binary search tree base on the follow cases
// Case 1, node to be deleted has 0 child (is a leaf)
// Case 2, node to be deleted has 1 child
// Case 3, node to be deleted has 2 children
func (bst *BinarySearchTree) removeNode(t **BinaryNode, removeNode *BinaryNode) (*BinaryNode, error) {
	if *t == nil {
		return nil, errors.New("error: tree is empty")
	} else if removeNode.Key < (*t).Key {
		(*t).Left, _ = bst.removeNode(&(*t).Left, removeNode)
	} else if removeNode.Key > (*t).Key {
		(*t).Right, _ = bst.removeNode(&(*t).Right, removeNode)
	} else {
		if (*t).Left == nil {
			return (*t).Right, nil
		} else if (*t).Right == nil {
			return (*t).Left, nil
		} else { // 3rd case of 2 children
			*t = bst.findSuccessor((*t).Left)
			removeNode = *t
			(*t).Left, _ = bst.removeNode(&(*t).Left, removeNode)
		}
	}
	return *t, nil
}

// findSuccessor find and return binary nodes next smaller value
func (bst *BinarySearchTree) findSuccessor(t *BinaryNode) *BinaryNode {
	for t.Right != nil { // Find node on extreme right
		t = t.Right
	}
	return t
}
