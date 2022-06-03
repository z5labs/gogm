// Copyright (c) 2022 MindStand Technologies, Inc
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
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestPrimaryKeyStrategy_validate(t *testing.T) {
	cases := []struct {
		name       string
		strategy   PrimaryKeyStrategy
		shouldPass bool
	}{
		{
			name:       "Zero Test",
			strategy:   PrimaryKeyStrategy{},
			shouldPass: false,
		},
		{
			name: "Valid",
			strategy: PrimaryKeyStrategy{
				StrategyName: "ValidIndex",
				DBName:       "id",
				Type:         reflect.TypeOf(""),
				GenIDFunc: func() interface{} {
					return ""
				},
			},
			shouldPass: true,
		},
		{
			name: "Invalid",
			strategy: PrimaryKeyStrategy{
				StrategyName: "invalid",
				DBName:       "id",
				Type:         reflect.TypeOf(""),
				GenIDFunc: func() interface{} {
					return 1
				},
			},
			shouldPass: false,
		},
	}

	req := require.New(t)
	for _, c := range cases {
		t.Log("testing case [" + c.name + "]")
		err := c.strategy.validate()
		if c.shouldPass {
			req.Nil(err)
		} else {
			req.NotNil(err)
		}
	}
}
