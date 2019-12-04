package gogm

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

	sess, err := NewSession(false)
	req.Nil(err)
	defer sess.Close()

	log.Println("testIndexManagement")
	testIndexManagement(req)

	log.Println("test save")
	testSave(sess, req)

	req.Nil(sess.PurgeDatabase())
}

// runs with integration test
func testSave(sess *Session, req *require.Assertions) {
	req.Nil(sess.Begin())
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
}
