package query

import (
	"encoding/xml"
	"math"
)

type Link struct {
	Relationship string `xml:"rel,attr"`
	Type         string `xml:"type,attr"`
	Link         string `xml:"href,attr"`
}

type Entry struct {
	ID        int      `xml:"http://schemas.google.com/projecthosting/issues/2009 id"`
	Title     string   `xml:"title"`
	Published string   `xml:"published"`
	Updated   string   `xml:"updated"`
	Labels    []string `xml:"label"`
	Owner     string   `xml:"owner>username"`
	Author    string   `xml:"author>name"`
	CCs       []string `xml:"cc>username"`
	Stars     int      `xml:"stars"`
	State     string   `xml:"state"`
	Status    string   `xml:"status"`
	Links     []Link   `xml:"link"`
}

func (e *Entry) FullDataURL() string {
	for _, link := range e.Links {
		if link.Relationship == "" && link.Type == "" {

		}
	}
	return ""
}

type Feed struct {
	XMLName      xml.Name `xml:"feed"`
	Links        []Link   `xml:"link"`
	TotalResults int      `xml:"totalResults"`
	StartIndex   int      `xml:"startIndex"`
	ItemsPerPage int      `xml:"itemsPerPage"`
	Entries      []*Entry `xml:"entry"`
}

func (f *Feed) NumPages() int {
	return int(math.Ceil(float64(f.TotalResults) / float64(f.ItemsPerPage)))
}
