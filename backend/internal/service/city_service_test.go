package service

import (
	"testing"
)

var testCities = []string{"Barishal", "Chattogram", "Chittagong", "Cox's Bazar", "Dhaka", "Rajshahi", "Sylhet"}

func TestCityKeyTreatsSpellingsAsOneCity(t *testing.T) {
	same := [][]string{
		{"Chattogram", "Chittagong", "chittagong", "CTG"},
		{"Barishal", "Barisal"},
		{"Cox's Bazar", "Coxs Bazar", "cox bazar", "Cox’s Bazar"},
		{"Dhaka", "dhaka", " DHAKA "},
		{"Cumilla", "Comilla"},
	}
	for _, group := range same {
		want := cityKey(group[0])
		for _, g := range group[1:] {
			if got := cityKey(g); got != want {
				t.Errorf("cityKey(%q)=%q, want %q (same as %q)", g, got, want, group[0])
			}
		}
	}
	if cityKey("Dhaka") == cityKey("Sylhet") {
		t.Error("different cities must not share a key")
	}
}

func TestResolveCityNamesFindsEveryStoredSpelling(t *testing.T) {
	got := resolveCityNames("chittagong", testCities)
	if len(got) != 2 || got[0] != "Chattogram" || got[1] != "Chittagong" {
		t.Fatalf("got %v, want both Chattogram and Chittagong", got)
	}
	if got := resolveCityNames("Atlantis", testCities); len(got) != 0 {
		t.Fatalf("unknown city should resolve to nothing, got %v", got)
	}
	if got := resolveCityNames("cox bazar", testCities); len(got) != 1 || got[0] != "Cox's Bazar" {
		t.Fatalf("got %v", got)
	}
}

func TestSuggestCities(t *testing.T) {
	cases := []struct {
		query string
		want  string // expected first suggestion
		match string
	}{
		{"dha", "Dhaka", "prefix"},
		{"Dhaka", "Dhaka", "exact"},
		{"rajshai", "Rajshahi", "similar"}, // typo
		{"rajsahi", "Rajshahi", "similar"}, // typo
		{"barisal", "Barishal", "exact"},   // other spelling counts as the same city
		{"chitt", "Chattogram", "alias"},   // half-typed other spelling
		{"coxs", "Cox's Bazar", "prefix"},
		{"syl", "Sylhet", "prefix"},
	}
	for _, tc := range cases {
		got := suggestCities(tc.query, testCities, 5)
		if len(got) == 0 {
			t.Errorf("%q: no suggestions", tc.query)
			continue
		}
		if got[0].Name != tc.want && !(tc.want == "Chattogram" && got[0].Name == "Chittagong") {
			t.Errorf("%q: first suggestion %q, want %q (all: %v)", tc.query, got[0].Name, tc.want, got)
		}
		if got[0].Match != tc.match {
			t.Errorf("%q: match %q, want %q", tc.query, got[0].Match, tc.match)
		}
	}
}

func TestSuggestCitiesDedupesSpellingsAndHandlesEdgeCases(t *testing.T) {
	got := suggestCities("ch", testCities, 5)
	count := 0
	for _, s := range got {
		if cityKey(s.Name) == "chattogram" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Chattogram and Chittagong should appear once, got %v", got)
	}

	if got := suggestCities("zzzzzz", testCities, 5); len(got) != 0 {
		t.Errorf("nonsense should give nothing, got %v", got)
	}
	if got := suggestCities("", testCities, 3); len(got) != 3 {
		t.Errorf("empty query should list the first cities, got %v", got)
	}
	if got := suggestCities("d", testCities, 1); len(got) != 1 {
		t.Errorf("limit not respected: %v", got)
	}
}

func TestLevenshtein(t *testing.T) {
	if levenshtein("rajshai", "rajshahi") != 1 || levenshtein("same", "same") != 0 || levenshtein("", "abc") != 3 {
		t.Error("levenshtein distances are wrong")
	}
}
