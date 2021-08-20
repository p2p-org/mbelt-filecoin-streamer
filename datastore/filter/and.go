package filter

import (
	"fmt"
	"strings"
)

type And struct {
	filters Filters
}

func (f *And) Build() (string, []interface{}, error) {
	filters := make([]string, 0, len(f.filters))
	args := make([]interface{}, 0, len(f.filters))

	for _, filter := range f.filters {
		if filter == nil {
			continue
		}

		sqlFilter, argsFilter, err := filter.Build()
		if err != nil {
			return "", nil, err
		}

		filters = append(filters, sqlFilter)
		args = append(args, argsFilter...)
	}

	filterOr := fmt.Sprintf(
		"(%s)",
		strings.Join(filters, " AND "),
	)

	return filterOr, args, nil
}

func (f *And) And(filters ...Filter) {
	f.filters.Add(filters...)
}

func NewAnd(f ...Filter) *And {
	filters := make(Filters, 0, len(f))
	filters.Add(f...)

	return &And{
		filters: filters,
	}
}

