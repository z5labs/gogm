package gogm

import (
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestParseStruct(t *testing.T){
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	parseO2O(req)

	parseM2O(req)

	parseM2M(req)
}

func parseO2O(req *require.Assertions) {
	//test single save
	comp1 := &a{
		TestField: "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt: 600,
		Id:        1,
	}

	b1 := &b{
		TestField: "test",
		Id: 2,
	}

	c1 := &c{
		Start: comp1,
		End: b1,
		Test: "testing",
	}

	comp1.SingleSpecA = c1
	b1.SingleSpec = c1

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}

	val := reflect.ValueOf(comp1)

	err := parseStruct("", "", false, 0, nil, &val, 0, 5, &nodes, &relations)
	req.Nil(err)
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
}

func parseM2O(req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField: "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt: 600,
		Id:        1,
		ManyA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		Id: 2,
	}

	b1.ManyB = a1
	a1.ManyA = append(a1.ManyA, b1)

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}

	val := reflect.ValueOf(a1)

	err := parseStruct("", "", false, 0, nil, &val, 0, 5, &nodes, &relations)
	req.Nil(err)
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
}

func parseM2M(req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField: "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt: 600,
		Id:        1,
		ManyA: []*b{},
		MultiA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		Id: 2,
		Multi: []*a{},
	}

	b1.Multi = append(b1.Multi, a1)
	a1.MultiA = append(a1.MultiA, b1)

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}

	val := reflect.ValueOf(a1)

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

	conn, err := driverPool.Open(driver.ReadWriteMode)
	if err != nil {
		require.Nil(t, err)
	}
	driverPool.Reclaim(conn)

	req.Nil(saveDepth(conn, comp2, defaultSaveDepth))
}