package gateway

import "strings"

type TreeNode struct {
	Name       string
	Children   []*TreeNode
	RouterName string
	IsEnd      bool
	GwName     string
}

// put path: /user/get/:id
// 放入来判断尾节点
func (t *TreeNode) Put(path string, gwName string) {
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.Children
		isMatch := false
		for _, node := range children {
			if node.Name == name {
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
			node := &TreeNode{Name: name, Children: make([]*TreeNode, 0), IsEnd: isEnd, GwName: gwName}
			children = append(children, node)
			t.Children = children
			t = node
		}
	}
	t = root
}

// get path: /user/get/1
func (t *TreeNode) Get(path string) *TreeNode {
	strs := strings.Split(path, "/")
	routerName := ""
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.Children
		isMatch := false
		for _, node := range children {
			if node.Name == name || node.Name == "*" || strings.Contains(node.Name, ":") {
				isMatch = true
				routerName += "/" + node.Name
				t = node
				t.RouterName = routerName
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				if node.Name == "**" {
					routerName += "/" + node.Name
					t.RouterName = routerName
					return node
				}
			}
		}
	}
	return nil
}
