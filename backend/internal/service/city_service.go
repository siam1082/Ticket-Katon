package service

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

// cityAliasToKey maps other common spellings to one group key. All keys and aliases
// are written in normalised form (lower case, no apostrophes or punctuation).
var cityAliasToKey = map[string]string{
	"chittagong":  "chattogram",
	"chittagonj":  "chattogram",
	"chattagram":  "chattogram",
	"chittagram":  "chattogram",
	"ctg":         "chattogram",
	"barisal":     "barishal",
	"borishal":    "barishal",
	"comilla":     "cumilla",
	"bogra":       "bogura",
	"jessore":     "jashore",
	"dacca":       "dhaka",
	"mymensing":   "mymensingh",
	"narayangonj": "narayanganj",
	"cox bazar":   "coxs bazar",
	"coxsbazar":   "coxs bazar",
	"cox bazaar":  "coxs bazar",
	"coxs bazaar": "coxs bazar",
}

var cityAliasesByKey = func() map[string][]string {
	m := map[string][]string{}
	for alias, key := range cityAliasToKey {
		m[key] = append(m[key], alias)
	}
	return m
}()

var nonCityChars = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// normCity lower-cases a city and drops punctuation: "Cox's Bazar" -> "coxs bazar".
func normCity(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer("'", "", "\u2019", "").Replace(s)
	return strings.TrimSpace(nonCityChars.ReplaceAllString(s, " "))
}

// cityKey is the same for every spelling of one city.
func cityKey(s string) string {
	n := normCity(s)
	if key, ok := cityAliasToKey[n]; ok {
		return key
	}
	return n
}

func min3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min3(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}

type scoredCity struct {
	name  string
	match string
	score int
}

// matchCity scores one known city against what the user typed (lower score = better).
func matchCity(query, name string) (scoredCity, bool) {
	q := normCity(query)
	n := normCity(name)
	key := cityKey(name)

	if cityKey(query) == key {
		return scoredCity{name, "exact", 0}, true
	}

	// Names this city is known by, in normalised form.
	names := append([]string{n, key}, cityAliasesByKey[key]...)

	for _, c := range names {
		if strings.HasPrefix(c, q) {
			match := "prefix"
			if c != n && c != key {
				match = "alias"
			}
			return scoredCity{name, match, 1}, true
		}
	}
	for _, c := range names {
		if strings.Contains(c, q) {
			return scoredCity{name, "contains", 2}, true
		}
	}

	// Typo tolerance: allow 1 edit for short words, 2 for longer ones.
	if len([]rune(q)) >= 4 {
		limit := 1
		if len([]rune(q)) >= 7 {
			limit = 2
		}
		best := 99
		for _, c := range names {
			if d := levenshtein(q, c); d < best {
				best = d
			}
			// also compare against the start of the name, for half-typed misspellings
			if r := []rune(c); len(r) > len([]rune(q)) {
				if d := levenshtein(q, string(r[:len([]rune(q))])); d < best {
					best = d
				}
			}
		}
		if best <= limit {
			return scoredCity{name, "similar", 3 + best}, true
		}
	}
	return scoredCity{}, false
}

// suggestCities returns up to limit known cities for what the user typed.
// Cities written differently but meaning the same place appear once.
func suggestCities(query string, known []string, limit int) []domain.CitySuggestion {
	if limit <= 0 {
		limit = 8
	}

	if normCity(query) == "" {
		out := []domain.CitySuggestion{}
		seen := map[string]bool{}
		for _, name := range known {
			if k := cityKey(name); !seen[k] {
				seen[k] = true
				out = append(out, domain.CitySuggestion{Name: name, Match: "prefix"})
			}
			if len(out) == limit {
				break
			}
		}
		return out
	}

	var scored []scoredCity
	for _, name := range known {
		if sc, ok := matchCity(query, name); ok {
			scored = append(scored, sc)
		}
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score < scored[j].score
		}
		return scored[i].name < scored[j].name
	})

	out := []domain.CitySuggestion{}
	seen := map[string]bool{}
	for _, sc := range scored {
		k := cityKey(sc.name)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, domain.CitySuggestion{Name: sc.name, Match: sc.match})
		if len(out) == limit {
			break
		}
	}
	return out
}

// resolveCityNames returns every known city that is the same place as input.
func resolveCityNames(input string, known []string) []string {
	key := cityKey(input)
	if key == "" {
		return nil
	}
	var out []string
	for _, name := range known {
		if cityKey(name) == key {
			out = append(out, name)
		}
	}
	return out
}

// CityLookup is what the trip search needs from the city service.
type CityLookup interface {
	Resolve(ctx context.Context, input string) ([]string, error)
}

// CityService serves city suggestions and spelling-tolerant lookups. The list of
// cities is cached for a short time, so new routes show up within the minute.
type CityService struct {
	repo repository.CityRepository

	mu      sync.Mutex
	cached  []string
	expires time.Time
}

const cityCacheTTL = 30 * time.Second

func NewCityService(repo repository.CityRepository) *CityService {
	return &CityService{repo: repo}
}

func (s *CityService) known(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cached != nil && time.Now().Before(s.expires) {
		return s.cached, nil
	}
	cities, err := s.repo.ListCities(ctx)
	if err != nil {
		return nil, err
	}
	s.cached = cities
	s.expires = time.Now().Add(cityCacheTTL)
	return cities, nil
}

func (s *CityService) Suggest(ctx context.Context, query string, limit int) ([]domain.CitySuggestion, error) {
	known, err := s.known(ctx)
	if err != nil {
		return nil, err
	}
	return suggestCities(query, known, limit), nil
}

func (s *CityService) Resolve(ctx context.Context, input string) ([]string, error) {
	known, err := s.known(ctx)
	if err != nil {
		return nil, err
	}
	return resolveCityNames(input, known), nil
}
