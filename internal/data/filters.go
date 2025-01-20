package data

import (
	"strings"

	"github.com/mf751/greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, filters Filters) {
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.Page < 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filters.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(filters.Sort, filters.SortSafeList...), "sort", "invalid sort value")
}

func (filters Filters) sortColumn() string {
	for _, safeValue := range filters.SortSafeList {
		if filters.Sort == safeValue {
			return strings.TrimPrefix(filters.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + filters.Sort)
}

func (filters Filters) sortDirection() string {
	if strings.HasPrefix(filters.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
