package gogm

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestParseStruct(t *testing.T){
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	comp2 := &a{
		TestField: "test",
		Id:        1,
	}

	b2 := &b{
		TestField: "test",
		Id: 2,
	}

	c1 := &c{
		Start: comp2,
		End: b2,
		Test: "testing",
	}

	comp2.SingleSpec = c1
	b2.SingleSpec = c1

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}

	val := reflect.ValueOf(comp2)

	err := parseStruct("", "", false, 0, nil, &val, 0, 5, &nodes, &relations)
	req.Nil(err)
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
}
