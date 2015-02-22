package main

import (
	"sort"
	"strconv"
	"time"

	"github.com/tbuckley/go-issuetracker/query"
)

type IntPropertyFunc func(entry *query.Entry) (int, bool)
type StringPropertyFunc func(entry *query.Entry) (string, bool)
type StringListPropertyFunc func(entry *query.Entry) []string
type TimePropertyFunc func(entry *query.Entry) (time.Time, bool)

type IssuePair interface {
	HasKeyLessThan(p IssuePair) bool
	Issues() []*query.Entry
	KeyString() string
}

type KeySortedPairList []IssuePair

func (l KeySortedPairList) Len() int {
	return len(l)
}
func (l KeySortedPairList) Less(i, j int) bool {
	return l[i].HasKeyLessThan(l[j])
}
func (l KeySortedPairList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type NumIssuesSortedPairList []IssuePair

func (l NumIssuesSortedPairList) Len() int {
	return len(l)
}
func (l NumIssuesSortedPairList) Less(i, j int) bool {
	return len(l[i].Issues()) < len(l[j].Issues())
}
func (l NumIssuesSortedPairList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// INT

type IntPair struct {
	Key     *int
	Entries []*query.Entry
}

func (p *IntPair) HasKeyLessThan(pair IssuePair) bool {
	switch v := pair.(type) {
	case *IntPair:
		switch {
		case v.Key == nil:
			return false
		case p.Key == nil:
			return true
		default:
			return *p.Key < *v.Key
		}
	default:
		return false
	}
}

func (p *IntPair) Issues() []*query.Entry {
	return p.Entries
}

func (p *IntPair) KeyString() string {
	if p.Key == nil {
		return "None"
	}
	return strconv.Itoa(*p.Key)
}

type IntGroups struct {
	Groups map[int][]*query.Entry
	None   []*query.Entry
}

func (g *IntGroups) Pairs() []IssuePair {
	pairs := make([]IssuePair, 0, len(g.Groups)+1)
	for k, v := range g.Groups {
		i := k
		pairs = append(pairs, &IntPair{&i, v})
	}
	if len(g.None) > 0 {
		pairs = append(pairs, &IntPair{nil, g.None})
	}
	return pairs
}

func (g *IntGroups) PairsByValue() []IssuePair {
	pairs := g.Pairs()
	sort.Sort(KeySortedPairList(pairs))
	return pairs
}

func (g *IntGroups) PairsByNumEntries() []IssuePair {
	pairs := g.Pairs()
	sort.Sort(NumIssuesSortedPairList(pairs))
	return pairs
}

func GroupIntProperty(entries []*query.Entry, propFunc IntPropertyFunc) *IntGroups {
	groups := &IntGroups{
		Groups: make(map[int][]*query.Entry),
	}
	for _, entry := range entries {
		val, ok := propFunc(entry)
		if !ok {
			groups.None = append(groups.None, entry)
		} else {
			groups.Groups[val] = append(groups.Groups[val], entry)
		}
	}
	return groups
}

// STRING

type StringPair struct {
	Key     *string
	Entries []*query.Entry
}

func (p *StringPair) HasKeyLessThan(pair IssuePair) bool {
	switch v := pair.(type) {
	case *StringPair:
		switch {
		case v.Key == nil:
			return false
		case p.Key == nil:
			return true
		default:
			return *p.Key < *v.Key
		}
	default:
		return false
	}
}

func (p *StringPair) Issues() []*query.Entry {
	return p.Entries
}

func (p *StringPair) KeyString() string {
	if p.Key == nil {
		return "None"
	}
	return *p.Key
}

type StringGroups struct {
	Groups map[string][]*query.Entry
	None   []*query.Entry
}

func (g *StringGroups) Pairs() []IssuePair {
	pairs := make([]IssuePair, 0, len(g.Groups)+1)
	for k, v := range g.Groups {
		i := k
		pairs = append(pairs, &StringPair{&i, v})
	}
	if len(g.None) > 0 {
		pairs = append(pairs, &StringPair{nil, g.None})
	}
	return pairs
}

func (g *StringGroups) PairsByValue() []IssuePair {
	pairs := g.Pairs()
	sort.Sort(KeySortedPairList(pairs))
	return pairs
}

func (g *StringGroups) PairsByNumEntries() []IssuePair {
	pairs := g.Pairs()
	sort.Sort(NumIssuesSortedPairList(pairs))
	return pairs
}

func GroupStringProperty(entries []*query.Entry, propFunc StringPropertyFunc) *StringGroups {
	groups := &StringGroups{
		Groups: make(map[string][]*query.Entry),
	}
	for _, entry := range entries {
		val, ok := propFunc(entry)
		if !ok {
			groups.None = append(groups.None, entry)
		} else {
			groups.Groups[val] = append(groups.Groups[val], entry)
		}
	}
	return groups
}

// TIME

type TimePair struct {
	Key     *time.Time
	Entries []*query.Entry
}

func (p *TimePair) HasKeyLessThan(pair IssuePair) bool {
	switch v := pair.(type) {
	case *TimePair:
		switch {
		case v.Key == nil:
			return false
		case p.Key == nil:
			return true
		default:
			return p.Key.Before(*v.Key)
		}
	default:
		return false
	}
}

func (p *TimePair) Issues() []*query.Entry {
	return p.Entries
}

func (p *TimePair) KeyString() string {
	if p.Key == nil {
		return "None"
	}
	return p.Key.Format("2006-01-02")
}

type TimeGroups struct {
	Groups map[time.Time][]*query.Entry
	None   []*query.Entry
}

func (g *TimeGroups) Pairs() []IssuePair {
	pairs := make([]IssuePair, 0, len(g.Groups)+1)
	for k, v := range g.Groups {
		i := k
		pairs = append(pairs, &TimePair{&i, v})
	}
	if len(g.None) > 0 {
		pairs = append(pairs, &TimePair{nil, g.None})
	}
	return pairs
}

func (g *TimeGroups) PairsByValue() []IssuePair {
	pairs := g.Pairs()
	sort.Sort(KeySortedPairList(pairs))
	return pairs
}

func (g *TimeGroups) PairsByNumEntries() []IssuePair {
	pairs := g.Pairs()
	sort.Sort(NumIssuesSortedPairList(pairs))
	return pairs
}

func GroupTimeProperty(entries []*query.Entry, propFunc TimePropertyFunc) *TimeGroups {
	groups := &TimeGroups{
		Groups: make(map[time.Time][]*query.Entry),
	}
	for _, entry := range entries {
		val, ok := propFunc(entry)
		if !ok {
			groups.None = append(groups.None, entry)
		} else {
			groups.Groups[val] = append(groups.Groups[val], entry)
		}
	}
	return groups
}
