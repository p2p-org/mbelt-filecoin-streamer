package filter

import "fmt"

type Filter interface {
	Build() (sqlFilter string, args []interface{}, err error)
}

type Filters []Filter

func (f *Filters) Add(filter ...Filter) {
	if filter != nil && len(filter) > 0 {
		*f = append(*f, filter...)
	}
}

// Render filters and include in query
// Expected query like: SELECT 1 FROM table WHERE %s
func RenderQuery(query string, filters ...Filter) (string, []interface{}, error) {
	var queryWhere []interface{}
	var filtersArgs = make([]interface{}, 0)

	if filters == nil {
		queryWhere = append(queryWhere, "1=1")
	} else {
		for _, filter := range filters {
			if filter == nil {
				queryWhere = append(queryWhere, "1=1")
				continue
			}
			filterQueryWhere, filterArgs, err := filter.Build()
			if err != nil {
				return "", nil, err
			}
			queryWhere = append(queryWhere, filterQueryWhere)
			filtersArgs = append(filtersArgs, filterArgs...)
		}
	}
	return fmt.Sprintf(query, queryWhere...), filtersArgs, nil
}
