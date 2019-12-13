// Copyright (c) 2019 MindStand Technologies, Inc
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
	"github.com/mindstand/gogm"
	"reflect"
)

type SpecialEdge struct {
	gogm.BaseNode

	Start *ExampleObject2
	End   *ExampleObject

	SomeField string `gogm:"name=some_field"`
}

func (s *SpecialEdge) GetStartNode() interface{} {
	return s.Start
}

func (s *SpecialEdge) GetStartNodeType() reflect.Type {
	return reflect.TypeOf(&ExampleObject2{})
}

func (s *SpecialEdge) SetStartNode(v interface{}) error {
	s.Start = v.(*ExampleObject2)
	return nil
}

func (s *SpecialEdge) GetEndNode() interface{} {
	return s.End
}

func (s *SpecialEdge) GetEndNodeType() reflect.Type {
	return reflect.TypeOf(&ExampleObject{})
}

func (s *SpecialEdge) SetEndNode(v interface{}) error {
	s.End = v.(*ExampleObject)
	return nil
}
