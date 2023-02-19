package v3

import (
	"testing"
)

func TestExample(t *testing.T) {
	var numbers = []int{81, 99, 10, 32, 8, 19, 3, 78}

	var tree *AATree[int]
	for _, number := range numbers {
		tree = Insert(tree, number)
	}

	PrintTree(tree, "")
}
