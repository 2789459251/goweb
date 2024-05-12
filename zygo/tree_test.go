package zygo

import (
	"fmt"
	"testing"
)

func TestTreeNode(t *testing.T) {
	root := &treeNode{
		name:     "/",
		children: make([]*treeNode, 0),
	}
	root.Put("/user/get/:id")
	root.Put("/user/creat/hello")
	root.Put("/user/creat/aaa")
	root.Put("/order/get/aaa")

	node := root.Get("/user/get/1")
	fmt.Println(node)
	node = root.Get("/user/creat/hello")
	fmt.Println(node)
	node = root.Get("/user/creat/aaa")
	fmt.Println(node)
	node = root.Get("/order/get/aaa")
	fmt.Println(node)
}
