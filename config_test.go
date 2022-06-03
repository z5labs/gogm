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
	"testing"
)

func TestConfig_ConnectionString(t *testing.T) {
	cases := []struct {
		Name     string
		Config   *Config
		Expected string
	}{
		{
			Name: "Protocol Defined",
			Config: &Config{
				Host:     "localhost",
				Port:     7687,
				Protocol: "neo4j",
			},
			Expected: "neo4j://localhost:7687",
		},
		{
			Name: "IsCluster False",
			Config: &Config{
				Host:      "localhost",
				Port:      7687,
				IsCluster: false,
			},
			Expected: "bolt://localhost:7687",
		},
		{
			Name: "IsCluster True",
			Config: &Config{
				Host:      "localhost",
				Port:      7687,
				IsCluster: true,
			},
			Expected: "neo4j://localhost:7687",
		},
		{
			Name: "IsCluster and Protocol defined",
			Config: &Config{
				Host:      "localhost",
				Port:      7687,
				IsCluster: false,
				Protocol:  "neo4j",
			},
			Expected: "neo4j://localhost:7687",
		},
	}
	req := require.New(t)
	for _, _case := range cases {
		t.Run(_case.Name, func(t *testing.T) {
			req.Equal(_case.Expected, _case.Config.ConnectionString(), "Connection strings should be equal")
		})
	}
}
