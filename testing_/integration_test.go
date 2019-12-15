package testing_

import (
	"github.com/mindstand/gogm"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIntegration(t *testing.T) {
	req := require.New(t)

	conf := gogm.Config{
		Username:      "neo4j",
		Password:      "password",
		Host:          "0.0.0.0",
		Port:          7687,
		PoolSize:      15,
		IndexStrategy: gogm.IGNORE_INDEX,
	}

	req.Nil(gogm.Init(&conf, &TreeNode{}, &RootTreeNode{}, &SideTreeNode{}))

	sess, err := gogm.NewSession(false)
	req.Nil(err)
	defer sess.Close()

	sides := make([]*SideTreeNode, 2, 2)
	sides[0] = &SideTreeNode{}
	sides[1] = &SideTreeNode{}

	treeNodes := make([]*TreeNode, 4, 4)

	for i := 0; i < 4; i++ {
		treeNodes[i] = &TreeNode{}
		req.Nil(treeNodes[i].LinkToSideTreeNodeOnFieldSides(sides[0]))
	}

	for i := 0; i < 3; i++ {
		req.Nil(treeNodes[i].LinkToTreeNodeOnFieldParents(treeNodes[3]))
	}

	root := &RootTreeNode{}
	req.Nil(root.LinkToSideTreeNodeOnFieldSides(sides[1]))
	req.Nil(root.LinkToTreeNodeOnFieldTrees(treeNodes[3]))

	req.Nil(sess.SaveDepth(root, 5))
}