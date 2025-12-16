package main

import "fmt"

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func NewNode(val int) *TreeNode {
	return &TreeNode{Val: val}
}

func BuildTree(values []interface{}) *TreeNode {
	if len(values) == 0 || values[0] == nil {
		return nil
	}

	root := &TreeNode{Val: values[0].(int)}
	queue := []*TreeNode{root}
	i := 1

	for len(queue) > 0 && i < len(values) {
		node := queue[0]
		queue = queue[1:]

		// Left Child
		if i < len(values) && values[i] != nil {
			node.Left = &TreeNode{Val: values[i].(int)}
			queue = append(queue, node.Left)
		}
		i++

		// Right Child
		if i < len(values) && values[i] != nil {
			node.Right = &TreeNode{Val: values[i].(int)}
			queue = append(queue, node.Right)
		}
		i++
	}
	return root
}

func main() {
	// Test Tree : [1, 2, 3, 4, 5]
	tree := BuildTree([]interface{}{1, 2, 3, 4, 5})
	fmt.Println("Tree Created:", tree.Val)
	fmt.Println("Root.Left:", tree.Left.Val)
	fmt.Println("Root.Right:", tree.Right.Val)
	fmt.Println("Root.Left.Left:", tree.Left.Left.Val)
	fmt.Println("Root.Left.Right:", tree.Left.Right.Val)
}
