package testing_

import "github.com/mindstand/gogm"

type TreeNode struct {
	gogm.BaseNode

	Parents []*TreeNode `gogm:"direction=incoming;relationship=tree"`
	Children []*TreeNode `gogm:"direction=outgoing;relationship=tree"`

	Roots []*RootTreeNode `gogm:"direction=outgoing;relationship=root"`

	Sides []*SideTreeNode `gogm:"direction=outgoing;relationship=sides"`
}

type RootTreeNode struct {
	gogm.BaseNode

	Trees []*TreeNode `gogm:"direction=incoming;relationship=root"`
	Sides []*SideTreeNode `gogm:"direction=outgoing;relationship=sides"`
}

type SideTreeNode struct {
	gogm.BaseNode

	Trees []*TreeNode `gogm:"direction=incoming;relationship=sides"`
	Roots []*RootTreeNode `gogm:"direction=incoming;relationship=sides"`
}
