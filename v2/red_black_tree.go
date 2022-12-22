package market

import (
	"fmt"
)

type dye bool

const (
	black, red dye = true, false
)

// TreeNode is a single leaf of the tree
type TreeNode struct {
	Queue  *OrderQueue
	Left   *TreeNode
	Right  *TreeNode
	parent *TreeNode
	Price  Decimal
	color  dye
}

func (n *TreeNode) Color() dye {
	if n == nil {
		return black
	}
	return n.color
}

func (n *TreeNode) grandparent() *TreeNode {
	if n != nil && n.parent != nil {
		return n.parent.parent
	}

	return nil
}

func (n *TreeNode) uncle() *TreeNode {
	if n == nil || n.parent == nil || n.parent.parent == nil {
		return nil
	}

	return n.parent.sibling()
}

func (n *TreeNode) sibling() *TreeNode {
	if n == nil || n.parent == nil {
		return nil
	}

	if n == n.parent.Left {
		return n.parent.Right
	}

	return n.parent.Left
}

func (n *TreeNode) maxNode() *TreeNode {
	if n == nil {
		return nil
	}

	for n.Right != nil {
		n = n.Right
	}

	return n
}

// Size returns the number of elements stored in the subtree.
// Computed dynamically on each call, meaning the subtree is traversed to count the number of the nodes.
func (n *TreeNode) Size() int {
	if n == nil {
		return 0
	}
	size := 1

	if n.Left != nil {
		size += n.Left.Size()
	}

	if n.Right != nil {
		size += n.Right.Size()
	}

	return size
}

func (n TreeNode) String() string {
	return fmt.Sprintf("%v", n.Price)
}

// RedBlackTree holds elements of the red-black tree
type RedBlackTree struct {
	Root *TreeNode
	size int
}

// Put inserts node into the tree.
func (t *RedBlackTree) Put(price Decimal, queue *OrderQueue) {
	if t.Root == nil {
		t.Root = &TreeNode{Price: price, Queue: queue, color: red}
		insert(t, t.Root)
		t.size++
		return
	}

	node := t.Root
	isLooping := true
	var insertedNode *TreeNode

	for isLooping {
		priceCmp := price.Cmp(node.Price)
		switch {
		case priceCmp == 0:
			node.Price = price
			node.Queue = queue
			return

		case priceCmp < 0:
			if node.Left == nil {
				node.Left = &TreeNode{Price: price, Queue: queue, color: red}
				insertedNode = node.Left
				isLooping = false
			} else {
				node = node.Left
			}

		case priceCmp > 0:
			if node.Right == nil {
				node.Right = &TreeNode{Price: price, Queue: queue, color: red}
				insertedNode = node.Right
				isLooping = false
			} else {
				node = node.Right
			}
		}
	}

	insertedNode.parent = node
	insert(t, insertedNode)
	t.size++
}

// Remove the node from the tree.
func (t *RedBlackTree) Remove(price Decimal) {
	node := t.lookup(price)
	if node == nil {
		return
	}

	if node.Left != nil && node.Right != nil {
		prev := node.Left.maxNode()
		node.Price = prev.Price
		node.Queue = prev.Queue
		node = prev
	}

	var child *TreeNode
	if node.Left == nil || node.Right == nil {
		if node.Right == nil {
			child = node.Left
		} else {
			child = node.Right
		}

		if node.color == black {
			node.color = child.Color()
			deleteNode(t, node)
		}

		t.replaceNode(node, child)

		if node.parent == nil && child != nil {
			child.color = black
		}
	}

	t.size--
}

// Empty returns true if tree does not contain any nodes
func (t *RedBlackTree) Empty() bool {
	return t.size == 0
}

func (t *RedBlackTree) lookup(price Decimal) *TreeNode {
	node := t.Root
	for node != nil {
		priceCmp := price.Cmp(node.Price)
		switch {
		case priceCmp == 0:
			return node
		case priceCmp < 0:
			node = node.Left
		case priceCmp > 0:
			node = node.Right
		}
	}
	return nil
}

func (t *RedBlackTree) rotateLeft(node *TreeNode) {
	right := node.Right
	t.replaceNode(node, right)
	node.Right = right.Left

	if right.Left != nil {
		right.Left.parent = node
	}

	right.Left = node
	node.parent = right
}

func (t *RedBlackTree) rotateRight(node *TreeNode) {
	left := node.Left
	t.replaceNode(node, left)
	node.Left = left.Right

	if left.Right != nil {
		left.Right.parent = node
	}

	left.Right = node
	node.parent = left
}

func (t *RedBlackTree) replaceNode(old, new *TreeNode) {
	if old.parent == nil {
		t.Root = new
	} else {
		if old == old.parent.Left {
			old.parent.Left = new
		} else {
			old.parent.Right = new
		}
	}

	if new != nil {
		new.parent = old.parent
	}
}

func insert(t *RedBlackTree, node *TreeNode) {
	if node.parent == nil {
		node.color = black
		return
	}

	insert2(t, node)
}

func insert2(t *RedBlackTree, node *TreeNode) {
	if node.parent.Color() == black {
		return
	}

	insert3(t, node)
}

func insert3(t *RedBlackTree, node *TreeNode) {
	uncle := node.uncle()
	if uncle.Color() == red {
		node.parent.color = black
		uncle.color = black
		node.grandparent().color = red
		insert(t, node.grandparent())
		return
	}

	insert4(t, node)
}

func insert4(t *RedBlackTree, node *TreeNode) {
	grandparent := node.grandparent()
	if node == node.parent.Right && node.parent == grandparent.Left {
		t.rotateLeft(node.parent)
		node = node.Left
		insert5(t, node)
		return
	}

	if node == node.parent.Left && node.parent == grandparent.Right {
		t.rotateRight(node.parent)
		node = node.Right
		insert5(t, node)
		return
	}

	insert5(t, node)
}

func insert5(t *RedBlackTree, node *TreeNode) {
	node.parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red
	if node == node.parent.Left && node.parent == grandparent.Left {
		t.rotateRight(grandparent)
		return
	}

	if node == node.parent.Right && node.parent == grandparent.Right {
		t.rotateLeft(grandparent)
	}
}

func deleteNode(t *RedBlackTree, node *TreeNode) {
	if node.parent == nil {
		return
	}

	delete2(t, node)
}

func delete2(t *RedBlackTree, node *TreeNode) {
	sibling := node.sibling()
	if sibling.Color() == red {
		node.parent.color = red
		sibling.color = black
		if node == node.parent.Left {
			t.rotateLeft(node.parent)
		} else {
			t.rotateRight(node.parent)
		}
	}

	delete3(t, node)
}

func delete3(t *RedBlackTree, node *TreeNode) {
	sibling := node.sibling()
	if node.parent.Color() == black &&
		sibling.Color() == black &&
		sibling.Left.Color() == black &&
		sibling.Right.Color() == black {
		sibling.color = red
		deleteNode(t, node.parent)
		return
	}

	delete4(t, node)
}

func delete4(t *RedBlackTree, node *TreeNode) {
	sibling := node.sibling()
	if node.parent.Color() == red &&
		sibling.Color() == black &&
		sibling.Left.Color() == black &&
		sibling.Right.Color() == black {
		sibling.color = red
		node.parent.color = black
		return
	}

	delete5(t, node)
}

func delete5(t *RedBlackTree, node *TreeNode) {
	sibling := node.sibling()
	if node == node.parent.Left &&
		sibling.Color() == black &&
		sibling.Left.Color() == red &&
		sibling.Right.Color() == black {
		sibling.color = red
		sibling.Left.color = black
		t.rotateRight(sibling)
		delete6(t, node)
		return
	}

	if node == node.parent.Right &&
		sibling.Color() == black &&
		sibling.Right.Color() == red &&
		sibling.Left.Color() == black {
		sibling.color = red
		sibling.Right.color = black
		t.rotateLeft(sibling)
		delete6(t, node)
		return
	}

	delete6(t, node)
}

func delete6(t *RedBlackTree, node *TreeNode) {
	sibling := node.sibling()
	sibling.color = node.parent.Color()
	node.parent.color = black
	if node == node.parent.Left && sibling.Right.Color() == red {
		sibling.Right.color = black
		t.rotateLeft(node.parent)
		return
	}

	if sibling.Left.Color() == red {
		sibling.Left.color = black
		t.rotateRight(node.parent)
	}
}

// Min gets the min value and flag if found
func (t *RedBlackTree) Min() (*OrderQueue, bool) {
	node, found := t.minFromNode(t.Root)
	if node != nil {
		return node.Queue, found
	}

	return nil, false
}

// Max gets the max value and flag if found
func (t *RedBlackTree) Max() (*OrderQueue, bool) {
	node, found := t.maxFromNode(t.Root)
	if node != nil {
		return node.Queue, found
	}

	return nil, false
}

func (t *RedBlackTree) minFromNode(node *TreeNode) (*TreeNode, bool) {
	if node == nil {
		return nil, false
	}

	if node.Left == nil {
		return node, true
	}

	return t.minFromNode(node.Left)
}

func (t *RedBlackTree) maxFromNode(node *TreeNode) (*TreeNode, bool) {
	if node == nil {
		return nil, false
	}

	if node.Right == nil {
		return node, true
	}

	return t.maxFromNode(node.Right)
}
