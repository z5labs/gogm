// Copyright (c) 2020 MindStand Technologies, Inc
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

package gogm

import (
	uuid2 "github.com/google/uuid"
	"sync"

	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// This test is to make sure retuning raw results from neo4j actually work. This
// proves that the bug causing empty interfaces to be returned has been fixed.
func TestRawQuery(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	req := require.New(t)

	conf := Config{
		Username:      "neo4j",
		Password:      "password",
		Host:          "0.0.0.0",
		IsCluster:     false,
		Port:          7687,
		PoolSize:      15,
		IndexStrategy: IGNORE_INDEX,
	}

	req.Nil(Init(&conf, &a{}, &b{}, &c{}))

	sess, err := NewSession(false)
	req.Nil(err)

	uuid := uuid2.New().String()

	req.Nil(sess.Save(&a{
		BaseNode: BaseNode{
			UUID: uuid,
		},
	}))

	raw, _, err := sess.QueryRaw("match (n) where n.uuid=$uuid return n", map[string]interface{}{
		"uuid": uuid,
	})
	req.Nil(err)
	req.NotEmpty(raw)
}

type tdArr []string
type tdArrOfTd []tdString
type tdMap map[string]interface{}
type tdMapTdSlice map[string]tdArr
type tdMapTdSliceOfTd map[string]tdArrOfTd

type propTest struct {
	BaseNode

	MapInterface   map[string]interface{} `gogm:"name=prop1;properties"`
	MapPrim        map[string]string      `gogm:"name=prop2;properties"`
	MapTdPrim      map[string]tdString    `gogm:"name=prop3;properties"`
	MapSlicePrim   map[string][]string    `gogm:"name=prop4;properties"`
	MapSliceTdPrim map[string][]tdString  `gogm:"name=prop5;properties"`
	SlicePrim      []string               `gogm:"name=prop6;properties"`
	SliceTdPrim    []tdString             `gogm:"name=prop7;properties"`

	TypeDefArr       tdArr            `gogm:"name=prop8;properties"`
	TypeDefArrOfTD   tdArrOfTd        `gogm:"name=prop9;properties"`
	TdMap            tdMap            `gogm:"name=prop10;properties"`
	TdMapOfTdSlice   tdMapTdSlice     `gogm:"name=prop11;properties"`
	TdMapTdSliceOfTd tdMapTdSliceOfTd `gogm:"name=prop12;properties"`
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	req := require.New(t)

	conf := Config{
		Username:      "neo4j",
		Password:      "changeme",
		Host:          "0.0.0.0",
		IsCluster:     false,
		Port:          7687,
		PoolSize:      15,
		IndexStrategy: IGNORE_INDEX,
	}

	req.Nil(Init(&conf, &a{}, &b{}, &c{}, &propTest{}))

	log.Println("opening session")

	log.Println("testIndexManagement")
	testIndexManagement(req)

	sess, err := NewSession(false)
	req.Nil(err)

	log.Println("test save")
	testSave(sess, req)

	req.Nil(sess.PurgeDatabase())

	// Test Opening and Closing Session using SessionConfig
	sessConf, err := NewSessionWithConfig(SessionConfig{
		AccessMode: AccessModeRead,
	})
	req.Nil(err)
	req.Nil(sessConf.Close())

	//testLoad(req, 500, 5)
	//req.Nil(sess.PurgeDatabase())

	req.Nil(sess.Close())

	req.Nil(driver.Close())

}

func testLoad(req *require.Assertions, numThreads, msgPerThread int) {
	var wg sync.WaitGroup
	wg.Add(numThreads)
	for i := 0; i < numThreads; i++ {
		go func(w *sync.WaitGroup, n int) {
			defer wg.Done()
			sess, err := NewSession(false)
			req.Nil(err)
			req.NotNil(sess)
			defer sess.Close()
			for j := 0; j < n; j++ {
				req.Nil(sess.Save(&a{}))
			}
		}(&wg, msgPerThread)
	}
	wg.Wait()
}

// runs with integration test
func testSave(sess *Session, req *require.Assertions) {
	req.Nil(sess.Begin())
	a2 := &a{
		TestField: "test",
		PropTest0: map[string]interface{}{
			"test.test": "test",
			"test2":     1,
		},
		PropTest1: map[string]string{
			"test": "test",
		},
		PropsTest2: []string{"test", "test"},
		PropsTest3: []int{1, 2},
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

	req.Nil(sess.SaveDepth(a2, 5))

	req.Nil(sess.Commit())
	req.Nil(sess.Begin())

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

	req.Nil(sess.SaveDepth(a2, 5))
	req.Nil(sess.Commit())
	req.Nil(a2.SingleSpecA)
	req.Nil(b2.SingleSpec)

	//test save something that isn't connected to anything
	singleSave := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		ManyA:             []*b{},
		MultiA:            []*b{},
		Created:           time.Now().UTC(),
	}

	req.Nil(sess.Begin())
	req.Nil(sess.SaveDepth(singleSave, 1))
	req.Nil(sess.Commit())

	// property test
	prop1 := propTest{
		BaseNode: BaseNode{},
		MapInterface: map[string]interface{}{
			"test": int64(1),
		},
		MapPrim: map[string]string{
			"test": "test1",
		},
		MapTdPrim: map[string]tdString{
			"test": "test2",
		},
		MapSlicePrim: map[string][]string{
			"test": {"test1", "test2"},
		},
		MapSliceTdPrim: map[string][]tdString{
			"test": {"test1", "test2"},
		},
		SlicePrim:      []string{"test2"},
		SliceTdPrim:    []tdString{"test3"},
		TypeDefArr:     []string{"test1"},
		TypeDefArrOfTD: []tdString{"test1"},
		TdMap: map[string]interface{}{
			"test": "test",
		},
		TdMapOfTdSlice: map[string]tdArr{
			"test": []string{"test1", "test2"},
		},
		TdMapTdSliceOfTd: map[string]tdArrOfTd{
			"test": []tdString{"test1", "test2"},
		},
	}

	req.Nil(sess.SaveDepth(&prop1, 0))

	var prop2 propTest
	req.Nil(sess.Load(&prop2, prop1.UUID))

	req.EqualValues(prop1.MapInterface, prop2.MapInterface)
	req.EqualValues(prop1.MapPrim, prop2.MapPrim)
	req.EqualValues(prop1.MapTdPrim, prop2.MapTdPrim)
	req.EqualValues(prop1.MapSlicePrim, prop2.MapSlicePrim)
	req.EqualValues(prop1.MapSliceTdPrim, prop2.MapSliceTdPrim)
	req.EqualValues(prop1.SlicePrim, prop2.SlicePrim)
	req.EqualValues(prop1.SliceTdPrim, prop2.SliceTdPrim)
	req.EqualValues(prop1.TypeDefArr, prop2.TypeDefArr)
	req.EqualValues(prop1.TypeDefArrOfTD, prop2.TypeDefArrOfTD)
	req.EqualValues(prop1.TdMap, prop2.TdMap)
	req.EqualValues(prop1.TdMapOfTdSlice, prop2.TdMapOfTdSlice)
	req.EqualValues(prop1.TdMapTdSliceOfTd, prop2.TdMapTdSliceOfTd)
}
