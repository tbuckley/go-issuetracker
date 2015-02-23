package gcode

import (
	"math"
)

type Link struct {
	Relationship string `xml:"rel,attr"`
	Type         string `xml:"type,attr"`
	URL          string `xml:"href,attr"`
}

type Feed struct {
	Links        []Link `xml:"link"`
	TotalResults int    `xml:"totalResults"`
	StartIndex   int    `xml:"startIndex"`
	ItemsPerPage int    `xml:"itemsPerPage"`
}

func (f *Feed) NumPages() int {
	return int(math.Ceil(float64(f.TotalResults) / float64(f.ItemsPerPage)))
}

type Entry struct {
	Title     string `xml:"title"`
	Content   string `xml:"content" datastore:",noindex"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
	Author    string `xml:"author>name"`
	Links     []Link `xml:"link"  datastore:",noindex"`
}

type Reply struct {
	Entry
	CCChanges    []string `xml:"updates>ccUpdate"`
	LabelChanges []string `xml:"updates>label"`
	StatusChange string   `xml:"updates>status"`
}

type RepliesFeed struct {
	Feed
	Replies []*Reply `xml:"entry"`
}

type Issue struct {
	Entry
	ID      int      `xml:"http://schemas.google.com/projecthosting/issues/2009 id"`
	Labels  []string `xml:"label"`
	Owner   string   `xml:"owner>username"`
	CCs     []string `xml:"cc>username"`
	Stars   int      `xml:"stars"`
	State   string   `xml:"state"`
	Status  string   `xml:"status"`
	Replies []*Reply
}

func (e *Issue) RepliesURL() (string, bool) {
	for _, link := range e.Links {
		if link.Relationship == "replies" && link.Type == "application/atom+xml" {
			return link.URL, true
		}
	}
	return "", false
}

type IssuesFeed struct {
	Feed
	Issues []*Issue `xml:"entry"`
}
