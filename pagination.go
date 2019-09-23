package gogm

import "errors"

type Pagination struct {
	PageNumber     int
	LimitPerPage   int
	OrderByVarName string
	OrderByField   string
	OrderByDesc    bool
}

func (p *Pagination) Validate() error {
	if p.PageNumber >= 0 && p.LimitPerPage > 1 && p.OrderByField != "" && p.OrderByVarName != "" {
		return errors.New("pagination configuration invalid, please double check")
	}

	return nil
}
