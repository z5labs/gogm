// Copyright (c) 2022 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package testing_

import (
	"github.com/mindstand/gogm/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func int64ptr(n int64) *int64 {
	return &n
}

func TestLinking(t *testing.T) {
	req := require.New(t)

	id1 := "SDFdasasdf"
	id2 := "aasdfasdfa"

	obj1 := &ExampleObject{
		BaseUUIDNode: gogm.BaseUUIDNode{
			UUID: id1,
			BaseNode: gogm.BaseNode{
				Id:      int64ptr(0),
				LoadMap: map[string]*gogm.RelationConfig{},
			},
		},
	}

	obj2 := &ExampleObject{
		BaseUUIDNode: gogm.BaseUUIDNode{
			UUID: id2,
			BaseNode: gogm.BaseNode{
				Id:      int64ptr(1),
				LoadMap: map[string]*gogm.RelationConfig{},
			},
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
		BaseUUIDNode: gogm.BaseUUIDNode{
			UUID: "adfadsfasd",
		},
	}

	req.Nil(obj3.LinkToExampleObjectOnFieldSpecial(obj1, specEdge))
	req.Equal(obj1.Special.Start.UUID, obj3.UUID)
	req.Equal(1, len(obj3.Special))

	req.Nil(obj3.UnlinkFromExampleObjectOnFieldSpecial(obj1))
	req.Nil(obj1.Special)
	req.Equal(0, len(obj3.Special))
}
