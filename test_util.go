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

type mockResult struct {
	records  [][]interface{}
	curIndex int
}

func newMockResult(records [][]interface{}) *mockResult {
	return &mockResult{
		records:  records,
		curIndex: -1,
	}
}

func (m *mockResult) Keys() ([]string, error) {
	panic("implement me")
}

func (m *mockResult) Next() bool {
	if m.records == nil || len(m.records) == 0 {
		return false
	}

	if m.curIndex+1 == len(m.records) {
		return false
	}

	m.curIndex++
	return true
}

func (m *mockResult) NextRecord(record **neo4j.Record) bool {
	panic("implement me")
}

func (m *mockResult) Err() error {
	panic("implement me")
}

func (m *mockResult) Record() *neo4j.Record {
	return &neo4j.Record{
		Values: m.records[m.curIndex],
	}
}

func (m *mockResult) Collect() ([]*neo4j.Record, error) {
	panic("implement me")
}

func (m *mockResult) Single() (*neo4j.Record, error) {
	panic("implement me")
}

func (m *mockResult) Consume() (neo4j.ResultSummary, error) {
	panic("implement me")
}
