package testing_

import (
	"github.com/mindstand/gogm"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLinking(t *testing.T) {
	req := require.New(t)

	id1 := "SDFdasasdf"
	id2 := "aasdfasdfa"

	obj1 := &ExampleObject{
		BaseNode: gogm.BaseNode{
			Id:      0,
			UUID:    id1,
			LoadMap: map[string]*gogm.RelationConfig{},
		},
	}

	obj2 := &ExampleObject{
		BaseNode: gogm.BaseNode{
			Id:      1,
			UUID:    id2,
			LoadMap: map[string]*gogm.RelationConfig{},
		},
	}

	req.Nil(obj1.LinkToExampleObjectOnFieldParents(obj2))

	req.Equal(1, len(obj2.Children))
	req.NotNil(obj1.Parents)

	req.Nil(obj1.UnlinkFromExampleObjectOnFieldParents(obj2))
	req.Equal(0, len(obj2.Children))
	req.Nil(obj1.Parents)

	// test special edge
	specEdge := &SpecialEdge{
		SomeField: "asdfad",
	}

	obj3 := &ExampleObject2{
		BaseNode:  gogm.BaseNode{
			UUID: "adfadsfasd",
		},
	}

	req.Nil(obj3.LinkToExampleObjectOnFieldSpecial(obj1, specEdge))
	req.Equal(obj1.Special.End.UUID, obj3.UUID)
	req.Equal(1, len(obj3.Special))

	req.Nil(obj1.UnlinkFromExampleObject2OnFieldSpecial(obj3))
	req.Nil(obj1.Special)
	req.Equal(0, len(obj3.Special))
}
