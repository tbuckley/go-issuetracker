package query

import (
	"errors"
	"log"
	"sync"
)

var (
	UnknownTask = errors.New("Cannot handle task")
)

type task interface {
	SetError(err error)
}

type queryResult struct {
	Feed  *Feed
	Error error
}

type queryTask struct {
	Query      *Query
	ResultChan chan *queryResult
}

func (t *queryTask) SetResponse(feed *Feed) {
	t.ResultChan <- &queryResult{Feed: feed}
}
func (t *queryTask) SetError(err error) {
	t.ResultChan <- &queryResult{Error: err}
}

type WorkGroup struct {
	taskChan chan task
}

func NewWorkGroup(numWorkers int) *WorkGroup {
	taskChan := make(chan task)

	for i := 0; i < numWorkers; i++ {
		go func(num int) {
			for {
				task, ok := <-taskChan
				if !ok {
					return
				}

				switch actualTask := task.(type) {
				case *queryTask:
					log.Printf("[%v] Fetching query: %v", num, actualTask.Query.URL())
					feed, err := actualTask.Query.fetchPage()
					if err != nil {
						actualTask.SetError(err)
					} else {
						actualTask.SetResponse(feed)
					}
				default:
					log.Printf("[%v] Cannot handle task: %#v", num, actualTask)
					task.SetError(UnknownTask)
				}

			}
		}(i)
	}

	return &WorkGroup{taskChan}
}

func (g *WorkGroup) NewQuery(project string) *Query {
	return newQuery(project, g)
}

func (g *WorkGroup) addQueryTaskWithOutput(query *Query, resultChan chan *queryResult) {
	go func() {
		g.taskChan <- &queryTask{
			Query:      query,
			ResultChan: resultChan,
		}
	}()
}

func (g *WorkGroup) addQueryTask(query *Query) chan *queryResult {
	resultChan := make(chan *queryResult)
	g.addQueryTaskWithOutput(query, resultChan)
	return resultChan
}

func (g *WorkGroup) addQueryTasks(queries []*Query) chan []*queryResult {
	multiResultChan := make(chan []*queryResult)

	go func() {
		wg := new(sync.WaitGroup)

		results := make([]*queryResult, len(queries))
		for i, query := range queries {
			wg.Add(1)
			go func(i int, query *Query) {
				results[i] = <-g.addQueryTask(query)
				wg.Done()
			}(i, query)
		}

		wg.Wait()

		multiResultChan <- results
	}()

	return multiResultChan
}
