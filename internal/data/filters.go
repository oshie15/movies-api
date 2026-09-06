package data

import (
	"slices"
	"strings"

	"greenlight.oshie.net/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// Check that the client-provided Sort field matches one of the entries in our safelist
// and if it does, extract the column name from the Sort field by stripping the leading
// hyphen character (if one exists).
func (f Filters) sortColumn() string {
	if slices.Contains(f.SortSafelist, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}

	panic("unsafe sort parameter: " + f.Sort)
}

// Return the sort direction (ASC or DESC) depending on the prefix character of the Sort field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 millions")
	v.Check(f.PageSize > 0, "page_size", "must be a greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check the sort parameter matches a value in the safelist
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// Define a new Metadata struct for holding the pagination metadata.
type Metadata struct {
	CurrentPage  int `json:"current_page, omitzero"`
	PageSize     int `json:"page_size, omitzero"`
	FirstPage    int `json:"first_page, omitzero"`
	LastPage     int `json:"last_page, omitzero"`
	TotalRecords int `json:"total_records, omitzero"`
}

// The calculateMetadata() function calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values. Note
// that when the last page values is calculated we are dividing two int values, and
// when dviding integer types in Go the result will also be an integer type, with
// the modules dropped. So, for example, if there were 12 records in total and a page
// size of 5, the last page value would be (12+5-1)/5 = 3.2, which is then truncated to 3 by Go
func calculateMetadata(TotalRecords, page, pageSize int) Metadata {
	if TotalRecords == 0 {
		// Note that we return an empty Metaadata struct if there are no records/
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (TotalRecords + pageSize - 1) / pageSize,
		TotalRecords: TotalRecords,
	}
}
