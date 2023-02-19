package v3

import (
	"fmt"
)

type Comparable interface {
	~int | ~uint
}

type AATree[K Comparable] struct {
	key   K
	left  *AATree[K]
	right *AATree[K]
	level int
}

func Skew[K Comparable](tree *AATree[K]) *AATree[K] {
	if tree == nil || tree.left == nil {
		return tree
	}

	if tree.left.level == tree.level { // there is a red node to the left => do a right rotation
		l := tree.left
		tree.left = l.right
		l.right = tree
		return l
	}

	return tree
}

func Split[K Comparable](tree *AATree[K]) *AATree[K] {
	if tree == nil || tree.right == nil || tree.right.right == nil {
		return tree
	}

	if tree.right.right.level == tree.level { // there is a red chain to the right-right => do a left rotation
		l := tree.right
		tree.right = l.left
		l.left = tree
		l.level++
		return l
	}

	return tree
}

func Insert[K Comparable](tree *AATree[K], key K) *AATree[K] {
	if tree == nil {
		return &AATree[K]{
			key:   key,
			left:  nil,
			right: nil,
			level: 1,
		}
	}

	if key < tree.key {
		tree.left = Insert(tree.left, key)
	} else {
		tree.right = Insert(tree.right, key)
	}

	return Split(Skew(tree))
}

func PrintTree[K Comparable](tree *AATree[K], space string) {
	if tree.left != nil {
		PrintTree(tree.left, space+"  ")
	}

	fmt.Printf("%s->%v\n", space, tree.key)

	if tree.right != nil {
		PrintTree(tree.right, space+"  ")
	}
}
