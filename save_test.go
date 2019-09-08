package gogm

import (
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
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

	comp2.SingleSpecA = c1
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

func TestSave(t *testing.T){
	t.Skip()
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	comp2 := &a{
		TestField: "test",
		Id:        1,
	}

	b2 := &b{
		TestField: "test",
		TestTime: time.Now().UTC(),
		Id: 2,
	}

	c1 := &c{
		Start: comp2,
		End: b2,
		Test: "testing",
	}

	comp2.SingleSpecA = c1
	b2.SingleSpec = c1

	err := dsl.Init(&dsl.ConnectionConfig{
		Username: "neo4j",
		Password: "password",
		Host: "0.0.0.0",
		Port: 7687,
		PoolSize: 15,
	})
	require.Nil(t, err)

	sess := dsl.NewSession()

	req.Nil(saveDepth(sess, comp2, defaultSaveDepth))
}