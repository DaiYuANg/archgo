package render

import "testing"

type nestedFilter struct {
	IDs []int `json:"ids"`
}

type queryParams struct {
	Name   string       `db:"name"`
	Status string       `json:"status"`
	Filter nestedFilter `json:"filter"`
}

func TestLookupPrefersFieldThenTag(t *testing.T) {
	params := queryParams{
		Name:   "alice",
		Status: "ACTIVE",
		Filter: nestedFilter{IDs: []int{1, 2, 3}},
	}

	if got := lookup(params, "Name"); got.IsAbsent() || got.MustGet() != "alice" {
		t.Fatalf("expected field lookup to resolve Name")
	}
	if got := lookup(params, "name"); got.IsAbsent() || got.MustGet() != "alice" {
		t.Fatalf("expected tag lookup to resolve db tag name")
	}
	if got := lookup(params, "status"); got.IsAbsent() || got.MustGet() != "ACTIVE" {
		t.Fatalf("expected tag lookup to resolve json tag status")
	}
	if got := lookup(params, "filter.ids"); got.IsAbsent() {
		t.Fatalf("expected nested tag lookup to resolve filter.ids")
	}
}
