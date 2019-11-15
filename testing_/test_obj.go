package testing_

import "github.com/mindstand/gogm"

type ExampleObject struct {
	gogm.BaseNode

	Children []*ExampleObject `gogm:"direction=incoming;relationship=test"`
	Parents *ExampleObject `gogm:"direction=outgoing;relationship=test"`
}
