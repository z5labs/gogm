package gogm

import "github.com/neo4j/neo4j-go-driver/v4/neo4j"

type testResult struct {
	empty bool
	num   int
}

func (t *testResult) Keys() ([]string, error) {
	panic("implement me")
}

func (t *testResult) Next() bool {
	toRet := !t.empty && t.num == 0

	if !t.empty {
		t.num++
	}

	return toRet
}

func (t *testResult) NextRecord(record **neo4j.Record) bool {
	panic("implement me")
}

func (t *testResult) Err() error {
	panic("implement me")
}

func (t *testResult) Record() *neo4j.Record {
	return &neo4j.Record{}
}

func (t *testResult) Collect() ([]*neo4j.Record, error) {
	panic("implement me")
}

func (t *testResult) Single() (*neo4j.Record, error) {
	panic("implement me")
}

func (t *testResult) Consume() (neo4j.ResultSummary, error) {
	panic("implement me")
}

//
//func (t *testResult) Next() bool {
//	panic("implement me")
//}
//
//func (t *testResult) Record() j.Record {
//	panic("implement me")
//}
//
//func (t *testResult) Summary() (j.ResultSummary, error) {
//	panic("implement me")
//}
//
//func (t *testResult) Consume() (j.ResultSummary, error) {
//	panic("implement me")
//}
//
//func (t *testResult) Collect() ([]*neo4j.Record, error) {
//	panic("implement me")
//}
//
//func (t *testResult) Single() (*neo4j.Record, error) {
//	panic("implement me")
//}
//
//func (t *testResult) Keys() ([]string, error) {
//	panic("implement me")
//}
//
//func (t *testResult) NextRecord(record **neo4j.Record) bool {
//	toRet := !t.empty && t.num == 0
//
//	if !t.empty {
//		t.num++
//	}
//
//	return toRet
//}
//
//func (t *testResult) Err() error {
//	panic("implement me")
//}
//
//func (t *testResult) Record() neo4j.Record {
//	return neo4j.Record{}
//}
//
//func (t *testResult) Summary() (neo4j.ResultSummary, error) {
//	panic("implement me")
//}
//
//func (t *testResult) Consume() (neo4j.ResultSummary, error) {
//	panic("implement me")
//}
