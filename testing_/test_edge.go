package testing_

import (
	"github.com/mindstand/gogm"
	"reflect"
)

type SpecialEdge struct {
	gogm.BaseNode

	Start *ExampleObject
	End   *ExampleObject2

	SomeField string `gogm:"name=some_field"`
}

func (s *SpecialEdge) GetStartNode() interface{} {
	return s.Start
}

func (s *SpecialEdge) GetStartNodeType() reflect.Type {
	return reflect.TypeOf(&ExampleObject{})
}

func (s *SpecialEdge) SetStartNode(v interface{}) error {
	s.Start = v.(*ExampleObject)
	return nil
}

func (s *SpecialEdge) GetEndNode() interface{} {
	return s.End
}

func (s *SpecialEdge) GetEndNodeType() reflect.Type {
	return reflect.TypeOf(&ExampleObject2{})
}

func (s *SpecialEdge) SetEndNode(v interface{}) error {
	s.End = v.(*ExampleObject2)
	return nil
}
