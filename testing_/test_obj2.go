package testing_

import "github.com/mindstand/gogm"

type ExampleObject2 struct {
	gogm.BaseNode

	Children2 []*ExampleObject2 `gogm:"direction=incoming;relationship=test"`
	Parents2 *ExampleObject2 `gogm:"direction=outgoing;relationship=test"`
}
