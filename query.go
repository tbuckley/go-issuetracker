package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
type Link struct {
	Relationship string `xml:"rel,attr"`
	Type         string `xml:"type,attr"`
	Link         string `xml:"href,attr"`
}
type Feed struct {
	XMLName      xml.Name `xml:"feed"`
	Links        []Link   `xml:"link"`
	TotalResults int      `xml:"totalResults"`
	StartIndex   int      `xml:"startIndex"`
	ItemsPerPage int      `xml:"itemsPerPage"`
	Entries      []Entry  `xml:"entry"`
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
	return &Query{project: project}
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

func (q *Query) FetchAllIssues() ([]Entry, error) {
	entries := make([]Entry, 0)
	wg := new(sync.WaitGroup)

	// Create 10 goroutines to handle queries
	queryChan, resultChan := StartQueryFetchers(10)

	// Handle results as they come in
	go func() {
		for result := range resultChan {
			if result.Error != nil {
				fmt.Printf("Error: %v\n", result.Error.Error())
			} else {
				for _, entry := range result.Feed.Entries {
					entries = append(entries, entry)
				}
			}
			wg.Done()
		}
	}()

	// Fetch the first page
	feed, err := q.FetchPage()
	if err != nil {
		close(queryChan)
		return nil, err
	}
	wg.Add(1)
	resultChan <- &Result{
		Query: q,
		Feed:  feed,
	}

	// Add a query for all additional pages
	pages := math.Ceil(float64(feed.TotalResults) / float64(q.limit))
	for i := 1; i < int(pages); i++ {
		wg.Add(1)
		queryChan <- q.Offset(i * q.limit)
	}
	close(queryChan)

	wg.Wait()

	return entries, nil
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

func main() {
	q := NewQuery("chromium").Open().Limit(25)
	q = q.Label("cr-ui-settings")
	// q = q.OpenedBefore(time.Now().Add(-24 * time.Hour))

	feed, err := q.FetchPage()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	} else {
		fmt.Printf("Total issues: %v\n", feed.TotalResults)
	}

	issues, err := q.FetchAllIssues()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	} else {
		fmt.Printf("Found: %v\n", len(issues))
	}
}
