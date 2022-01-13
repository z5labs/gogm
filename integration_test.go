// Copyright (c) 2021 MindStand Technologies, Inc
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
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	uuid2 "github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	assert2 "github.com/stretchr/testify/assert"

	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type indexTestStruct struct {
	BaseUUIDNode
	StringIndex  string `gogm:"name=string_index;index"`
	StringUnique string `gogm:"name=string_unique;unique"`
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}
	suite.Run(t, &IntegrationTestSuite{})
}

type IntegrationTestSuite struct {
	suite.Suite
	gogm *Gogm
	conf *Config
}

func (integrationTest *IntegrationTestSuite) TearDownSuite() {
	sess, err := integrationTest.gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
	integrationTest.Require().Nil(err)
	integrationTest.Require().NotNil(sess)
	sess.QueryRaw(context.Background(), "match (n) detach delete n", nil)
	integrationTest.Require().Nil(sess.Close())
	integrationTest.Require().Nil(integrationTest.gogm.Close())
}

func (integrationTest *IntegrationTestSuite) SetupSuite() {
	conf := Config{
		Username:  "neo4j",
		Password:  "changeme",
		Host:      "0.0.0.0",
		IsCluster: false,
		Port:      7687,
		PoolSize:  15,
		// this is ignore because index management is part of the test
		IndexStrategy:             IGNORE_INDEX,
		EnableDriverLogs:          true,
		DefaultTransactionTimeout: 2 * time.Minute,
	}

	gogm, err := New(&conf, UUIDPrimaryKeyStrategy, &a{}, &b{}, &c{}, &propTest{})
	integrationTest.Require().Nil(err)
	integrationTest.Require().NotNil(gogm)
	integrationTest.gogm = gogm
}

func (integrationTest *IntegrationTestSuite) TestV4Index() {
	if integrationTest.gogm.neoVersion < 4 {
		integrationTest.T().Log("skipping because of incompatible version", integrationTest.gogm.neoVersion)
		integrationTest.T().Skip()
		return
	}

	assertCopy := *integrationTest.conf
	assertCopy.IndexStrategy = ASSERT_INDEX
	_, err := New(&assertCopy, UUIDPrimaryKeyStrategy, &indexTestStruct{})
	integrationTest.Assert().Nil(err)

	validateCopy := *integrationTest.conf
	validateCopy.IndexStrategy = VALIDATE_INDEX
	_, err = New(&validateCopy, UUIDPrimaryKeyStrategy, &indexTestStruct{})
	integrationTest.Assert().Nil(err)
}

func (integrationTest *IntegrationTestSuite) TestSecureConnection() {
	if integrationTest.gogm.neoVersion < 4 {
		integrationTest.T().Log("skipping secure test for v3")
		return
	}

	conf := Config{
		Username:       "neo4j",
		Password:       "changeme",
		Host:           "0.0.0.0",
		Protocol:       "neo4j+ssc",
		CAFileLocation: os.Getenv("ROOT") + "/ca-public.crt",
		Port:           7687,
		PoolSize:       15,
		// this is ignore because index management is part of the test
		IndexStrategy:             IGNORE_INDEX,
		EnableDriverLogs:          true,
		DefaultTransactionTimeout: 2 * time.Minute,
	}
	integrationTest.conf = &conf
	gogm, err := New(&conf, UUIDPrimaryKeyStrategy, &a{}, &b{}, &c{}, &propTest{})
	integrationTest.Require().Nil(err)
	integrationTest.Require().NotNil(gogm)
	defer gogm.Close()

	sess, err := gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeRead})
	integrationTest.Require().Nil(err)
	integrationTest.Require().NotNil(sess)
	defer sess.Close()

	_, _, err = sess.QueryRaw(context.Background(), "return 1;", nil)
	integrationTest.Require().Nil(err)
}

func (integrationTest *IntegrationTestSuite) TestManagedTx() {
	//req := integrationTest.Require()
	if integrationTest.gogm.neoVersion < 4 {
		integrationTest.T().Log("skipping because of incompatible version", integrationTest.gogm.neoVersion)
		integrationTest.T().Skip()
		return
	}
	assert := integrationTest.Assert()

	va := a{}
	va.UUID = uuid2.New().String()
	vb := b{}
	vb.UUID = uuid2.New().String()

	sess, err := integrationTest.gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
	integrationTest.Require().Nil(err)
	integrationTest.Require().NotNil(sess)
	integrationTest.Require().Nil(sess.SaveDepth(context.Background(), &va, 0))
	integrationTest.Require().Nil(sess.SaveDepth(context.Background(), &vb, 0))
	integrationTest.Require().Nil(sess.Close())

	var wg sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(assert *assert2.Assertions, wg *sync.WaitGroup, va *a, vb *b, t int) {
			defer wg.Done()
			sess, err := integrationTest.gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
			if !assert.NotNil(sess) || !assert.Nil(err) {
				fmt.Println("exiting routine")
				return
			}

			defer sess.Close()
			for j := 0; j < 30; j++ {
				//fmt.Printf("pass %v on thread %v\n", j, t)
				ctx := context.Background()
				err = sess.ManagedTransaction(ctx, func(tx TransactionV2) error {
					va.TestField = time.Now().UTC().String()
					err = tx.SaveDepth(ctx, va, 0)
					if err != nil {
						return err
					}

					vb.TestField = time.Now().UTC().String()
					return tx.SaveDepth(ctx, vb, 0)
				})
				if !assert.Nil(err) {
					fmt.Printf("error: %s, exiting thread", err.Error())
					return
				}
			}
		}(assert, &wg, &va, &vb, i)
	}
	wg.Wait()
}

// This test is to make sure retuning raw results from neo4j actually work. This
// proves that the bug causing empty interfaces to be returned has been fixed.
func (integrationTest *IntegrationTestSuite) TestRawQuery() {
	req := integrationTest.Require()
	sess, err := integrationTest.gogm.NewSession(SessionConfig{AccessMode: AccessModeWrite})
	req.Nil(err)
	defer sess.Close()

	uuid := uuid2.New().String()

	req.Nil(sess.Save(&a{
		BaseUUIDNode: BaseUUIDNode{
			UUID: uuid,
		},
	}))

	raw, err := sess.QueryRaw("match (n) where n.uuid=$uuid return n", map[string]interface{}{
		"uuid": uuid,
	})
	req.Nil(err)
	req.NotEmpty(raw)
}

func (integrationTest *IntegrationTestSuite) TestRawQueryV2() {
	req := integrationTest.Require()
	sess, err := integrationTest.gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
	req.Nil(err)
	defer sess.Close()

	uuid := uuid2.New().String()

	req.Nil(sess.Save(context.Background(), &a{
		BaseUUIDNode: BaseUUIDNode{
			UUID: uuid,
		},
	}))

	raw, sum, err := sess.QueryRaw(context.Background(), "match (n) where n.uuid=$uuid return n", map[string]interface{}{
		"uuid": uuid,
	})
	req.Nil(err)
	req.NotZero(sum)
	req.NotEmpty(raw)
}

type tdArr []string
type tdArrOfTd []tdString
type tdMap map[string]interface{}
type tdMapTdSlice map[string]tdArr
type tdMapTdSliceOfTd map[string]tdArrOfTd

type propTest struct {
	BaseUUIDNode

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

func (integrationTest *IntegrationTestSuite) TestIntegration() {
	log.Println("opening session")

	log.Println("testIndexManagement")
	testIndexManagement(integrationTest.Require())

	sess, err := integrationTest.gogm.NewSession(SessionConfig{AccessMode: AccessModeWrite})
	integrationTest.Require().Nil(err)

	log.Println("test save")
	testSave(sess, integrationTest.Require())

	// Test Opening and Closing Session using SessionConfig
	sessConf, err := integrationTest.gogm.NewSession(SessionConfig{
		AccessMode: AccessModeRead,
	})
	integrationTest.Require().Nil(err)
	integrationTest.Require().Nil(sessConf.Close())

	testLoad(integrationTest.Require(), integrationTest.gogm, 500, 5)

	integrationTest.Require().Nil(sess.Close())
}

func (integrationTest *IntegrationTestSuite) TestIntegrationV2() {
	req := integrationTest.Require()
	log.Println("testIndexManagement")
	testIndexManagement(req)

	sess, err := integrationTest.gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
	req.Nil(err)

	log.Println("test save")
	testSaveV2(sess, req)

	_, _, err = sess.QueryRaw(context.Background(), "match (n) detach delete n", nil)
	req.Nil(err)

	// Test Opening and Closing Session using SessionConfig
	sessConf, err := integrationTest.gogm.NewSession(SessionConfig{
		AccessMode: AccessModeRead,
	})
	req.Nil(err)
	req.Nil(sessConf.Close())

	req.Nil(sess.Close())
}

func testLoad(req *require.Assertions, gogm *Gogm, numThreads, msgPerThread int) {
	var wg sync.WaitGroup
	wg.Add(numThreads)
	for i := 0; i < numThreads; i++ {
		go func(w *sync.WaitGroup, n int) {
			defer wg.Done()
			sess, err := gogm.NewSession(SessionConfig{AccessMode: AccessModeRead})
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
func testSave(sess ISession, req *require.Assertions) {
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
			Ids:          []int64{*b2.Id},
			RelationType: Single,
		},
		"ManyA": {
			Ids:          []int64{*b3.Id},
			RelationType: Multi,
		},
	}, a2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"SingleSpec": {
			Ids:          []int64{*a2.Id},
			RelationType: Single,
		},
	}, b2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"ManyB": {
			Ids:          []int64{*a2.Id},
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
		BaseUUIDNode: BaseUUIDNode{},
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

func testSaveV2(sess SessionV2, req *require.Assertions) {
	logger := GetDefaultLogger()
	ctx := context.Background()
	req.Nil(sess.Begin(ctx))
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

	edgeC1 := &c{
		Start: a2,
		End:   b2,
		Test:  "testing",
	}

	a2.SingleSpecA = edgeC1
	a2.ManyA = []*b{b3}
	b2.SingleSpec = edgeC1
	b3.ManyB = a2

	req.Nil(sess.SaveDepth(ctx, a2, 5))

	req.Nil(sess.Commit(ctx))
	req.Nil(sess.Begin(ctx))

	req.EqualValues(map[string]*RelationConfig{
		"SingleSpecA": {
			Ids:          []int64{*b2.Id},
			RelationType: Single,
		},
		"ManyA": {
			Ids:          []int64{*b3.Id},
			RelationType: Multi,
		},
	}, a2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"SingleSpec": {
			Ids:          []int64{*a2.Id},
			RelationType: Single,
		},
	}, b2.LoadMap)
	req.EqualValues(map[string]*RelationConfig{
		"ManyB": {
			Ids:          []int64{*a2.Id},
			RelationType: Single,
		},
	}, b3.LoadMap)
	a2.SingleSpecA = nil
	b2.SingleSpec = nil

	req.Nil(sess.SaveDepth(ctx, a2, 5))
	req.Nil(sess.Commit(ctx))
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

	req.Nil(sess.SaveDepth(ctx, singleSave, 1))

	// property test
	prop1 := propTest{
		BaseUUIDNode: BaseUUIDNode{},
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

	req.Nil(sess.SaveDepth(ctx, &prop1, 0))

	var prop2 propTest
	logger.Debug("----------------------------------------------------------------------------------")
	req.Nil(sess.LoadDepth(ctx, &prop2, prop1.UUID, 0))

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

const testUuid = "f64953a5-8b40-4a87-a26b-6427e661570c"

func (i *IntegrationTestSuite) TestSchemaLoadStrategy() {
	req := i.Require()

	i.gogm.config.LoadStrategy = SCHEMA_LOAD_STRATEGY

	// create required nodes
	testSchemaLoadStrategy_Setup(i.gogm, req)

	sess, err := i.gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeRead})
	req.Nil(err)
	defer req.Nil(sess.Close())

	ctx := context.Background()
	req.Nil(sess.Begin(ctx))
	defer req.Nil(sess.Close())

	// test raw query (verify SchemaLoadStrategy + Neo driver decoding)
	query, err := SchemaLoadStrategyOne(i.gogm, "n", "a", "uuid", "uuid", false, 1, nil)
	req.Nil(err, "error generating SchemaLoadStrategy query")

	cypher, err := query.ToCypher()
	req.Nil(err, "error decoding cypher from generated SchemaLoadStrategy query")
	raw, _, err := sess.QueryRaw(ctx, cypher, map[string]interface{}{"uuid": "f64953a5-8b40-4a87-a26b-6427e661570c"})
	req.Nil(err)

	req.Len(raw, 1, "Raw result should have one record")
	req.Len(raw[0], 2, "Raw record should have two items")

	// inspecting first node
	node, ok := raw[0][0].(neo4j.Node)
	req.True(ok)
	req.ElementsMatch(node.Labels, []string{"a"})

	// inspecting nested query result
	req.Len(raw[0][1], 5)

	var res a
	err = sess.LoadDepth(ctx, &res, testUuid, 2)
	req.Nil(err, "Load should not fail")

	req.Len(res.ManyA, 1, "B node should be loaded properly")
	req.True(res.SingleSpecA.Test == "testing", "C spec rel should be loaded properly")
	req.True(res.SingleSpecA.End.TestField == "dasdfasd", "B node should be loaded through spec rel")
}

func testSchemaLoadStrategy_Setup(gogm *Gogm, req *require.Assertions) {
	sess, err := gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
	req.Nil(err)
	defer req.Nil(sess.Close())

	a1 := &a{
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

	b1 := &b{
		TestField: "dasdfasd",
	}

	c1 := &c{
		Start: a1,
		End:   b1,
		Test:  "testing",
	}

	a1.SingleSpecA = c1
	a1.ManyA = []*b{b1}
	b1.SingleSpec = c1
	b1.ManyB = a1

	a1.UUID = testUuid

	ctx := context.Background()
	req.Nil(sess.Begin(ctx))

	req.Nil(sess.SaveDepth(ctx, a1, 3))
	req.Nil(sess.Commit(ctx))
}
