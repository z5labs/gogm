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
	"errors"

	dsl "github.com/mindstand/go-cypherdsl"
)

// Pagination is used to control the pagination behavior of `LoadAllDepthFilterPagination``
type Pagination struct {
	// PageNumber specifies which page number to load
	PageNumber int
	// LimitPerPage limits how many records per page
	LimitPerPage int
	// OrderByVarName specifies variable to order by
	OrderByVarName string
	// OrderByField specifies field to order by on
	OrderByField string
	// OrderByDesc specifies whether orderby is desc or asc
	OrderByDesc bool
}

func (p *Pagination) Paginate(query dsl.Cypher) error {
	if p.OrderByField != "" && p.OrderByVarName != "" {
		query.OrderBy(dsl.OrderByConfig{
			Name:   p.OrderByVarName,
			Member: p.OrderByField,
			Desc:   p.OrderByDesc,
		})
	} else if p.OrderByField != "" || p.OrderByVarName != "" {
		return errors.New("ordering configuration invalid: OrderByVarName and OrderByField must be defined simultaneously, or not at all")
	}

	if p.PageNumber < 0 {
		return errors.New("pagination configuration invalid: PageNumber is below 0")
	}

	if p.LimitPerPage < 0 {
		return errors.New("pagination configuration invalid: LimitPerPage is below 0")
	}

	if p.PageNumber > 0 {
		if p.LimitPerPage < 1 {
			return errors.New("pagination configuration invalid: PageNumber is set but LimitPerPage is below 1")
		}
		query.Skip(p.LimitPerPage * p.PageNumber)
	}

	if p.LimitPerPage > 0 {
		query.Limit(p.LimitPerPage)
	}

	return nil
}
