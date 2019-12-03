package gogm

import (
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestParseStruct(t *testing.T) {
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	parseO2O(req)

	parseM2O(req)

	parseM2M(req)
}

func parseO2O(req *require.Assertions) {
	//test single save
	comp1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseNode: BaseNode{
			Id:   1,
			UUID: "comp1uuid",
			LoadMap: map[string]*RelationConfig{
				"SingleSpecA": {
					Ids:          []int64{2},
					RelationType: Single,
				},
			},
		},
	}

	b1 := &b{
		BaseNode: BaseNode{
			Id:   2,
			UUID: "b1uuid",
			LoadMap: map[string]*RelationConfig{
				"SingleSpec": {
					Ids:          []int64{1},
					RelationType: Single,
				},
			},
		},
		TestField: "test",
	}

	c1 := &c{
		Start: comp1,
		End:   b1,
		Test:  "testing",
	}

	comp1.SingleSpecA = c1
	b1.SingleSpec = c1

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	ids := []*string{}

	val := reflect.ValueOf(comp1)
	nodeRef := map[string]*reflect.Value{}

	err := parseStruct("", "", false, 0, nil, &val, 0, 5, &nodes, &relations, &oldRels, &ids, &nodeRef)
	req.Nil(err)
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))

	req.Equal(2, len(oldRels))
	req.Equal(2, len(curRels))
	req.Equal(int64(2), curRels["comp1uuid"]["SingleSpecA"].Ids[0])
	req.Equal(int64(1), curRels["b1uuid"]["SingleSpec"].Ids[0])
	req.EqualValues(oldRels, curRels)
}

func parseM2O(req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseNode: BaseNode{
			Id:   1,
			UUID: "a1uuid",
			LoadMap: map[string]*RelationConfig{
				"ManyA": {
					Ids:          []int64{2},
					RelationType: Multi,
				},
			},
		},
		ManyA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		BaseNode: BaseNode{
			Id:   2,
			UUID: "b1uuid",
			LoadMap: map[string]*RelationConfig{
				"ManyB": {
					Ids:          []int64{1},
					RelationType: Single,
				},
			},
		},
	}

	b1.ManyB = a1
	a1.ManyA = append(a1.ManyA, b1)

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	ids := []*string{}

	val := reflect.ValueOf(a1)
	nodeRef := map[string]*reflect.Value{}
	err := parseStruct("", "", false, 0, nil, &val, 0, 5, &nodes, &relations, &oldRels, &ids, &nodeRef)
	req.Nil(err)
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	req.EqualValues(oldRels, curRels)
}

func parseM2M(req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseNode: BaseNode{
			Id:   1,
			UUID: "a1uuid",
			LoadMap: map[string]*RelationConfig{
				"MultiA": {
					Ids:          []int64{2},
					RelationType: Multi,
				},
			},
		},
		ManyA:  []*b{},
		MultiA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		BaseNode: BaseNode{
			Id:   2,
			UUID: "b1uuid",
			LoadMap: map[string]*RelationConfig{
				"Multi": {
					Ids:          []int64{1},
					RelationType: Multi,
				},
			},
		},
		Multi: []*a{},
	}

	b1.Multi = append(b1.Multi, a1)
	a1.MultiA = append(a1.MultiA, b1)

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	ids := []*string{}

	nodeRef := map[string]*reflect.Value{}

	val := reflect.ValueOf(a1)

	err := parseStruct("", "", false, 0, nil, &val, 0, 5, &nodes, &relations, &oldRels, &ids, &nodeRef)
	req.Nil(err)
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	req.EqualValues(oldRels, curRels)
}

func TestCalculateDels(t *testing.T) {
	req := require.New(t)

	//test node removed
	dels := calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{
		"node1": {2},
	}, dels)

	//test field removed
	dels = calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
		"node2": {
			"RelFieldNew": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{
		"node1": {2},
		"node2": {1},
	}, dels)

	//test field empty
	dels = calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{
		"node1": {2},
		"node2": {1},
	}, dels)

	//test nothing changed
	dels = calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{}, dels)
}

func TestSave(t *testing.T) {
	t.Skip()
	req := require.New(t)

	conf := Config{
		Username:      "neo4j",
		Password:      "password",
		Host:          "0.0.0.0",
		Port:          7687,
		PoolSize:      15,
		IndexStrategy: IGNORE_INDEX,
	}

	req.Nil(Init(&conf, &a{}, &b{}, &c{}))

	a2 := &a{
		TestField: "test",
	}

	b2 := &b{
		TestField: "test",
		TestTime:  time.Now().UTC(),
	}

	b3 := &b{
		TestField: "dasdfasd",
	}

	c1 := &c{
		Start: a2,
		End:   b2,
		Test:  "testing",
	}

	a2.SingleSpecA = c1
	a2.ManyA = []*b{b3}
	b2.SingleSpec = c1
	b3.ManyB = a2

	conn, err := driverPool.Open(driver.ReadWriteMode)
	if err != nil {
		require.Nil(t, err)
	}
	defer driverPool.Reclaim(conn)

	req.Nil(saveDepth(conn, a2, 5))
	req.EqualValues(map[string]*RelationConfig{
		"SingleSpecA": {
			Ids:          []int64{b2.Id},
			RelationType: Single,
		},
		"ManyA": {
			Ids:          []int64{b3.Id},
			RelationType: Multi,
		},
	}, a2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"SingleSpec": {
			Ids:          []int64{a2.Id},
			RelationType: Single,
		},
	}, b2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"ManyB": {
			Ids:          []int64{a2.Id},
			RelationType: Single,
		},
	}, b3.LoadMap)
	a2.SingleSpecA = nil
	b2.SingleSpec = nil

	req.Nil(saveDepth(conn, a2, 5))
	log.Println("done")
}
