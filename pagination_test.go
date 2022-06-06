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
	"fmt"
	"testing"

	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
)

func TestPagination_Paginate(t *testing.T) {
	req := require.New(t)

	// the query here doesn't really matter
	rawQuery := "match (n) return n"

	tests := []struct {
		Name        string
		Pagination  Pagination
		ShouldPass  bool
		ShouldEqual string
	}{
		{
			Name:        "no pagination or ordering",
			Pagination:  Pagination{},
			ShouldPass:  true,
			ShouldEqual: rawQuery,
		},
		{
			Name:        "only limit",
			Pagination:  Pagination{LimitPerPage: 5},
			ShouldPass:  true,
			ShouldEqual: rawQuery + " LIMIT 5",
		},
		{
			Name: "valid pagination",
			Pagination: Pagination{
				LimitPerPage: 5,
				PageNumber:   3,
			},
			ShouldPass:  true,
			ShouldEqual: fmt.Sprintf(rawQuery+" SKIP %v LIMIT %v", 5*3, 5),
		},
		{
			Name: "valid order by",
			Pagination: Pagination{
				OrderByVarName: "n",
				OrderByField:   "field",
			},
			ShouldPass:  true,
			ShouldEqual: rawQuery + " ORDER BY n.field",
		},
		{
			Name:       "invalid order by",
			Pagination: Pagination{OrderByField: "n"},
			ShouldPass: false,
		},
		{
			Name:       "invalid page number",
			Pagination: Pagination{PageNumber: -1},
			ShouldPass: false,
		},
		{
			Name:       "invalid limit",
			Pagination: Pagination{LimitPerPage: -1},
			ShouldPass: false,
		},
		{
			Name: "invalid pagination",
			Pagination: Pagination{
				PageNumber:   1,
				LimitPerPage: 0,
			},
			ShouldPass: false,
		},
	}

	for _, test := range tests {
		query := dsl.QB().Cypher(rawQuery)

		t.Log("running test -", test.Name)
		err := test.Pagination.Paginate(query)

		if test.ShouldPass {
			req.Nil(err, "pagination should not fail")
		} else {
			req.NotNil(err, "pagination should fail")
			continue
		}

		cypher, err := query.ToCypher()
		req.Nil(err, "valid cypher should be generated")

		if test.ShouldEqual != "" {
			req.Equal(test.ShouldEqual, cypher, "generated cypher is not what was expected")
		}
	}
}
