package query

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Query struct {
	project string
	client  *http.Client
	query   []string
	params  map[string]string

	offset int
	limit  int

	workGroup *WorkGroup
}

func newQuery(project string, workGroup *WorkGroup) *Query {
	return &Query{
		project: project,
		client:  http.DefaultClient,
		query:   nil,
		params:  map[string]string{"can": "open"},

		offset: 0,
		limit:  25,

		workGroup: workGroup,
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
		project:   q.project,
		client:    q.client,
		query:     query,
		params:    params,
		offset:    q.offset,
		limit:     q.limit,
		workGroup: q.workGroup,
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

func (q *Query) fetchPage() (*Feed, error) {
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

func (q *Query) FetchPage() (*Feed, error) {
	result := <-q.workGroup.addQueryTask(q)
	return result.Feed, result.Error
}

func (q *Query) FetchAllIssues() ([]*Entry, error) {
	entries := make([]*Entry, 0)

	// Fetch the first page
	firstPage, err := q.FetchPage()
	if err != nil {
		return nil, err
	}

	// Get results for all additional pages
	numPages := firstPage.NumPages()
	queries := make([]*Query, numPages-1)
	for i := 1; i < numPages; i++ {
		queries[i-1] = q.Offset(i * q.limit)
	}
	results := <-q.workGroup.addQueryTasks(queries)

	entries = append(entries, firstPage.Entries...)

	// Merge the entries together
	for _, result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		entries = append(entries, result.Feed.Entries...)
	}

	return entries, nil
}

func (q *Query) FetchChangesForRange(start, end time.Time, duration time.Duration) {

}
