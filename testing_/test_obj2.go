package testing_

import "github.com/mindstand/gogm"

type ExampleObject2 struct {
	gogm.BaseNode

	Children2 []*ExampleObject2 `gogm:"direction=incoming;relationship=test" json:"children_2"`
	Parents2 *ExampleObject2 `gogm:"direction=outgoing;relationship=test" json:"parents_2"`
}
