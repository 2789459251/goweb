package zygo

import "strings"

type treeNode struct {
	name       string
	children   []*treeNode
	routerName string
	isEnd      bool
}

// put path: /user/get/:id
// 放入来判断尾节点
func (t *treeNode) Put(path string) {
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.name == name {
				isMatch = true
				t = node
				break
			}

		}
		if !isMatch {
			isEnd := false
			if index == len(strs)-1 {
				isEnd = true
			}
			node := &treeNode{name: name, children: make([]*treeNode, 0), isEnd: isEnd}
			children = append(children, node)
			t.children = children
			t = node
		}
	}
	t = root
}

// get path: /user/get/1
func (t *treeNode) Get(path string) *treeNode {
	strs := strings.Split(path, "/")
	routerName := ""
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.name == name || node.name == "*" || strings.Contains(node.name, ":") {
				isMatch = true
				routerName += "/" + node.name
				t = node
				t.routerName = routerName
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				if node.name == "**" {
					routerName += "/" + node.name
					t.routerName = routerName
					return node
				}
			}
		}
	}
	return nil
}
