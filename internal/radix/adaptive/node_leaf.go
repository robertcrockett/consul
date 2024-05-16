// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package adaptive

import (
	"bytes"
)

type NodeLeaf[T any] struct {
	value  T
	keyLen uint32
	key    []byte
}

func (n *NodeLeaf[T]) getPartialLen() uint32 {
	// no-op
	return 0
}

func (n *NodeLeaf[T]) setPartialLen(partialLen uint32) {
	// no-op
}

func (n *NodeLeaf[T]) getArtNodeType() nodeType {
	return leafType
}

func (n *NodeLeaf[T]) getNumChildren() uint8 {
	return 0
}

func (n *NodeLeaf[T]) setNumChildren(numChildren uint8) {
	// no-op
}

func (n *NodeLeaf[T]) isLeaf() bool {
	return true
}

func (n *NodeLeaf[T]) getValue() T {
	return n.value
}

func (n *NodeLeaf[T]) setValue(value T) {
	n.value = value
}

func (n *NodeLeaf[T]) getKeyLen() uint32 {
	return n.keyLen
}

func (n *NodeLeaf[T]) setKeyLen(keyLen uint32) {
	n.keyLen = keyLen
}

func (n *NodeLeaf[T]) getKey() []byte {
	return n.key
}

func (n *NodeLeaf[T]) setKey(key []byte) {
	n.key = key
}

func (n *NodeLeaf[T]) getPartial() []byte {
	//no-op
	return []byte{}
}

func (n *NodeLeaf[T]) setPartial(partial []byte) {
	// no-op
}

func (l *NodeLeaf[T]) prefixContainsMatch(key []byte) bool {
	if len(key) == 0 || len(l.key) == 0 {
		return false
	}
	if key == nil {
		return false
	}

	return bytes.HasPrefix(getKey(key), getKey(l.key))
}

func (n *NodeLeaf[T]) Iterator() *Iterator[T] {
	stack := make([]Node[T], 0)
	stack = append(stack, n)
	nodeT := Node[T](n)
	return &Iterator[T]{
		stack: stack,
		root:  &nodeT,
	}
}

func (n *NodeLeaf[T]) PathIterator(path []byte) *PathIterator[T] {
	nodeT := Node[T](n)
	return &PathIterator[T]{parent: &nodeT,
		path:  getTreeKey(path),
		stack: []Node[T]{nodeT},
	}
}

func (n *NodeLeaf[T]) matchPrefix(prefix []byte) bool {
	if len(n.key) == 0 {
		return false
	}
	actualKey := n.key[:len(n.key)-1]
	actualPrefix := prefix[:len(prefix)-1]
	return bytes.HasPrefix(actualKey, actualPrefix)
}

func (n *NodeLeaf[T]) getChild(index int) *Node[T] {
	return nil
}

func (n *NodeLeaf[T]) Clone() Node[T] {
	newNode := &NodeLeaf[T]{
		keyLen: n.getKeyLen(),
		key:    make([]byte, len(n.getKey())),
		value:  n.getValue(),
	}
	copy(newNode.key[:], n.key[:])
	nodeT := Node[T](newNode)
	return nodeT
}

func (n *NodeLeaf[T]) setChild(int, *Node[T]) {
	return
}
