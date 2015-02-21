package main

import (
	"encoding/xml"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

type Query struct {
	project string
	client  *http.Client
	query   []string
	params  map[string]string

	offset int
	limit  int
}

func NewQuery(project string) *Query {
	return &Query{
		project: project,
		client:  http.DefaultClient,
		query:   nil,
		params:  map[string]string{"can": "open"},

		offset: 0,
		limit:  25,
	}
}

func (q *Query) clone() *Query {
	query := make([]string, len(q.query))
	for i, value := range q.query {
		query[i] = value
	}

	params := make(map[string]string)
	for key, value := range q.params {
		params[key] = value
	}

	return &Query{
		project: q.project,
		client:  q.client,
		query:   query,
		params:  params,
		offset:  q.offset,
		limit:   q.limit,
	}
}

func (q *Query) Client(client *http.Client) *Query {
	clone := q.clone()
	clone.client = client
	return clone
}

func (q *Query) Can(can string) *Query {
	clone := q.clone()
	clone.params["can"] = can
	return clone
}

func (q *Query) Open() *Query {
	return q.Can("open")
}

func (q *Query) All() *Query {
	return q.Can("all")
}

func (q *Query) Label(label string) *Query {
	clone := q.clone()
	clone.params["label"] = label
	return clone
}

func (q *Query) Query(query string) *Query {
	clone := q.clone()
	clone.query = append(clone.query, query)
	return clone
}

func (q *Query) addDateQuery(attribute string, date time.Time) *Query {
	dateString := date.Format("2006/01/02")
	query := attribute + ":" + dateString
	return q.Query(query)
}

func (q *Query) OpenedBefore(date time.Time) *Query {
	return q.addDateQuery("opened-before", date)
}

func (q *Query) OpenedAfter(date time.Time) *Query {
	return q.addDateQuery("opened-after", date)
}

func (q *Query) OpenedInRange(start time.Time, end time.Time) *Query {
	return q.All().OpenedAfter(start).OpenedBefore(end)
}

func (q *Query) ClosedBefore(date time.Time) *Query {
	return q.addDateQuery("closed-before", date)
}

func (q *Query) ClosedAfter(date time.Time) *Query {
	return q.addDateQuery("closed-after", date)
}

func (q *Query) ClosedInRange(start time.Time, end time.Time) *Query {
	return q.All().ClosedAfter(start).ClosedBefore(end)
}

func (q *Query) Offset(offset int) *Query {
	clone := q.clone()
	clone.offset = offset
	return clone
}

func (q *Query) Limit(limit int) *Query {
	clone := q.clone()
	clone.limit = limit
	return clone
}

func (q *Query) URL() string {
	values := url.Values{}
	for key, value := range q.params {
		values.Set(key, value)
	}
	values.Set("max-results", strconv.Itoa(q.limit))
	values.Set("start-index", strconv.Itoa(q.offset+1))

	if len(q.query) > 0 {
		values.Set("q", strings.Join(q.query, " "))
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "code.google.com",
		Path:     "/feeds/issues/p/" + q.project + "/issues/full",
		RawQuery: values.Encode(),
	}
	return u.String()
}

func (q *Query) FetchPage() (*Feed, error) {
	client := http.DefaultClient
	if q.client != nil {
		client = q.client
	}

	resp, err := client.Get(q.URL())
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := new(Feed)
	err = xml.Unmarshal(data, feed)
	return feed, err
}

func (q *Query) FetchAllIssues() ([]*Entry, error) {
	entries := make([]*Entry, 0)

	workGroup := NewWorkGroup(20)

	// Fetch the first page
	result := <-workGroup.AddTask(q)
	if result.Error != nil {
		return nil, result.Error
	}

	// Get results for all additional pages
	numPages := result.Feed.NumPages()
	queries := make([]*Query, numPages-1)
	for i := 1; i < numPages; i++ {
		queries[i-1] = q.Offset(i * q.limit)
	}
	results := <-workGroup.AddTasks(queries)

	// Merge the entries together
	results = append(results, result)
	for _, result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		for _, entry := range result.Feed.Entries {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func (q *Query) FetchChangesForRange(start, end time.Time, duration time.Duration) {

}
